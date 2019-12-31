package httpguard

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	chaining "github.com/Laisky/go-chaining"
	utils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

type Auth struct {
	*utils.JWT
}

func NewAuth(secret string) *Auth {
	j, err := utils.NewJWT([]byte(secret))
	if err != nil {
		utils.Logger.Panic("init auth got error", zap.Error(err))
	}

	return &Auth{
		JWT: j,
	}
}

func (a *Auth) Entrypoint(c *chaining.Chain) (interface{}, error) {
	ctx := c.GetVal().(*CtxMeta)

	uinfo, err := a.loadUserInfo(ctx)
	if err != nil {
		ctx.Ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
		return nil, err
	}

	username := uinfo.username
	expires := uinfo.expires
	perms, err := a.loadPermissionsByName(username)
	if err != nil {
		ctx.Ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
		return nil, errors.Wrap(err, "token is illegal")
	}

	ok := a.validateMethodAndPath(string(ctx.Ctx.Method()), string(ctx.Ctx.Path()), perms)
	if !ok {
		ctx.Ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
		return nil, fmt.Errorf("method [%v] is illegle", string(ctx.Ctx.Method()))
	}

	ctx.Meta[Username] = username
	ctx.Meta[ExpiresAt] = expires
	return ctx, nil
}

type userInfo struct {
	username string
	expires  time.Time
}

var basicAuthPrefix = []byte("Basic ")

func (a *Auth) loadUserInfo(ctx *CtxMeta) (userinfo *userInfo, err error) {
	token := ctx.Ctx.Request.Header.Cookie("token")
	if len(token) > 0 { // auth by jwt
		utils.Logger.Debug("validate by jwt...")
		payload, err := a.validateToken(string(token))
		if err != nil {
			// return nil, errors.Wrap(err, "token is illegal")

			// compatible for legacy token
			utils.Logger.Warn("compatible legacy token", zap.Error(err), zap.String("payload", fmt.Sprint(payload)))
			if !strings.Contains(err.Error(), "method") {
				return nil, errors.Wrap(err, "token is illegal")
			}
		}

		// compatible for legacy token
		var (
			usernamei, expiresi interface{}
			username            string
			expires             time.Time
			ok                  bool
		)
		if usernamei, ok = payload[a.GetUserIDKey()]; !ok {
			if usernamei, ok = payload["username"]; !ok {
				return nil, errors.Errorf("unknown user key: %+v", payload)
			}
		}
		username = usernamei.(string)
		if expiresi, ok = payload[a.GetExpiresKey()]; !ok {
			if expiresi, ok = payload["expires_at"]; !ok {
				return nil, errors.Errorf("unknown user key: %+v", payload)
			}
			if expires, err = time.Parse(time.RFC3339, expiresi.(string)); err != nil {
				return nil, errors.Wrapf(err, "unkndown format of expires: %v", expiresi.(string))
			}
		} else {
			expires = expiresi.(time.Time)
		}

		return &userInfo{
			username: username,
			expires:  expires,
		}, nil
	}

	// auth by basicaith
	auth := ctx.Ctx.Request.Header.Peek("Authorization")
	utils.Logger.Debug("got authorization auth", zap.String("auth'", string(auth)))
	if len(auth) > 0 {
		utils.Logger.Debug("validate by basic auth...")
		payload, err := base64.StdEncoding.DecodeString(string(auth[len(basicAuthPrefix):]))
		if err != nil {
			return nil, errors.Wrap(err, "decode auth got error")
		}

		pair := bytes.SplitN(payload, []byte(":"), 2)
		if len(pair) == 2 {
			username := string(pair[0])
			if pw, ok := a.loadPasswdByName(username); ok {
				if string(pair[1]) == pw {
					return &userInfo{
						username: username,
					}, nil
				}
			}
		}
		return nil, fmt.Errorf("auth length expect 2, but got %v", len(pair))
	}

	return nil, errors.New("auth failed")
}

func (a *Auth) loadPasswdByName(username string) (passwd string, ok bool) {
	utils.Logger.Debug("loadPasswdByName", zap.String("username", username))
	var umi map[interface{}]interface{}
	for _, ui := range utils.Settings.Get("users").([]interface{}) {
		umi = ui.(map[interface{}]interface{})
		if umi["username"].(string) == username {
			if pw, ok := umi["password"]; ok {
				return pw.(string), true
			}
		}
	}

	utils.Logger.Debug("can not load password", zap.String("username", username))
	return "", false
}

func (a *Auth) loadPermissionsByName(username string) (perm map[string][]string, err error) {
	var (
		umi, pi                 map[interface{}]interface{}
		methodI, pathesI, pathI interface{}
		pathes                  []string
	)
	perm = map[string][]string{}
	for _, ui := range utils.Settings.Get("users").([]interface{}) {
		umi = ui.(map[interface{}]interface{})
		if umi["username"].(string) == username {
			pi = umi["permissions"].(map[interface{}]interface{})
			for methodI, pathesI = range pi {
				pathes = []string{}
				for _, pathI = range pathesI.([]interface{}) {
					pathes = append(pathes, pathI.(string))
				}
				perm[methodI.(string)] = pathes
			}

			return
		}
	}

	return nil, fmt.Errorf("username [%v] not exists in config", username)
}

// validateToken check username is legal
func (a *Auth) validateToken(token string) (payload map[string]interface{}, err error) {
	payload, err = a.Validate(token)
	if err != nil {
		return payload, errors.Wrap(err, "token not validate")
	}

	return
}

// validateMethodAndPath chech path & method is legal
func (a *Auth) validateMethodAndPath(method, path string, permissions map[string][]string) (ok bool) {
	utils.Logger.Debug("validateMethodAndPath", zap.String("method", method), zap.String("path", path))
	for pm, pps := range permissions {
		if strings.EqualFold(pm, method) {
			utils.Logger.Debug("check method", zap.String("method", method))
			for _, pp := range pps {
				if strings.Index(path, pp) == 0 {
					return true
				}
			}
		}
	}

	return false
}

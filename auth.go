package httpguard

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	chaining "github.com/Laisky/go-chaining"
	gutils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

type Auth struct {
	plugins []AuthPlugin
}

type authToken struct {
	jwt.RegisteredClaims
}

func NewAuth(plugins ...AuthPlugin) *Auth {
	return &Auth{plugins: plugins}
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
	for _, plugin := range a.plugins {
		userinfo, err = plugin.loadUserInfo(ctx)
		if err == nil {
			return userinfo, nil
		}
	}

	return nil, errors.Wrap(err, "authentication failed")
}

func (a *Auth) loadPasswdByName(username string) (passwd string, ok bool) {
	Logger.Debug("loadPasswdByName", zap.String("username", username))
	var umi map[interface{}]interface{}
	for _, ui := range gutils.Settings.Get("users").([]interface{}) {
		umi = ui.(map[interface{}]interface{})
		if umi["username"].(string) == username {
			if pw, ok := umi["password"]; ok {
				return pw.(string), true
			}
		}
	}

	Logger.Debug("can not load password", zap.String("username", username))
	return "", false
}

func (a *Auth) loadPermissionsByName(username string) (perm configUserPerm, err error) {
	user, ok := Config.UsersMap[username]
	if !ok {
		return perm, errors.Errorf("user `%s` not exist", username)
	}

	return user.Perms, nil
}

// validateMethodAndPath chech path & method is legal
func (a *Auth) validateMethodAndPath(method, path string, perm configUserPerm) (ok bool) {
	Logger.Debug("validateMethodAndPath", zap.String("method", method), zap.String("path", path))
	var allowedUrls []string
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		allowedUrls = perm.Read
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		allowedUrls = perm.Write
	default:
		Logger.Panic("unknown http method", zap.String("method", method))
	}

	for _, url := range allowedUrls {
		if strings.HasPrefix(path, url) {
			return true
		}
	}

	return false
}

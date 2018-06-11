package httpguard

import (
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"

	chaining "github.com/Laisky/go-chaining"
	utils "github.com/Laisky/go-utils"
	"github.com/pkg/errors"
)

type Auth struct {
	utils.JWT
}

func NewAuth(secret string) *Auth {
	a := &Auth{}
	a.Setup(secret)
	return a
}

func (a *Auth) Entrypoint(c *chaining.Chain) (interface{}, error) {
	ctx := c.GetVal().(*CtxMeta)
	payload, err := a.validateToken(string(ctx.Ctx.Request.Header.Cookie("token")))
	if err != nil {
		ctx.Ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
		return nil, errors.Wrap(err, "token is illegal")
	}
	username := payload[a.TKUsername].(string)
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
	ctx.Meta[ExpiresAt] = payload[a.TKExpiresAt]
	return ctx, nil
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
		return nil, errors.Wrap(err, "token not validate")
	}

	return
}

// validateMethodAndPath chech path & method is legal
func (a *Auth) validateMethodAndPath(method, path string, permissions map[string][]string) (ok bool) {
	utils.Logger.Debugf("validateMethodAndPath for method %v, path %v, permissions %+v", method, path, permissions)
	for pm, pps := range permissions {
		if strings.ToLower(pm) == strings.ToLower(method) {
			utils.Logger.Debugf("check method %v", method)
			for _, pp := range pps {
				utils.Logger.Debugf("check path %v:%v", pp, path)
				if strings.Index(path, pp) == 0 {
					return true
				}
			}
		}
	}

	return false
}

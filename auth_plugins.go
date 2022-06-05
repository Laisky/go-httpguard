package httpguard

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	gutils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/pkg/errors"
)

type AuthPlugin interface {
	loadUserInfo(ctx *CtxMeta) (userinfo *userInfo, err error)
}

type JwtAuthPlugin struct {
	jwtEngine *gutils.JWT
	secret    string
}

func NewJwtAuthPlugin(secret string) *JwtAuthPlugin {
	jwtEngine, err := gutils.NewJWT(gutils.WithJWTSecretByte([]byte(secret)))
	if err != nil {
		Logger.Panic("init auth got error", zap.Error(err))
	}

	return &JwtAuthPlugin{
		secret:    secret,
		jwtEngine: jwtEngine,
	}
}

func (p *JwtAuthPlugin) loadUserInfo(ctx *CtxMeta) (userinfo *userInfo, err error) {
	token := string(ctx.Ctx.Request.Header.Peek("Authorization"))
	if token == "" {
		return nil, errors.Errorf("no token found")
	}

	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return nil, errors.Wrap(err, "no token found")
	}

	claims := new(authToken)
	if err = p.jwtEngine.ParseClaimsByHS256(token, claims); err != nil {
		return nil, errors.Wrap(err, "token is illegal")
	}

	if err = claims.Valid(); err != nil {
		return nil, errors.Wrap(err, "token invalid")
	}

	if !Config.UsersMap[claims.Subject].JWT.Enable {
		return nil, errors.Errorf("user %v not enable jwt", claims.Subject)
	}

	userinfo = &userInfo{
		username: claims.Subject,
		expires:  claims.ExpiresAt.Time,
	}

	return userinfo, nil
}

type BasicAuthPlugin struct {
}

func NewBasicAuthPlugin() *BasicAuthPlugin {
	return &BasicAuthPlugin{}
}

func (p *BasicAuthPlugin) loadUserInfo(ctx *CtxMeta) (userinfo *userInfo, err error) {
	auth := ctx.Ctx.Request.Header.Peek("Authorization")
	Logger.Debug("got authorization auth", zap.String("auth'", string(auth)))
	if len(auth) <= 0 {
		return nil, errors.New("no auth found")
	}

	Logger.Debug("validate by basic auth...")
	payload, err := base64.StdEncoding.DecodeString(string(auth[len(basicAuthPrefix):]))
	if err != nil {
		return nil, errors.Wrap(err, "decode auth got error")
	}

	pair := bytes.SplitN(payload, []byte(":"), 2)
	if len(pair) != 2 {
		return nil, fmt.Errorf("auth length expect 2, but got %v", len(pair))
	}

	username := string(pair[0])
	if userCfg, ok := Config.UsersMap[username]; !ok {
		return nil, errors.Errorf("user %v not found", username)
	} else {
		if !Config.UsersMap[username].BasicAuth.Enable {
			return nil, errors.Errorf("user %v not enable basicauth", username)
		}

		if string(pair[1]) == userCfg.BasicAuth.Password {
			return &userInfo{
				username: username,
			}, nil
		}

		return nil, errors.Errorf("password not match")
	}

	return nil, errors.New("auth failed")
}

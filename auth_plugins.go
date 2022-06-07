package httpguard

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Laisky/go-aws-auth"
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
	userCfg, ok := Config.UsersMap[username]
	if !ok {
		return nil, errors.Errorf("user %v not found", username)
	}

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

type AwsAuthPlugin struct {
}

func NewAwsAuthPlugin() *AwsAuthPlugin {
	return &AwsAuthPlugin{}
}

var regexpAwsUsername = regexp.MustCompile(`Credential=([^/]+)/`)
var regexpSignature = regexp.MustCompile(`Signature=(\w+)`)

func (p *AwsAuthPlugin) loadUserInfo(ctx *CtxMeta) (userinfo *userInfo, err error) {
	rawAuth := ctx.Ctx.Request.Header.Peek("Authorization")

	matched := regexpAwsUsername.FindAllStringSubmatch(string(rawAuth), -1)
	if len(matched) == 0 {
		return nil, errors.New("no aws auth found")
	}
	if len(matched[0]) != 2 {
		return nil, errors.New("aws auth format error")
	}

	expectReq := &http.Request{
		Method: string(ctx.Ctx.Request.Header.Method()),
		Host:   string(ctx.Ctx.Request.Host()),
		Proto:  string(ctx.Ctx.Request.Header.Protocol()),
		Header: http.Header{},
	}

	requrl, err := url.ParseRequestURI(string(ctx.Ctx.Request.RequestURI()))
	if err != nil {
		return nil, errors.Wrapf(err, "parse request uri `%s`", string(ctx.Ctx.Request.RequestURI()))
	}
	expectReq.URL = requrl

	rawHeaders := map[string]string{}
	ctx.Ctx.Request.Header.VisitAll(func(key, value []byte) {
		rawHeaders[string(key)] = string(value)
		expectReq.Header.Add(string(key), string(value))
	})

	username := matched[0][1]
	expectReq = awsauth.Sign4(expectReq, awsauth.Credentials{
		AccessKeyID:     username,
		SecretAccessKey: Config.UsersMap[username].S3.AppSecret,
	})

	var (
		rawSig, expectSig string
	)
	if matched := regexpSignature.FindAllStringSubmatch(
		string(rawAuth), -1,
	); len(matched) > 0 {
		rawSig = matched[0][1]
	}
	if matched := regexpSignature.FindAllStringSubmatch(
		expectReq.Header.Get("Authorization"), -1,
	); len(matched) > 0 {
		expectSig = matched[0][1]
	}

	if rawSig != expectSig {
		Logger.Debug("expect auth",
			zap.String("expect", expectReq.Header.Get("Authorization")),
			zap.String("actual", string(rawAuth)))
		return nil, errors.New("aws auth invalid")
	}

	return &userInfo{
		username: username,
	}, nil

}

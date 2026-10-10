package httpguard

import (
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/Laisky/go-chaining"
	"github.com/golang-jwt/jwt/v4"
	"github.com/valyala/fasthttp"
)

// TestAuthenticationDependencyContracts exercises real middleware with synthetic users and tokens.
func TestAuthenticationDependencyContracts(t *testing.T) {
	previous := Config
	t.Cleanup(func() { Config = previous })
	secret := "synthetic-test-signing-key"
	makeToken := func(method jwt.SigningMethod, key string, expiry time.Time) string {
		t.Helper()
		token, err := jwt.NewWithClaims(method, authToken{RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "fixture",
			ExpiresAt: jwt.NewNumericDate(expiry),
		}}).SignedString([]byte(key))
		if err != nil {
			t.Fatal(err)
		}
		return "Bearer " + token
	}
	now := time.Now().UTC()
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte("fixture:synthetic-password"))
	wrongBasic := "Basic " + base64.StdEncoding.EncodeToString([]byte("fixture:wrong-password"))
	validJWT := makeToken(jwt.SigningMethodHS256, secret, now.Add(time.Hour))
	invalidJWT := makeToken(jwt.SigningMethodHS256, "different-synthetic-key", now.Add(time.Hour))
	expiredJWT := makeToken(jwt.SigningMethodHS256, secret, now.Add(-time.Hour))
	wrongAlgorithm := makeToken(jwt.SigningMethodHS384, secret, now.Add(time.Hour))
	cases := []struct {
		name, method, path, authorization string
		basicDisabled, jwtDisabled        bool
		status                            int
		username                          string
	}{
		{name: "basic permitted read", method: http.MethodGet, path: "/read/item", authorization: basic, status: 200, username: "fixture"},
		{name: "basic wrong password", method: http.MethodGet, path: "/read/item", authorization: wrongBasic, status: 403},
		{name: "basic forbidden write", method: http.MethodPost, path: "/read/item", authorization: basic, status: 403},
		{name: "basic disabled", method: http.MethodGet, path: "/read/item", authorization: basic, basicDisabled: true, status: 403},
		{name: "missing credential", method: http.MethodGet, path: "/read/item", status: 403},
		{name: "configured public route", method: http.MethodGet, path: "/public/item", status: 200, username: "guest"},
		{name: "jwt permitted read", method: http.MethodGet, path: "/read/item", authorization: validJWT, status: 200, username: "fixture"},
		{name: "jwt wrong signing key", method: http.MethodGet, path: "/read/item", authorization: invalidJWT, status: 403},
		{name: "jwt expired", method: http.MethodGet, path: "/read/item", authorization: expiredJWT, status: 403},
		{name: "existing HMAC family compatibility", method: http.MethodGet, path: "/read/item", authorization: wrongAlgorithm, status: 200, username: "fixture"},
		{name: "jwt disabled", method: http.MethodGet, path: "/read/item", authorization: validJWT, jwtDisabled: true, status: 403},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			Config = (&ConfigType{
				Bypass: configUserPerm{Read: []string{"/public/"}},
				Users: []configUser{{
					Username:  "fixture",
					BasicAuth: configUserBasicAuth{Enable: !tc.basicDisabled, Password: "synthetic-password"},
					JWT:       configUserJWT{Enable: !tc.jwtDisabled},
					Perms:     configUserPerm{Read: []string{"/read/"}, Write: []string{"/write/"}},
				}},
			}).Init()
			authorized := false
			controller := NewController(NewAuth(NewJwtAuthPlugin(secret), NewBasicAuthPlugin()), contractMiddleware{
				handle: func(chain *chaining.Chain) (interface{}, error) {
					metadata := chain.GetVal().(*CtxMeta)
					authorized = true
					metadata.Ctx.Response.SetBodyString(metadata.Meta[Username].(string))
					return metadata, nil
				},
			})
			request := &fasthttp.RequestCtx{}
			request.Request.Header.SetMethod(tc.method)
			request.Request.SetRequestURI(tc.path)
			request.Request.Header.Set("Authorization", tc.authorization)
			getRequestHandler(controller)(request)
			if got := request.Response.StatusCode(); got != tc.status {
				t.Fatalf("status = %d, want %d", got, tc.status)
			}
			if authorized != (tc.status == 200) {
				t.Fatalf("authenticated continuation = %v for status %d", authorized, tc.status)
			}
			if tc.status == 200 && string(request.Response.Body()) != tc.username {
				t.Fatalf("response identity = %q, want %q", request.Response.Body(), tc.username)
			}
		})
	}
}

type contractMiddleware struct {
	handle func(*chaining.Chain) (interface{}, error)
}

// Entrypoint passes the authenticated chain to the synthetic response handler.
func (m contractMiddleware) Entrypoint(c *chaining.Chain) (interface{}, error) { return m.handle(c) }

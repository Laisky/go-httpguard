package httpguard

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/Laisky/go-chaining"
	utils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

var (
	httpClient = &http.Client{ // default http client
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 20,
		},
		Timeout: time.Duration(30) * time.Second,
	}

	bbPool = &sync.Pool{
		New: func() interface{} {
			return []byte{}
		},
	}
)

type Backend struct{}

func NewBackend() *Backend {
	return &Backend{}
}

func (b *Backend) Entrypoint(c *chaining.Chain) (ret interface{}, err error) {
	ctx := c.GetVal().(*CtxMeta)
	url := utils.Settings.GetString("backend") + string(ctx.Ctx.RequestURI())
	utils.Logger.Debug("request to backend for url", zap.String("url", url))
	if utils.Settings.GetBool("dry") {
		return ctx, nil
	}

	err = b.RequestBackend(ctx.Ctx, url)
	if err != nil {
		ctx.Ctx.Response.SetStatusCode(fasthttp.StatusBadGateway)
		fmt.Fprintln(ctx.Ctx, "try to request backend got error")
		return nil, errors.Wrap(err, "try to request backend got error")
	}

	return ctx, nil
}

// RequestBackend request backend by internal http client
func (b *Backend) RequestBackend(ctx *fasthttp.RequestCtx, url string) (err error) {
	req, err := http.NewRequest(string(ctx.Method()), url, bytes.NewBuffer(ctx.PostBody()))
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	bb := bbPool.Get().([]byte)
	bb, err = ioutil.ReadAll(resp.Body)
	ctx.SetBody(bb)
	bbPool.Put(bb)
	ctx.SetStatusCode(resp.StatusCode)

	return nil
}

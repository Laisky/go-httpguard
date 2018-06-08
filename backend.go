package httpguard

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Laisky/go-chaining"
	"github.com/pkg/errors"

	utils "github.com/Laisky/go-utils"
	"github.com/valyala/fasthttp"
)

var (
	httpClient = &http.Client{ // default http client
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 20,
		},
		Timeout: time.Duration(30) * time.Second,
	}

	Backend = &BackendType{}
)

type BackendType struct{}

func (b *BackendType) Entrypoint(c *chaining.Chain) (ret interface{}, err error) {
	ctx := c.GetVal().(*CtxMeta)
	url := utils.Settings.GetString("backend") + string(ctx.Ctx.RequestURI())
	err = b.RequestBackend(ctx.Ctx, url)
	if err != nil {
		ctx.Ctx.Response.SetStatusCode(fasthttp.StatusBadGateway)
		fmt.Fprintln(ctx.Ctx, "try to request backend got error")
		return nil, errors.Wrap(err, "try to request backend got error")
	}

	return ctx, nil
}

// RequestBackend request backend by internal http client
func (b *BackendType) RequestBackend(ctx *fasthttp.RequestCtx, url string) (err error) {
	req, err := http.NewRequest(string(ctx.Method()), url, bytes.NewBuffer(ctx.PostBody()))
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	bb, err := ioutil.ReadAll(resp.Body)
	ctx.SetBody(bb)
	ctx.SetStatusCode(resp.StatusCode)

	return nil
}

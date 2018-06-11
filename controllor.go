package httpguard

import (
	"fmt"

	"github.com/Laisky/go-chaining"

	"github.com/Laisky/go-utils"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

var ()

type CtxMeta struct {
	Ctx  *fasthttp.RequestCtx
	Meta map[string]interface{}
}

type Middleware interface {
	Entrypoint(c *chaining.Chain) (interface{}, error)
}

type Controllor struct {
	middlewares []Middleware
}

func NewController(middlewares ...Middleware) *Controllor {
	return &Controllor{
		middlewares: middlewares,
	}
}

func (co *Controllor) MiddlewareChain(c *chaining.Chain) *chaining.Chain {
	for _, m := range co.middlewares {
		c = c.Next(m.Entrypoint)
	}

	return c
}

func (c *Controllor) Run() (err error) {
	addr := utils.Settings.GetString("addr")
	utils.Logger.Infof("listen to %v", addr)
	if err := fasthttp.ListenAndServe(addr, fasthttp.CompressHandler(getRequestHandler(c))); err != nil {
		return errors.Wrap(err, "try to listen server error")
	}

	return nil
}

func getRequestHandler(co *Controllor) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if err := recover(); err != nil {
				utils.Logger.Errorf("requests got error %+v", err)
				fmt.Fprintf(ctx, "got unhandle error %v", err)
			}
		}()

		c := co.MiddlewareChain(chaining.New(&CtxMeta{
			Ctx:  ctx,
			Meta: map[string]interface{}{},
		}, nil))

		if c.GetError() != nil {
			utils.Logger.Infof("controllor got error: %v", c.GetError())
		}
	}
}

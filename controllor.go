package httpguard

import (
	"context"

	"github.com/Laisky/go-chaining"
	"github.com/Laisky/zap"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

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

func (c *Controllor) MiddlewareChain(chain *chaining.Chain) *chaining.Chain {
	for _, m := range c.middlewares {
		chain = chain.Next(m.Entrypoint)
	}

	return chain
}

func (c *Controllor) Run(ctx context.Context) (err error) {
	listen := Config.Listen
	Logger.Info("listen addr", zap.String("addr", listen))
	if err := fasthttp.ListenAndServe(listen, fasthttp.CompressHandler(getRequestHandler(c))); err != nil {
		return errors.Wrap(err, "try to listen server error")
	}

	return nil
}

func getRequestHandler(co *Controllor) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if err := recover(); err != nil {
				Logger.Error("requests got error", zap.Error(err.(error)))
			}
		}()

		c := co.MiddlewareChain(chaining.New(&CtxMeta{
			Ctx:  ctx,
			Meta: map[string]interface{}{},
		}, nil))

		if c.GetError() != nil {
			Logger.Info("controllor got error", zap.Error(c.GetError()))
		}
	}
}

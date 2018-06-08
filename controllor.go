package httpguard

import (
	"fmt"

	"github.com/Laisky/go-chaining"

	"github.com/Laisky/go-utils"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

var (
	Controllor = &ControllorType{}
)

type CtxMeta struct {
	Ctx  *fasthttp.RequestCtx
	Meta map[string]interface{}
}

type ControllorType struct{}

func (c *ControllorType) Setup() {
	Auth.Setup(utils.Settings.GetString("secret"))
}

func (c *ControllorType) Run() (err error) {
	addr := utils.Settings.GetString("addr")
	utils.Logger.Infof("listen to %v", addr)
	if err := fasthttp.ListenAndServe(addr, fasthttp.CompressHandler(requestHandler)); err != nil {
		return errors.Wrap(err, "try to listen server error")
	}

	return nil
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	defer func() {
		if err := recover(); err != nil {
			utils.Logger.Errorf("requests got error %+v", err)
			fmt.Fprintf(ctx, "got unhandle error %v", err)
		}
	}()

	c := chaining.New(&CtxMeta{
		Ctx:  ctx,
		Meta: map[string]interface{}{},
	}, nil).
		Next(Auth.Entrypoint).
		Next(Audit.Entrypoint).
		Next(Backend.Entrypoint)

	if c.GetError() != nil {
		utils.Logger.Infof("controllor got error: %v", c.GetError())
	}
}

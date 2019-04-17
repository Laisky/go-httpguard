package httpguard

import (
	"github.com/Laisky/go-chaining"
	utils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
)

type Audit struct{}

func NewAudit() *Audit {
	return &Audit{}
}

func (a *Audit) Entrypoint(c *chaining.Chain) (interface{}, error) {
	ctx := c.GetVal().(*CtxMeta)
	if utils.Settings.GetBool("dry") {
		return ctx, nil
	}

	go a.PushAuditRecord(map[string]interface{}{
		"username":   ctx.Meta[Username],
		"expires_at": ctx.Meta[ExpiresAt],
		"path":       string(ctx.Ctx.Path()),
		"method":     string(ctx.Ctx.Method()),
		"body":       string(ctx.Ctx.PostBody()),
		"@timestamp": utils.UTCNow(),
	})

	return ctx, nil
}

func (a *Audit) PushAuditRecord(data map[string]interface{}) {
	reqData := &utils.RequestData{
		Data: data,
	}
	var resp interface{}
	err := utils.RequestJSONWithClient(httpClient, "post", utils.Settings.GetString("audit"), reqData, &resp)
	if err != nil {
		utils.Logger.Error("try to push audit log got error", zap.Error(err))
	}
}

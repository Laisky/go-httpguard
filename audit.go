package httpguard

import (
	"github.com/Laisky/go-chaining"
	gutils "github.com/Laisky/go-utils"
	"github.com/Laisky/zap"
)

type Audit struct{}

func NewAudit() *Audit {
	return &Audit{}
}

func (a *Audit) Entrypoint(c *chaining.Chain) (interface{}, error) {
	ctx := c.GetVal().(*CtxMeta)

	// go a.PushAuditRecord(map[string]interface{}{
	// 	"username":   ctx.Meta[Username].(string),
	// 	"expires_at": ctx.Meta[ExpiresAt].(time.Time).Format(time.RFC3339Nano),
	// 	"path":       string(ctx.Ctx.Path()),
	// 	"method":     string(ctx.Ctx.Method()),
	// 	"body":       string(ctx.Ctx.PostBody()),
	// 	"@timestamp": gutils.UTCNow(),
	// })

	if gutils.InArray([]string{"POST", "PUT", "DELETE", "PATCH"}, string(ctx.Ctx.Method())) {
		Logger.Info("audit",
			zap.Any("user", ctx.Meta[Username]),
			zap.ByteString("path", ctx.Ctx.Path()),
			zap.ByteString("method", ctx.Ctx.Method()),
		)
	}

	return ctx, nil
}

func (a *Audit) PushAuditRecord(data map[string]interface{}) {
	reqData := &gutils.RequestData{
		Data: data,
	}
	var resp interface{}
	err := gutils.RequestJSONWithClient(httpClient, "post", gutils.Settings.GetString("audit"), reqData, &resp)
	if err != nil {
		Logger.Error("try to push audit log got error", zap.Error(err))
	}
}

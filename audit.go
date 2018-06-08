package httpguard

import (
	"github.com/Laisky/go-chaining"
	utils "github.com/Laisky/go-utils"
)

var (
	Audit = AuditType{}
)

type AuditType struct{}

func (a *AuditType) Entrypoint(c *chaining.Chain) (interface{}, error) {
	ctx := c.GetVal().(*CtxMeta)
	go a.PushAuditRecord(map[string]interface{}{
		"username":   ctx.Meta[Auth.TKUsername],
		"expires_at": ctx.Meta[Auth.TKExpiresAt],
		"path":       string(ctx.Ctx.Path()),
		"method":     string(ctx.Ctx.Method()),
		"body":       string(ctx.Ctx.PostBody()),
		"@timestamp": utils.UTCNow(),
	})

	return ctx, nil
}

func (a *AuditType) PushAuditRecord(data map[string]interface{}) {
	utils.Logger.Infof("user %v %v %v", data["username"], data["method"], data["path"])
	reqData := &utils.RequestData{
		Data: data,
	}
	var resp interface{}
	err := utils.RequestJSONWithClient(httpClient, "post", utils.Settings.GetString("audit"), reqData, &resp)
	if err != nil {
		utils.Logger.Errorf("try to push audit log got error %+v", err)
	}
}

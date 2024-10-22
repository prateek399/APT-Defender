package controller

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func OverrideVerdict(ctx *gin.Context) {
	var overrideVerdictRequest model.OverrideVerdictRequest
	var resp model.APIResponse

	if err := ctx.ShouldBindJSON(&overrideVerdictRequest); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	session, _ := auth.Store.Get(ctx.Request, "sessionid")
	defer service.CreateAuditLogs(&resp, fmt.Sprintf("Changed the verdict to %s", overrideVerdictRequest.Verdict), "OVERRIDE VERDICT", session.Values["admin_name"].(string))

	resp = service.OverrideVerdict(overrideVerdictRequest, ctx)
	ctx.JSON(resp.StatusCode, resp)
}

func GetOverriddenVerdictLogs(ctx *gin.Context) {
	resp := service.GetOverriddenVerdictLogs(ctx)
	ctx.JSON(resp.StatusCode, resp)
}

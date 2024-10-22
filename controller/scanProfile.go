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

func ScanProfile(ctx *gin.Context) {
	var scanProfileRequest model.ScanProfile
	var resp model.APIResponse

	if err := ctx.ShouldBindJSON(&scanProfileRequest); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	session, _ := auth.Store.Get(ctx.Request, "sessionid")
	defer service.CreateAuditLogs(&resp, fmt.Sprintf("%s changed the scan profile permissions", session.Values["admin_name"].(string)), "SCAN PROFILE", session.Values["admin_name"].(string))

	resp = service.ScanProfile(scanProfileRequest, session.Values["user_id"].(string))
	ctx.JSON(resp.StatusCode, resp)
}

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

func ChangePassword(ctx *gin.Context) {
	var resp model.APIResponse
	var adminPassChange model.ChangePass

	if err := ctx.ShouldBindJSON(&adminPassChange); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	session, err := auth.Store.Get(ctx.Request, "sessionid")
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SESSION_INVALID, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	defer service.CreateAuditLogs(&resp, fmt.Sprintf("Password changed successfully for user %s", session.Values["admin_name"].(string)), "CHANGED PASSWORD", session.Values["admin_name"].(string))

	if session.Values["user_id"] == nil {
		resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_UNAUTHORIZED_USER, extras.ErrUnauthorizedUser)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp = service.ChangePassword(adminPassChange, session.Values["user_id"].(string))
	ctx.JSON(resp.StatusCode, resp)
}

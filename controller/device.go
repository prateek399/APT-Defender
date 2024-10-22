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

func CreateDevice(ctx *gin.Context) {
	var resp model.APIResponse
	var deviceRequest model.Device
	if err := ctx.ShouldBindJSON(&deviceRequest); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	session, _ := auth.Store.Get(ctx.Request, "sessionid")
	defer service.CreateAuditLogs(&resp, fmt.Sprintf("Added a new device %s", deviceRequest.DeviceName), "DEVICE", session.Values["admin_name"].(string))

	resp = service.CreateDevice(deviceRequest)
	ctx.JSON(resp.StatusCode, resp)
}

func DeleteDevice(ctx *gin.Context) {
	var resp model.APIResponse

	session, _ := auth.Store.Get(ctx.Request, "sessionid")
	defer service.CreateAuditLogs(&resp, "Successfully deleted a device", "DEVICE", session.Values["admin_name"].(string))

	resp = service.DeleteDevice(ctx.Query("key"))
	ctx.JSON(resp.StatusCode, resp)
}

func UpdateDevice(ctx *gin.Context) {
	var resp model.APIResponse

	session, _ := auth.Store.Get(ctx.Request, "sessionid")
	defer service.CreateAuditLogs(&resp, "Successfully updated a device", "DEVICE", session.Values["admin_name"].(string))

	var deviceRequest model.Device
	if err := ctx.ShouldBindJSON(&deviceRequest); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp = service.UpdateDevice(deviceRequest)
	ctx.JSON(resp.StatusCode, resp)
}

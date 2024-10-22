package controller

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ExtendLicense(ctx *gin.Context) {
	var licesneKey string
	if err := ctx.ShouldBindJSON(&licesneKey); err != nil {
		resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp := service.CreateLicenseKey(licesneKey, "")
	ctx.JSON(resp.StatusCode, resp)
}

func ApplyDeviceConfigFile(ctx *gin.Context) {
	var req model.DeviceConfigFile
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp := service.ApplyDeviceConfigFile(req)
	ctx.JSON(resp.StatusCode, resp)
}

func DeviceConfigUiForm(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

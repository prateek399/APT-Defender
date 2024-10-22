package controller

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateLicenseKey(ctx *gin.Context) {
	var licenseKeyRequest model.KeysTable
	var resp model.APIResponse

	if err := ctx.ShouldBindJSON(&licenseKeyRequest); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp = service.CreateLicenseKey(licenseKeyRequest.ApplianceKey, "")
	ctx.JSON(resp.StatusCode, resp)
}

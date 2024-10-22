package controller

import (
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckLicenseKeyAvailability(ctx *gin.Context) {
	var resp model.APIResponse
	// TODO: implement this method to check license
	_, err := dao.FetchLicenseKeyProfile()
	if err == extras.ErrNoRecordForLicenseKey {
		resp = model.NewErrorResponse(http.StatusOK, extras.ERR_RECORD_NOT_FOUND, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "License key is present")
	ctx.JSON(resp.StatusCode, resp)
}

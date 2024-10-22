package controller

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Troubleshoot(ctx *gin.Context) {

	var req model.TroubleshootRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp := service.Troubleshoot(req)
	ctx.JSON(resp.StatusCode, resp)
}

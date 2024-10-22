package controller

import (
	"anti-apt-backend/service"

	"github.com/gin-gonic/gin"
)

func Dashboard(ctx *gin.Context) {
	resp := service.Dashboard(ctx.Query("action"))
	ctx.JSON(resp.StatusCode, resp)
}

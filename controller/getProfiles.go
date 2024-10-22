package controller

import (
	"anti-apt-backend/service"

	"github.com/gin-gonic/gin"
)

func GetAllProfiles(ctx *gin.Context) {
	resp := service.GetAllProfiles(ctx.Query("model"), ctx.Query("action"))
	ctx.JSON(resp.StatusCode, resp)
}

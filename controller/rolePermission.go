package controller

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateRolePermission(ctx *gin.Context) {
	var resp model.APIResponse
	var role model.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp = service.CreateRolePermission(role, false)
	ctx.JSON(resp.StatusCode, resp)
}

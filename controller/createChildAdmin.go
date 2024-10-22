package controller

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateChildAdmin(ctx *gin.Context) {
	var signupRequest model.SignupRequest

	if err := ctx.ShouldBindJSON(&signupRequest); err != nil {
		resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp := service.CreateChildAdmin(signupRequest)
	ctx.JSON(resp.StatusCode, resp)
}

func UpdateAdminPersonalDetails(ctx *gin.Context) {
	var signupRequest model.SignupRequest

	if err := ctx.ShouldBindJSON(&signupRequest); err != nil {
		resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	session, _ := auth.Store.Get(ctx.Request, "sessionid")

	resp := service.UpdateAdminPersonalDetails(signupRequest, session)
	if resp.StatusCode == http.StatusOK {
		if err := session.Save(ctx.Request, ctx.Writer); err != nil {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SESSION_NOT_SAVED, err)
			ctx.JSON(resp.StatusCode, resp)
			return
		}
	}

	ctx.JSON(resp.StatusCode, resp)
}

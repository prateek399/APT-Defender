package controller

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Signup(ctx *gin.Context) {
	var signupRequest model.SignupRequest
	var resp model.APIResponse

	if err := ctx.ShouldBindJSON(&signupRequest); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	session, err := auth.Store.New(ctx.Request, "sessionid")
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SESSION_INVALID, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp = service.Signup(signupRequest, session)
	ctx.JSON(resp.StatusCode, resp)
}

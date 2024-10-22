package controller

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	var loginRequest model.LoginRequest
	var resp model.APIResponse
	// ctx.Header("Origin", ctx.ClientIP())
	// ctx.Header("Access-Control-Allow-Origin", ctx.Request.URL.Scheme+"://"+ctx.ClientIP())
	// ctx.Header("Access-Control-Allow-Origin", ctx.ClientIP())

	if err := ctx.ShouldBindJSON(&loginRequest); err != nil {
		if ctx.GetHeader("Authorization") != "" {
			session, err := auth.Store.Get(ctx.Request, "sessionid")
			if err != nil || session == nil {
				// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
				resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SESSION_INVALID, err)
				ctx.JSON(resp.StatusCode, resp)
				return
			}

			bearerToken := strings.Split(ctx.GetHeader("Authorization"), " ")
			resp = service.Login(loginRequest, session, bearerToken)

			if err := session.Save(ctx.Request, ctx.Writer); err != nil {
				// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
				resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SESSION_NOT_SAVED, err)
				ctx.JSON(resp.StatusCode, resp)
				return
			}

			ctx.JSON(resp.StatusCode, resp)
			return
		}

		// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:error from client's end"))
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

	resp = service.Login(loginRequest, session, nil)
	if err := session.Save(ctx.Request, ctx.Writer); err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SESSION_NOT_SAVED, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}
	ctx.JSON(resp.StatusCode, resp)
}

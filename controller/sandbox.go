package controller

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	queues "anti-apt-backend/service/queue"
	"net/http"

	"github.com/gin-gonic/gin"
)

func FlushSandboxData(ctx *gin.Context) {

	err := queues.FlushSandboxData()
	if err != nil {
		resp := model.NewErrorResponse(http.StatusBadRequest, "error while flushing sandbox data", err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp := model.NewSuccessResponse(extras.ERR_SUCCESS, "Sandbox data flushed successfully.")
	ctx.JSON(resp.StatusCode, resp)
}

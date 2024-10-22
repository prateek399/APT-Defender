package controller

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func FirmwareUpdate(ctx *gin.Context) {
	if ctx.Request.ParseMultipartForm(extras.MAX_ALLOWED_FIRMWARE_FILE_SIZE+1) != nil { //4.5GB
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, extras.ERR_WHILE_PARSING_CONTENT, extras.ErrWhileParsingContent))
		return
	}

	resp := service.FirmwareUpdate(ctx.Request.MultipartForm, ctx)
	ctx.JSON(resp.StatusCode, resp)

	if resp.StatusCode != http.StatusOK {
		return
	}

	go func() {
		time.Sleep(3 * time.Second)
		service.StartBoot()
	}()
	ctx.JSON(http.StatusOK, model.NewSuccessResponse("Firmware update started successfully.", nil))
}

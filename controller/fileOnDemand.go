package controller

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateFileOnDemand(ctx *gin.Context) {
	var resp model.APIResponse

	session, err := auth.Store.Get(ctx.Request, "sessionid")
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SESSION_INVALID, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	if err := ctx.Request.ParseMultipartForm(extras.MAX_ALLOWED_FILE_SIZE + 1); err != nil { //100MB
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	// var filename string
	// if len(ctx.Request.MultipartForm.File["filename"]) > 0 {
	// 	filename = ctx.Request.MultipartForm.File["filename"][0].Filename
	// }

	// defer service.CreateAuditLogs(&resp, fmt.Sprintf("Added a new file for scanning - %s", filename), "SCAN TASKS", session.Values["admin_name"].(string))

	resp = service.CreateFileOnDemand(ctx.Request.MultipartForm, session.Values["admin_name"].(string), session.Values["organization_id"].(string))
	ctx.JSON(resp.StatusCode, resp)
}

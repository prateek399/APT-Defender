package controller

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateUrlOnDemand(ctx *gin.Context) {
	var resp model.APIResponse
	var urlRequest model.UrlOnDemand

	session, _ := auth.Store.Get(ctx.Request, "sessionid")

	if err := ctx.ShouldBindJSON(&urlRequest); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	// defer service.CreateAuditLogs(&resp, fmt.Sprintf("Added a new url for scanning - %s", urlRequest.UrlName), "SCAN TASKS", session.Values["admin_name"].(string))

	resp = service.CreateUrlOnDemand(urlRequest, session.Values["admin_name"].(string), session.Values["organization_id"].(string))
	ctx.JSON(resp.StatusCode, resp)
}

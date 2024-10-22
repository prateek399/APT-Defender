package controller

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetReport(ctx *gin.Context) {
	var resp model.APIResponse

	jobId := strings.TrimSpace(ctx.Query("job_id"))
	actionType := strings.TrimSpace(ctx.Query("type"))

	if jobId == extras.EMPTY_STRING {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_JOB_ID, extras.ErrInvalidJobId)
		ctx.JSON(resp.StatusCode, resp)
	}

	resp = service.GetReport(jobId, actionType)
	ctx.JSON(resp.StatusCode, resp)
}

func DownloadReport(ctx *gin.Context) {
	var resp model.APIResponse

	jobId := strings.TrimSpace(ctx.Query("job_id"))
	action_type := strings.TrimSpace(ctx.Query("type"))

	if jobId == extras.EMPTY_STRING {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_JOB_ID, extras.ErrInvalidJobId)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	// session, _ := auth.Store.Get(ctx.Request, "sessionid")
	// defer service.CreateAuditLogs(&resp, fmt.Sprintf("%s downloaded %s's report", session.Values["admin_name"].(string), fod.FileName), "DOWNLOAD REPORT", session.Values["admin_name"].(string))

	resp = service.DownloadReport(jobId, action_type)
	if resp.StatusCode != http.StatusOK {
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	if _, err := os.Stat(extras.REPORT_DOWNLOADS_PATH); os.IsNotExist(err) {
		err := os.MkdirAll(extras.REPORT_DOWNLOADS_PATH, 0777)
		if err != nil {
			log.Println("Error creating report downloads path")
		}
	}

	pdfFilePath := fmt.Sprintf("%sReport_%s.pdf", extras.REPORT_DOWNLOADS_PATH, jobId)

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", pdfFilePath))
	ctx.Header("Content-Type", "application/pdf")
	ctx.File(pdfFilePath)

	if err := os.Remove(pdfFilePath); err != nil {
		log.Println("Error removing pdf file")
	}
}

package controller

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func Backup(ctx *gin.Context) {
	var resp model.APIResponse
	// Check if the file exists
	stat, err := os.Stat(extras.CONFIG_FILE_NAME)
	if os.IsNotExist(err) {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FILE_NOT_FOUND, extras.ErrFileNotFound)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	backupTime := model.BackupTime{
		Time: time.Now(),
	}

	// session, _ := auth.Store.Get(ctx.Request, "sessionid")

	if err := dao.SaveProfile([]interface{}{backupTime}, extras.POST); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	mergedConfig := config.MergeInterfaceAndConfigData("BACKUP")

	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename="+time.Now().Format("2006 Jan 02 15:04:05")+".yaml")

	// Set the progress bar information in the response header
	ctx.Header("X-Progress", "0")
	ctx.Header("X-Progress-Max", strconv.FormatInt(stat.Size(), 10))

	// Send the file as the response
	bytesSent := int64(0)

	mergedConfigDataInBytes, err := yaml.Marshal(&mergedConfig)
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	if _, err := io.Copy(ctx.Writer, bytes.NewReader(mergedConfigDataInBytes)); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FILE_NOT_FOUND, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	// Update the progress bar information in the response header
	// service.CreateAuditLogs(&resp, fmt.Sprintf("%s backed up configuration", session.Values["admin_name"].(string)), "BACKUP", session.Values["admin_name"].(string))
	ctx.Header("X-Progress", strconv.FormatInt(bytesSent, 10))
	ctx.Status(http.StatusOK)
}

func Restore(ctx *gin.Context) {
	var resp model.APIResponse
	var err error

	curUsr, usrResp := service.GetCurUsr(ctx)
	if usrResp.StatusCode != http.StatusOK {
		ctx.JSON(usrResp.StatusCode, usrResp)
		return
	}

	if err = ctx.Request.ParseMultipartForm(extras.MAX_ALLOWED_FILE_SIZE + 1); err != nil { //100MB
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_WHILE_PARSING_CONTENT, extras.ErrWhileParsingContent)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	session, _ := auth.Store.Get(ctx.Request, "sessionid")
	defer service.CreateAuditLogs(&resp, fmt.Sprintf("%s restored configuration from backup", session.Values["admin_name"].(string)), "RESTORE", session.Values["admin_name"].(string))

	resp = service.Restore(ctx.Request.MultipartForm, curUsr)
	ctx.JSON(resp.StatusCode, resp)
}

func Erase(ctx *gin.Context) {

	curUsr, usrResp := service.GetCurUsr(ctx)
	if usrResp.StatusCode != http.StatusOK {
		ctx.JSON(usrResp.StatusCode, usrResp)
		return
	}

	resp := service.EraseConfig(extras.ERASE, curUsr)
	ctx.JSON(resp.StatusCode, resp)
}

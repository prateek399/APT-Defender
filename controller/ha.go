package controller

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/model/interface_model"
	"anti-apt-backend/service"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func CreateHa(ctx *gin.Context) {

	var req interface_model.CreateHaRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp := service.CreateHa(req)
	ctx.JSON(resp.StatusCode, resp)
}

func GetHa(ctx *gin.Context) {
	resp := service.GetHa()
	ctx.JSON(resp.StatusCode, resp)
}

func CompareDeviceInfoFromAnotherAppliance(ctx *gin.Context) {

	var req model.HaDeviceInfo
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	err := service.CompareDeviceInfoFromAnotherAppliance(req)
	if err != nil {
		resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}
	resp := model.NewSuccessResponse(extras.ERR_SUCCESS, "Device info compared successfully.")
	ctx.JSON(resp.StatusCode, resp)
}

func HaCopyConfig(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodPost {
		ctx.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method Not Allowed"})
		return
	}

	err := ctx.Request.ParseMultipartForm(extras.MAX_ALLOWED_FILE_SIZE + 1)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error parsing form: %v", err)})
		return
	}

	file, handler, err := ctx.Request.FormFile("configFile")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error retrieving file from form: %v", err)})
		return
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error reading file contents: %v", err)})
		return
	}

	configFilePath := extras.DATABASE_PATH + "ha/"

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		err = os.Mkdir(configFilePath, 0755)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error creating directory: %v", err)})
			return
		}
	}

	f, err := os.Create(filepath.Join(configFilePath, handler.Filename))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error creating config.yaml file: %v", err)})
		return
	}
	defer f.Close()

	_, err = io.Copy(f, bytes.NewReader(buf))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error copying file contents: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "config file received and saved successfully."})
}

func SyncBackup(ctx *gin.Context) {
	err := service.SyncBackup()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error syncing backup: %v", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Sync backup successful"})
}

func GenerateKeepalivedConfigForBackup(ctx *gin.Context) {
	err := service.GenerateKeepalivedConfigForBackupNsetIp()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Keepalived config generated successfully"})
}

func DisableHaInAnotherAppliance(ctx *gin.Context) {
	resp := service.DisableHa()
	if resp.StatusCode != http.StatusOK {
		ctx.JSON(resp.StatusCode, resp)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "HA disabled successfully"})
}

func UpdateLastSyncedAt(ctx *gin.Context) {

	var req model.LastSyncedAt
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error binding request: %v", err)})
		return
	}

	err := service.UpdateLastSyncedAt(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error updating last synced at: %v", err)})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Last synced at updated successfully"})
}

package controller

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UploadBuild(ctx *gin.Context) {
	if ctx.Request.ParseMultipartForm(extras.MAX_ALLOWED_FIRMWARE_FILE_SIZE+1) != nil { //4.5GB
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewErrorResponse(http.StatusBadRequest, extras.ERR_WHILE_PARSING_CONTENT, extras.ErrWhileParsingContent))
		return
	}
	// Get the file from the form
	// file, _, err := ctx.Request.FormFile("filename")
	// if err != nil {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }
	// defer file.Close()

	// // Create a directory to save the uploaded file
	// dir := "/home/prateek/Desktop/apt-backend"
	// if _, err := os.Stat(dir); os.IsNotExist(err) {
	// 	err = os.Mkdir(dir, 0755)
	// 	if err != nil {
	// 		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
	// 		return
	// 	}
	// }

	// // Create a new file in the directory with the same name as the uploaded file
	// filename := filepath.Join(dir, "temp")
	// out, err := os.Create(filename)
	// if err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
	// 	return
	// }
	// defer out.Close()

	// // Copy the uploaded file to the new file
	// _, err = io.Copy(out, file)
	// if err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
	// 	return
	// }

	// ctx.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "filename": filename})
	file, headers, err := ctx.Request.FormFile("filename")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	resp := service.UploadBuild(file, headers.Filename)
	ctx.JSON(resp.StatusCode, resp)
}

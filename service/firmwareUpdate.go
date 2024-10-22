package service

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func FirmwareUpdate(form *multipart.Form, ctx *gin.Context) model.APIResponse {
	if form.File == nil || form.File["file"] == nil || len(form.File["file"]) == 0 {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_REQUIRED_FIELD_EMPTY, extras.ErrRequiredFieldEmpty)
	}

	if filepath.Ext(form.File["file"][0].Filename) != ".img" {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_EXTENSION_NOT_SUPPORTED, extras.ErrExtensionNotSupported)
	}

	inputf, err := form.File["file"][0].Open()
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
	}
	defer inputf.Close()

	// buf := new(bytes.Buffer)
	// buf.ReadFrom(inputf)
	// mime, err := mimetype.DetectReader(bytes.NewReader(buf.Bytes()))
	// if err != nil {
	// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	// }
	// fmt.Println("Mime extension: ", mime.Extension())

	outputf, err := os.Create(extras.FIRMWARE_FILE_PATH + form.File["file"][0].Filename)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	inputFileSize := form.File["file"][0].Size
	writtenBytes := int64(0)
	buffer := make([]byte, extras.CHUNK_SIZE) //Setting to 1MB
	for {
		n, err := inputf.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		}

		if n > 0 {
			if _, err := outputf.Write(buffer[:n]); err != nil {
				return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			}
		}

		writtenBytes += int64(n)
		progress := float64(writtenBytes) / float64(inputFileSize) * 100
		fmt.Printf("\rprogress: %.2f%%", progress)

		if n == 0 {
			break
		}
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, "File uploaded successfully")
}

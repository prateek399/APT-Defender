package controller

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateFileOnDemandForFirewall(ctx *gin.Context) {
	var resp model.APIResponse

	if err := ctx.Request.ParseMultipartForm(extras.MAX_ALLOWED_FILE_SIZE + 1); err != nil { //100MB
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp = service.CreateFileOnDemand(ctx.Request.MultipartForm, "DEVICE", ctx.ClientIP())
	jobID := ""
	if resp.StatusCode == http.StatusOK {
		jobID = resp.Data.(string)
	}
	ctx.JSON(resp.StatusCode, jobID)
}

func CreateUrlOnDemandForFirewall(ctx *gin.Context) {
	var resp model.APIResponse
	var urlRequest model.UrlOnDemand
	if err := ctx.ShouldBindJSON(&urlRequest); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp = service.CreateUrlOnDemand(urlRequest, "DEVICE", "")
	jobID := ""
	if resp.StatusCode == http.StatusOK {
		jobID = resp.Data.(string)
	}
	ctx.JSON(resp.StatusCode, jobID)
}

func CreateJobForFw(ctx *gin.Context) {
	switch ctx.Query("req_type") {
	case "1":
		CheckHashOnDemand(ctx)
	case "2":
		CreateFileOnDemandForFirewall(ctx)
	case "3":
		CreateUrlOnDemandForFirewall(ctx)
	}
}

func CheckHashOnDemand(ctx *gin.Context) {
	resp := service.FileFromFireWall(ctx.Query("hash"))
	ctx.JSON(http.StatusOK, resp)
}

func TestAPTFw(ctx *gin.Context) {
	type PaylOad struct {
		Verdict string `json:"verdict"`
		TaskId  int    `json:"taskID"`
	}
	payload := PaylOad{}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// slog.Println("Verdict: ", payload.Verdict)

	ctx.JSON(http.StatusOK, gin.H{"message": payload})
}

// func SendAcknowledgement(ctx *gin.Context) {
// 	resp := service.SendAcknowledgement()
// }

// func GetJobForFw(ctx *gin.Context) {
// 	jobID := strings.TrimSpace(ctx.Query("jobid"))
// 	resp := service.FetchJobIDForFw(jobID)
// 	ctx.JSON(http.StatusOK, resp)
// }

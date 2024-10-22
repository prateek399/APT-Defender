package service

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// import (
// 	"anti-apt-backend/dao"
// 	"anti-apt-backend/extras"
// 	"anti-apt-backend/logger"
// 	"anti-apt-backend/model"
// 	"anti-apt-backend/service/hash"
// 	"anti-apt-backend/util"
// 	"log"
// 	"net/http"
// 	"strings"
// 	"time"

// 	"github.com/gin-gonic/gin"
// )

func OverrideVerdict(req model.OverrideVerdictRequest, ctx *gin.Context) model.APIResponse {
	var resp model.APIResponse

	curUsr, usrResp := GetCurUsr(ctx)
	if usrResp.StatusCode != http.StatusOK {
		return usrResp
	}

	// curUsr := "test-admin"

	switch strings.ToLower(req.Type) {
	case "file":
		resp = overrideFileVerdict(req, curUsr)
		if resp.StatusCode != http.StatusOK {
			return resp
		}
	case "url":
		resp = overrideUrlVerdict(req, curUsr)
		if resp.StatusCode != http.StatusOK {
			return resp
		}
	default:
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_TYPE_IN_OVERRIDE, extras.ErrInvalidTypeInOverride)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Successfully overridden the final verdict")
	return resp
}

func overrideFileVerdict(req model.OverrideVerdictRequest, updateBy string) model.APIResponse {
	var resp model.APIResponse

	jobId := req.JobID

	queryString := fmt.Sprintf("SELECT * FROM %s WHERE id = %d", extras.FileOnDemandTable, jobId)

	var fod model.FileOnDemand
	db := config.Db
	fileOnDemand := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &fod,
	}

	err := dao.GormOperations(&fileOnDemand, db, dao.EXEC)
	if err != nil {
		logger.LogAccToTaskId(jobId, fmt.Sprintf("ERROR WHILE FETCHING FOD, ERROR: %v", err))
	}

	if fod.Rating == extras.EMPTY_STRING {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_STILL_ANALYSING, extras.ErrReportNotGenerated)
		return resp
	}

	verdictReq := strings.ToLower(strings.TrimSpace(req.Verdict))

	if verdictReq != extras.ALLOW && verdictReq != extras.BLOCK {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_VERDICT, extras.ErrInvalidVerdict)
		return resp
	}

	if fod.FinalVerdict != verdictReq {

		queryString := fmt.Sprintf("UPDATE %s SET final_verdict = '%s', overridden_verdict = true, overridden_by = '%s' WHERE (md5 = '%s' or sha = '%s' or sha256 = '%s')", extras.FileOnDemandTable, verdictReq, updateBy, fod.Md5, fod.SHA, fod.SHA256)
		log.Printf("queryString: %s", queryString)
		fileOnDemand := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{queryString},
		}

		err = dao.GormOperations(&fileOnDemand, config.Db, dao.EXEC)
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
			// logger.LoggerFunc("error", logger.LoggerMessage(err.Error()))
			return resp
		}

		// go func() {
		// 	err = hash.SaveVerdict(fod.Md5, verdictReq)
		// 	if err != nil {
		// 		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		// 		logger.LoggerFunc("error", logger.LoggerMessage(err.Error()))
		// 	}
		// }()

	} else {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_ALREADY_OVERRIDE, extras.ErrAlreadyOverridden)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Successfully overridden the final verdict")
	return resp
}

func overrideUrlVerdict(req model.OverrideVerdictRequest, updateBy string) model.APIResponse {
	var resp model.APIResponse
	jobId := req.JobID

	queryString := fmt.Sprintf("SELECT * FROM %s WHERE id = %d", extras.UrlOnDemandTable, jobId)

	var uod model.UrlOnDemand
	db := config.Db
	urlOnDemand := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &uod,
	}

	err := dao.GormOperations(&urlOnDemand, db, dao.EXEC)
	if err != nil {
		logger.LogAccToTaskId(jobId, fmt.Sprintf("ERROR WHILE FETCHING FOD, ERROR: %v", err))
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	}

	if uod.Rating == extras.EMPTY_STRING {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_STILL_ANALYSING, extras.ErrReportNotGenerated)
		return resp
	}

	verdictReq := strings.ToLower(strings.TrimSpace(req.Verdict))

	if verdictReq != extras.ALLOW && verdictReq != extras.BLOCK {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_VERDICT, extras.ErrInvalidVerdict)
		return resp
	}

	if uod.FinalVerdict != verdictReq {

		queryString := fmt.Sprintf("UPDATE %s SET final_verdict = '%s', overridden_verdict = true, overridden_by = '%s' WHERE (url_name = '%s')", extras.UrlOnDemandTable, verdictReq, updateBy, uod.UrlName)

		urlOnDemand := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{queryString},
		}

		err = dao.GormOperations(&urlOnDemand, config.Db, dao.EXEC)
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
			// logger.LoggerFunc("error", logger.LoggerMessage(err.Error()))
			return resp
		}

	} else {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_ALREADY_OVERRIDE, extras.ErrAlreadyOverridden)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Successfully overridden the final verdict")
	return resp
}

func GetOverriddenVerdictLogs(ctx *gin.Context) model.APIResponse {
	var resp model.APIResponse

	session, err := auth.Store.Get(ctx.Request, "sessionid")
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage(err.Error()))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SESSION_INVALID, err)
		return resp
	}

	if _, found := session.Values["admin_name"].(string); !found {
		// logger.LoggerFunc("warn", logger.LoggerMessage(extras.ERR_SESSION_INVALID))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SESSION_INVALID, extras.ErrUnauthorizedUser)
		return resp
	}

	// i need to fetch from both fodTable and uodTable whose overridden_verdict is true
	var fods []model.FileOnDemand
	queryString := fmt.Sprintf("SELECT * FROM %s WHERE overridden_verdict = true", extras.FileOnDemandTable)

	fileOnDemand := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &fods,
	}

	err = dao.GormOperations(&fileOnDemand, config.Db, dao.EXEC)
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	var uods []model.UrlOnDemand
	queryString = fmt.Sprintf("SELECT * FROM %s WHERE overridden_verdict = true", extras.UrlOnDemandTable)

	urlOnDemand := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &uods,
	}

	err = dao.GormOperations(&urlOnDemand, config.Db, dao.EXEC)
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	var overriddenVerdicts []model.OverriddenVerdict

	for _, fod := range fods {
		originalVerdict := util.GetVerdict(fod.Score)
		overriddenVerdicts = append(overriddenVerdicts, model.OverriddenVerdict{
			JobID:           fmt.Sprintf("%d", fod.Id),
			Filename:        fod.FileName,
			SubmittedBy:     fod.SubmittedBy,
			SubmittedTime:   fod.SubmittedTime.String(),
			OriginalVerdict: string(originalVerdict),
			FinalVerdict:    fod.FinalVerdict,
			UpdatedBy:       fod.OverriddenBy,
		})
	}

	for _, uod := range uods {
		// originalVerdict := util.GetVerdict(uod.Score)
		overriddenVerdicts = append(overriddenVerdicts, model.OverriddenVerdict{
			JobID:         fmt.Sprintf("%d", uod.Id),
			URL:           uod.UrlName,
			SubmittedBy:   uod.SubmittedBy,
			SubmittedTime: uod.SubmittedTime.String(),
			// OriginalVerdict: string(originalVerdict),
			FinalVerdict: uod.FinalVerdict,
			UpdatedBy:    uod.OverriddenBy,
		})
	}

	// sort by submitted time
	sort.Slice(overriddenVerdicts, func(i, j int) bool {
		return overriddenVerdicts[i].SubmittedTime > overriddenVerdicts[j].SubmittedTime
	})

	// slog.Println("overriddenVerdicts: ", overriddenVerdicts)

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, overriddenVerdicts)
	return resp
}

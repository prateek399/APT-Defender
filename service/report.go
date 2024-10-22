package service

import (
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/jung-kurt/gofpdf"
	"github.com/levenlabs/golib/timeutil"
)

func GetReport(jobId string, actionType string) model.APIResponse {
	var resp model.APIResponse
	var jobReport model.JobInfo
	var urlJobReport model.UrlJobInfo

	if actionType == "file" {

		taskId, err := strconv.Atoi(jobId)
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_JOB_ID, extras.ErrInvalidJobId)
			return resp
		}

		queryString := fmt.Sprintf("SELECT * FROM %s WHERE id = %d", extras.FileOnDemandTable, taskId)

		var fod model.FileOnDemand
		db := config.Db
		fileOnDemand := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{queryString},
			Result:       &fod,
		}

		err = dao.GormOperations(&fileOnDemand, db, dao.EXEC)
		if err != nil {
			// logger.LogAccToTaskId(taskId, fmt.Sprintf("ERROR WHILE FETCHING FOD, ERROR: %v", err))
		}

		if fod.Rating == extras.EMPTY_STRING {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_STILL_ANALYSING, extras.ErrReportNotGenerated)
			return resp
		}

		submitType := "WiJungle Anti-APT Dashboard"
		if fod.FromDevice {
			submitType = "WiJungle Firewall Device"
		}

		totalScanTime := 0
		startedTime := ""
		endTime := ""
		vm := ""
		if fod.Status == extras.PREVIOUSLY_SCANNED_FILE {
			totalScanTime = 1
			// submitTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", url.SubmittedTime)
			startedTime = fod.SubmittedTime.Format(extras.TIME_FORMAT)
			endTime = fod.FinishedTime.Time.Format(extras.TIME_FORMAT)
			vm = "VM not assigned(found from hash)"
		} else {
			var duration int
			if fod.FinishedTime.Valid {
				duration = int(fod.FinishedTime.Time.Sub(fod.SubmittedTime).Seconds())
			}
			totalScanTime = duration
			startedTime = fod.SubmittedTime.Format(extras.TIME_FORMAT)
			endTime = fod.FinishedTime.Time.Format(extras.TIME_FORMAT)
			// vm = fod.TaskReport.Info.Machine.(map[string]any)["name"].(string)
			vm = "WiJungle Sandbox"
		}

		final_verdict := ""
		if fod.OverriddenVerdict {
			final_verdict = fod.FinalVerdict + " ( overridden by " + fod.OverriddenBy + " )"
		} else {
			final_verdict = fod.FinalVerdict
		}

		jobSummary := model.JobSummary{
			JobID:         jobId,
			Status:        "reported",
			ReceivedTime:  fod.SubmittedTime.Format(extras.TIME_FORMAT),
			RatedBy:       "WiJungle Anti-APT",
			SubmitType:    submitType,
			VmScanTimeout: 100,
			Rating:        string(fod.Rating),
			FinalVerdict:  final_verdict,
		}

		jobDetail := model.JobDetail{
			Filename:      fod.FileName,
			ScanStartTime: startedTime,
			ScanEndTime:   endTime,
			TotalScanTime: totalScanTime,
			FileType:      fod.ContentType,
			// FileSize:        int(fod.FileName),
			MD5:             fod.Md5,
			SHA1:            fod.SHA,
			SHA256:          fod.SHA256,
			SubmittedBy:     fod.SubmittedBy,
			SubmitDevice:    submitType,
			SubmittedDevice: "WiJungle Anti-APT Sandbox",
			VM:              vm,
			VMReason:        "APT Sandboxing PreFiltering",
		}

		jobReport = model.JobInfo{
			Summary:  jobSummary,
			Details:  jobDetail,
			Filename: fod.FileName,
		}
	} else if actionType == "url" {

		taskId, err := strconv.Atoi(jobId)
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_JOB_ID, extras.ErrInvalidJobId)
			return resp
		}

		queryString := fmt.Sprintf("SELECT * FROM %s WHERE id = %d", extras.UrlOnDemandTable, taskId)

		var uod model.UrlOnDemand
		db := config.Db
		urlOnDemands := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{queryString},
			Result:       &uod,
		}

		err = dao.GormOperations(&urlOnDemands, db, dao.EXEC)
		if err != nil {
			// logger.LogAccToTaskId(taskId, fmt.Sprintf("ERROR WHILE FETCHING FOD, ERROR: %v", err))
		}

		if uod.Status != extras.REPORTED {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_STILL_ANALYSING, extras.ErrReportNotGenerated)
			return resp
		}

		submitType := "WiJungle Anti-APT Dashboard"
		if uod.FromDevice {
			submitType = "WiJungle Firewall Device"
		}

		totalScanTime := 0
		startedTime := ""
		endTime := ""
		vm := ""
		if uod.Status == extras.PREVIOUSLY_SCANNED_URL {
			totalScanTime = 1
			// submitTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", url.SubmittedTime)
			startedTime = uod.SubmittedTime.Format(extras.TIME_FORMAT)
			endTime = uod.FinishedTime.Time.Format(extras.TIME_FORMAT)
			vm = "VM not assigned(found from hash)"

		} else {
			var duration int
			if uod.FinishedTime.Valid {
				duration = int(uod.FinishedTime.Time.Sub(uod.SubmittedTime).Seconds())
			}
			totalScanTime = duration
			totalScanTime = duration
			startedTime = uod.SubmittedTime.Format(extras.TIME_FORMAT)
			endTime = uod.FinishedTime.Time.Format(extras.TIME_FORMAT)
			// vm = uod.TaskReport.Info.Machine.(map[string]any)["name"].(string)
			vm = "WiJungle Sandbox"
		}

		final_verdict := ""
		if uod.OverriddenVerdict {
			final_verdict = uod.FinalVerdict + " ( overridden by " + uod.OverriddenBy + " )"
		} else {
			final_verdict = uod.FinalVerdict
		}

		jobSummary := model.JobSummary{
			JobID:         jobId,
			Status:        "reported",
			ReceivedTime:  uod.SubmittedTime.Format(extras.TIME_FORMAT),
			RatedBy:       "WiJungle Anti-APT",
			SubmitType:    submitType,
			VmScanTimeout: 100,
			Rating:        string(uod.Rating),
			FinalVerdict:  final_verdict,
		}

		jobDetail := model.UrlJobDetail{
			Url:             uod.UrlName,
			ScanStartTime:   startedTime,
			ScanEndTime:     endTime,
			TotalScanTime:   totalScanTime,
			Type:            "URL",
			SubmittedBy:     uod.SubmittedBy,
			SubmitDevice:    submitType,
			SubmittedDevice: "WiJungle Anti-APT Sandbox",
			VM:              vm,
			VMReason:        "APT Sandboxing PreFiltering",
		}

		urlJobReport = model.UrlJobInfo{
			Summary: jobSummary,
			Details: jobDetail,
			Url:     uod.UrlName,
		}

	} else {
		resp = model.NewErrorResponse(http.StatusBadRequest, "Invalid action type", extras.ErrInvalidActionType)
		return resp
	}

	if actionType == "url" {
		resp = model.NewSuccessResponse(extras.ERR_SUCCESS, urlJobReport)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, jobReport)
	return resp
}

func DownloadReport(jobId string, actionType string) model.APIResponse {

	resp := GetReport(jobId, actionType)

	if resp.StatusCode != http.StatusOK {
		return resp
	}

	if resp.Data == nil {
		resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_STILL_ANALYSING, extras.ErrReportNotGenerated)
		return resp
	}

	if actionType == "url" {
		jobInfo, ok := resp.Data.(model.UrlJobInfo)
		if !ok {
			resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_STILL_ANALYSING, extras.ErrReportNotGenerated)
			return resp
		}

		err := generatePDFurl(jobInfo)
		if err != nil {
			resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_WHILE_ANALYSING, err)
			return resp
		}

	} else if actionType == "file" {
		jobInfo, ok := resp.Data.(model.JobInfo)
		if !ok {
			resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_STILL_ANALYSING, extras.ErrReportNotGenerated)
			return resp
		}

		err := generatePDF(jobInfo)
		if err != nil {
			resp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_WHILE_ANALYSING, err)
			return resp
		}

	} else {
		resp := model.NewErrorResponse(http.StatusBadRequest, "Invalid action type", extras.ErrInvalidActionType)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "PDF generated successfully")
	return resp
}

func generatePDF(jobInfo model.JobInfo) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Times", "B", 20)

	pdf.SetMargins(10, 10, 10)

	pdf.AddPage()

	pdf.SetTextColor(0, 64, 128)
	pdf.CellFormat(0, 8, "WiJungle Anti-APT Detailed Report", "0", 0, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Times", "I", 14)

	pdf.SetTextColor(0, 64, 128)
	pdf.CellFormat(0, 8, "Summary", "0", 0, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Times", "", 10)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetFillColor(200, 220, 255)

	jobInfoType := reflect.TypeOf(jobInfo.Summary)
	for i := 0; i < jobInfoType.NumField(); i++ {
		field := jobInfoType.Field(i)
		value := reflect.ValueOf(jobInfo.Summary).Field(i)

		if field.Name == "Rating" {
			pdf.CellFormat(40, 8, field.Name+":", "1", 0, "L", true, 0, "")
			switch value.Interface() {
			case "Clean":
				pdf.SetTextColor(0, 128, 0)
			case "Low Risk":
				pdf.SetTextColor(255, 255, 0)
			case "Medium Risk":
				pdf.SetTextColor(255, 165, 0)
			case "High Risk":
				pdf.SetTextColor(255, 0, 0)
			}
			pdf.CellFormat(0, 8, fmt.Sprintf("%v", value.Interface()), "1", 1, "L", true, 0, "")
			pdf.SetTextColor(0, 0, 0)
			continue
		}

		pdf.CellFormat(40, 8, field.Name+":", "1", 0, "L", true, 0, "")
		pdf.CellFormat(0, 8, fmt.Sprintf("%v", value.Interface()), "1", 1, "L", true, 0, "")
	}

	pdf.SetTextColor(0, 64, 128)
	pdf.Ln(5)
	pdf.SetFont("Times", "I", 14)

	pdf.CellFormat(0, 8, "Details", "0", 0, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Times", "", 10)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(220, 230, 255)

	detailsType := reflect.TypeOf(jobInfo.Details)
	for i := 0; i < detailsType.NumField(); i++ {
		field := detailsType.Field(i)
		value := reflect.ValueOf(jobInfo.Details).Field(i)
		pdf.CellFormat(40, 8, field.Name+":", "1", 0, "L", true, 0, "")
		pdf.CellFormat(0, 8, fmt.Sprintf("%v", value.Interface()), "1", 1, "L", true, 0, "")
	}

	pdf.Ln(5)
	pdf.SetFont("Times", "B", 12)

	pdf.SetTextColor(0, 64, 128)

	resp := GetDevice().Data.(*model.DeviceSpecification)
	firmare := resp.OsVersion

	pdf.CellFormat(0, 10, "OS Version: "+firmare, "0", 0, "C", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(0, 10, "Job ID: "+jobInfo.Summary.JobID, "0", 0, "C", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(0, 10, "Generated at: "+timeutil.TimestampNow().Format("2006-01-02 15:04:05"), "0", 0, "C", false, 0, "")

	pdfFilePath := fmt.Sprintf("%sReport_%s.pdf", extras.REPORT_DOWNLOADS_PATH, jobInfo.Summary.JobID)

	if err := os.MkdirAll(extras.REPORT_DOWNLOADS_PATH, 0777); err != nil {
		return err
	}

	err := pdf.OutputFileAndClose(pdfFilePath)
	if err != nil {
		return err
	}

	return nil
}

func generatePDFurl(jobInfo model.UrlJobInfo) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Times", "B", 20)

	pdf.SetMargins(10, 10, 10)

	pdf.AddPage()

	pdf.SetTextColor(0, 64, 128)
	pdf.CellFormat(0, 8, "WiJungle Anti-APT Detailed Report", "0", 0, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Times", "I", 14)

	pdf.SetTextColor(0, 64, 128)
	pdf.CellFormat(0, 8, "Summary", "0", 0, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Times", "", 10)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetFillColor(200, 220, 255)

	jobInfoType := reflect.TypeOf(jobInfo.Summary)
	for i := 0; i < jobInfoType.NumField(); i++ {
		field := jobInfoType.Field(i)
		value := reflect.ValueOf(jobInfo.Summary).Field(i)

		if field.Name == "Rating" {
			pdf.CellFormat(40, 8, field.Name+":", "1", 0, "L", true, 0, "")
			switch value.Interface() {
			case "Clean":
				pdf.SetTextColor(0, 128, 0)
			case "Low Risk":
				pdf.SetTextColor(255, 255, 0)
			case "Medium Risk":
				pdf.SetTextColor(255, 165, 0)
			case "High Risk":
				pdf.SetTextColor(255, 0, 0)
			}
			pdf.CellFormat(0, 8, fmt.Sprintf("%v", value.Interface()), "1", 1, "L", true, 0, "")
			pdf.SetTextColor(0, 0, 0)
			continue
		}

		pdf.CellFormat(40, 8, field.Name+":", "1", 0, "L", true, 0, "")
		pdf.CellFormat(0, 8, fmt.Sprintf("%v", value.Interface()), "1", 1, "L", true, 0, "")
	}

	pdf.SetTextColor(0, 64, 128)
	pdf.Ln(5)
	pdf.SetFont("Times", "I", 14)

	pdf.CellFormat(0, 8, "Details", "0", 0, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Times", "", 10)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(220, 230, 255)

	detailsType := reflect.TypeOf(jobInfo.Details)
	for i := 0; i < detailsType.NumField(); i++ {
		field := detailsType.Field(i)
		value := reflect.ValueOf(jobInfo.Details).Field(i)
		pdf.CellFormat(40, 8, field.Name+":", "1", 0, "L", true, 0, "")
		pdf.CellFormat(0, 8, fmt.Sprintf("%v", value.Interface()), "1", 1, "L", true, 0, "")
	}

	pdf.Ln(5)
	pdf.SetFont("Times", "B", 12)

	pdf.SetTextColor(0, 64, 128)

	resp := GetDevice().Data.(*model.DeviceSpecification)
	firmare := resp.OsVersion

	pdf.CellFormat(0, 10, "OS Version: "+firmare, "0", 0, "C", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(0, 10, "Job ID: "+jobInfo.Summary.JobID, "0", 0, "C", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(0, 10, "Generated at: "+timeutil.TimestampNow().Format("2006-01-02 15:04:05"), "0", 0, "C", false, 0, "")

	pdfFilePath := fmt.Sprintf("%sReport_%s.pdf", extras.REPORT_DOWNLOADS_PATH, jobInfo.Summary.JobID)

	if err := os.MkdirAll(extras.REPORT_DOWNLOADS_PATH, 0777); err != nil {
		return err
	}
	err := pdf.OutputFileAndClose(pdfFilePath)
	if err != nil {
		return err
	}

	return nil
}

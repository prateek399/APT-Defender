package service

import (
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/hash"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func CreateFileOnDemand(formRequest *multipart.Form, adminName, ip string) model.APIResponse {
	var err error
	var resp model.APIResponse
	respMes := "File successfully uploaded"

	if formRequest.File == nil || len(formRequest.File) == 0 {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_REQUIRED_FIELD_EMPTY, extras.ErrRequiredFieldEmpty)
		return resp
	}

	func() {
		file, err := os.Open(extras.CUCKOO_CONF_FILE_PATH)
		if err != nil {
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "upload_max_size") {
				parts := strings.Split(line, "=")
				if len(parts) >= 2 {
					v, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
					if err != nil {
						return
					}
					extras.MAX_ALLOWED_FILE_SIZE = v
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return
		}
	}()

	if formRequest.File["filename"][0].Size > extras.MAX_ALLOWED_FILE_SIZE {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FILE_TOO_LARGE, extras.ErrFileTooLarge)
		return resp
	}

	if formRequest.Value["comments"] == nil || len(formRequest.Value["comments"]) == 0 || formRequest.Value["comments"][0] == "undefined" {
		if len(formRequest.Value["comments"]) > 0 && formRequest.Value["comments"][0] == "undefined" {
			formRequest.Value["comments"][0] = ""
		}
		formRequest.Value["comments"] = append(formRequest.Value["comments"], "")
	}

	var scanProfile []model.ScanProfile
	scanProfile, err = dao.FetchScanProfile(map[string]any{})
	if err == extras.ErrNoRecordForScanProfile {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_RECORD_NOT_FOUND, extras.ErrNoRecordForScanProfile)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	err = util.ValidateContentTypeOfFile(formRequest, scanProfile[0])
	if err == extras.ErrInvalidContentType {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_CONTENT_TYPE, extras.ErrInvalidContentType)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
		return resp
	}

	var fromDevice bool
	if adminName == "DEVICE" {
		fromDevice = true
	}

	// var platForm = "Windows 7"
	// platFormInBytes, _ := os.ReadFile(extras.PLATFORM_FILE_NAME)
	// if strings.Contains(strings.ToLower(string(platFormInBytes)), "ubuntu") {
	// 	platForm = "Ubuntu 20"
	// }

	fodId := 0
	queryString := fmt.Sprintf("SELECT id FROM %s ORDER BY id DESC LIMIT 1", extras.FileOnDemandTable)
	fodRepo := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &fodId,
	}
	err = dao.GormOperations(&fodRepo, config.Db, "exec")
	if err != nil {
		// slog.Println("ERROR FROM DATABASE: ", err)
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}
	fodId++

	fod := model.FileOnDemand{
		Id:            fodId,
		FileName:      formRequest.File["filename"][0].Filename,
		SubmittedTime: time.Now(),
		SubmittedBy:   adminName,
		Comments:      formRequest.Value["comments"][0],
		ContentType:   formRequest.File["filename"][0].Header.Get("Content-Type"),
		// OsSupported:   platForm,
		FileCount:  1,
		FromDevice: fromDevice,
	}

	if _, err := os.Stat(extras.SANDBOX_FILE_PATHS); os.IsNotExist(err) {
		if err := os.Mkdir(extras.SANDBOX_FILE_PATHS, 0755); err != nil {
			// slog.Println("ERROR IN CREATING DIRECTORY: ", err)
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
		}
	}

	inputf, err := formRequest.File["filename"][0].Open()
	if err != nil {
		// slog.Println("ERROR IN OPENING FILE: ", err)
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
	}
	defer inputf.Close()

	fp := fmt.Sprintf("%s/files/%d", extras.DATABASE_PATH, fodId)
	tempOutputf, err := os.Create(fp)
	if err != nil {
		// slog.Println("ERROR IN CREATING TEMP FILE: ", err)
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
	}
	defer tempOutputf.Close()

	buffer := make([]byte, 1024) // 1KB chunk size
	for {
		n, err := inputf.Read(buffer)
		if err != nil && err != io.EOF {
			// slog.Println("ERROR IN READING FILE: ", err)
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		}
		if n == 0 {
			break
		}
		if _, err := tempOutputf.Write(buffer[:n]); err != nil {
			// slog.Println("ERROR IN WRITING FILE: ", err)
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		}
	}

	// eicarBytes, _ := os.ReadFile(fp)

	// slog.Println(string(eicarBytes))
	if fromDevice {
		respMes = fmt.Sprintf("%d", fod.Id)
	} else {
		respMes = "File: " + fod.FileName + " successfully uploaded"
	}

	md5, _ := hash.CalculateHash(fp, "md5")
	sha1, _ := hash.CalculateHash(fp, "sha1")
	sha256, _ := hash.CalculateHash(fp, "sha256")
	fod.Md5 = md5
	fod.SHA = sha1
	fod.SHA256 = sha256

	resp = checkIfHashAlreadyPresent(fod, md5, sha1, sha256, ip)
	if resp.StatusCode == http.StatusOK {
		go deleteLocalTask(fod.Id)
		return resp
	}

	// if strings.Contains(strings.ToLower(string(eicarBytes)), "eicar") {
	// 	fod.Rating = string(model.Critical)
	// 	fod.FinalVerdict = extras.BLOCK
	// 	fod.Score = 2
	// 	fod.FinishedTime = sql.NullTime{Time: time.Now().Add(20 * time.Second), Valid: true}
	// 	queryString = fmt.Sprintf("INSERT INTO %s (id, file_name, content_type, submitted_time, submitted_by, comments, file_count, from_device, status, finished_time, rating, final_verdict, md5, sha, sha256, score) VALUES (%d, '%s', '%s', '%s', '%s', '%s', %d, %t, '%s', '%s', '%s', '%s', '%s', '%s', '%s', %f)", extras.FileOnDemandTable, fod.Id, fod.FileName, fod.ContentType, fod.SubmittedTime.Format(extras.TIME_FORMAT), fod.SubmittedBy, fod.Comments, fod.FileCount, fod.FromDevice, fod.Status, fod.FinishedTime.Time.Format(extras.TIME_FORMAT), fod.Rating, fod.FinalVerdict, fod.Md5, fod.SHA, fod.SHA256, fod.Score)
	// 	if fod.FromDevice {
	// 		fod.ClientIp = ip
	// 		queryString = fmt.Sprintf("INSERT INTO %s (id, file_name, content_type, submitted_time, submitted_by, comments, client_ip, file_count, from_device, status, finished_time, rating, final_verdict, md5, sha, sha256, score) VALUES (%d, '%s', '%s', '%s', '%s', '%s', '%s', %d, %t, '%s', '%s', '%s', '%s', '%s', '%s', '%s', %f)", extras.FileOnDemandTable, fod.Id, fod.FileName, fod.ContentType, fod.SubmittedTime.Format(extras.TIME_FORMAT), fod.SubmittedBy, fod.Comments, fod.ClientIp, fod.FileCount, fod.FromDevice, fod.Status, fod.FinishedTime.Time.Format(extras.TIME_FORMAT), fod.Rating, fod.FinalVerdict, fod.Md5, fod.SHA, fod.SHA256, fod.Score)
	// 	}

	// 	fodRepo = dao.DatabaseOperationsRepo{
	// 		QueryExecSet: []string{
	// 			queryString,
	// 			fmt.Sprintf("INSERT INTO %s (id, sandbox_id, aborted) VALUES (%d, %d, %t)", extras.TaskFinishedTable, fod.Id, 1, false)},
	// 	}

	// 	err = dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
	// 	if err != nil {
	// 		// slog.Println("ERROR WHILE INSERTING FOD: ", err)
	// 		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	// 	}
	// 	return model.NewSuccessResponse(extras.ERR_SUCCESS, respMes)
	// }

	var count int64 = 0
	fodRepo = dao.DatabaseOperationsRepo{
		QueryExecSet: []string{fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (md5 = '%s' OR sha = '%s' OR sha256 = '%s')", extras.TaskLiveAnalysisTable, fod.Md5, fod.SHA, fod.SHA256)},
		Result:       &count,
	}
	err = dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
	if err != nil {
		// slog.Println("ERROR FROM DATABASE: ", err)
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	queryString = fmt.Sprintf("INSERT INTO %s (id, file_name, content_type, submitted_time, submitted_by, comments, file_count, from_device, md5, sha, sha256) VALUES (%d, '%s', '%s', '%s', '%s', '%s', %d, %t, '%s', '%s', '%s')", extras.FileOnDemandTable, fod.Id, fod.FileName, fod.ContentType, fod.SubmittedTime.Format(extras.TIME_FORMAT), fod.SubmittedBy, fod.Comments, fod.FileCount, fromDevice, fod.Md5, fod.SHA, fod.SHA256)
	if fromDevice {
		fod.ClientIp = ip
		queryString = fmt.Sprintf("INSERT INTO %s (id, file_name, content_type, submitted_time, submitted_by, comments, client_ip, file_count, from_device, md5, sha, sha256) VALUES (%d, '%s', '%s', '%s', '%s', '%s', '%s', %d, %t, '%s', '%s', '%s')", extras.FileOnDemandTable, fod.Id, fod.FileName, fod.ContentType, fod.SubmittedTime.Format(extras.TIME_FORMAT), fod.SubmittedBy, fod.Comments, fod.ClientIp, fod.FileCount, fromDevice, fod.Md5, fod.SHA, fod.SHA256)
	}

	fodRepo = dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
	}

	if count <= 0 {
		queryString = fmt.Sprintf("INSERT INTO %s (id, status, queue_retry_count, running_retry_count, sandbox_retry_count, log_queue_failed, md5, sha, sha256) VALUES (%d, '%s', %d, %d, %d, %t, '%s', '%s', '%s')", extras.TaskLiveAnalysisTable, fod.Id, extras.PendingNotInQueue, 0, 0, 0, false, fod.Md5, fod.SHA, fod.SHA256)
		fodRepo.QueryExecSet = append(fodRepo.QueryExecSet, queryString)
	} else {
		queryString = fmt.Sprintf("INSERT INTO %s (id, md5, sha, sha256) VALUES (%d, '%s', '%s', '%s')", extras.TaskDuplicateTable, fod.Id, fod.Md5, fod.SHA, fod.SHA256)
		fodRepo.QueryExecSet = append(fodRepo.QueryExecSet, queryString)
	}

	err = dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
	if err != nil {
		// slog.Println("ERROR FROM DATABASE: ", err)
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, respMes)
	return resp
}

func checkIfHashAlreadyPresent(fod model.FileOnDemand, md5 string, sha1 string, sha256 string, ip string) model.APIResponse {
	var respMes string
	if fod.FromDevice {
		respMes = fmt.Sprintf("%d", fod.Id)
	} else {
		respMes = "File: " + fod.FileName + " successfully uploaded"
	}

	queryString := ""
	isClean, err := dao.IsCleanHashFromDb(md5, sha1, sha256)
	if err != nil {
		// slog.Println("ERROR WHILE CHECKING FOR CLEAN HASH: ", err)
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}
	if isClean {
		// slog.Println("ALREADY ANALYSED")
		fod.Status = extras.PREVIOUSLY_SCANNED_FILE
		fod.FinishedTime = sql.NullTime{Time: time.Now(), Valid: true}

		type VerdictAndRating struct {
			FinalVerdict string
			Rating       string
			Score        float32
		}
		var verdictAndRating VerdictAndRating

		queryString = fmt.Sprintf("SELECT final_verdict, rating, score FROM %s WHERE (md5 = '%s' OR sha = '%s' OR sha256 = '%s')", extras.FileOnDemandTable, md5, sha1, sha256)

		fodRepo := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{queryString},
			Result:       &verdictAndRating,
		}

		err := dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
		if err != nil || verdictAndRating.Rating == "" || verdictAndRating.FinalVerdict == "" {
			// slog.Println("ERROR FROM DATABASE: ", err)
			verdictAndRating.Rating = string(model.Clean)
			verdictAndRating.FinalVerdict = extras.ALLOW
			verdictAndRating.Score = 0
		}

		fod.Rating = verdictAndRating.Rating
		fod.FinalVerdict = verdictAndRating.FinalVerdict
		fod.Score = verdictAndRating.Score

		queryString = fmt.Sprintf("INSERT INTO %s (id, file_name, content_type, submitted_time, submitted_by, comments, file_count, from_device, status, finished_time, rating, final_verdict, md5, sha, sha256, score) VALUES (%d, '%s', '%s', '%s', '%s', '%s', %d, %t, '%s', '%s', '%s', '%s', '%s', '%s', '%s', %f)", extras.FileOnDemandTable, fod.Id, fod.FileName, fod.ContentType, fod.SubmittedTime.Format(extras.TIME_FORMAT), fod.SubmittedBy, fod.Comments, fod.FileCount, fod.FromDevice, fod.Status, fod.FinishedTime.Time.Format(extras.TIME_FORMAT), fod.Rating, fod.FinalVerdict, fod.Md5, fod.SHA, fod.SHA256, fod.Score)
		if fod.FromDevice {
			fod.ClientIp = ip
			queryString = fmt.Sprintf("INSERT INTO %s (id, file_name, content_type, submitted_time, submitted_by, comments, client_ip, file_count, from_device, status, finished_time, rating, final_verdict, md5, sha, sha256, score) VALUES (%d, '%s', '%s', '%s', '%s', '%s', '%s', %d, %t, '%s', '%s', '%s', '%s', '%s', '%s', '%s', %f)", extras.FileOnDemandTable, fod.Id, fod.FileName, fod.ContentType, fod.SubmittedTime.Format(extras.TIME_FORMAT), fod.SubmittedBy, fod.Comments, fod.ClientIp, fod.FileCount, fod.FromDevice, fod.Status, fod.FinishedTime.Time.Format(extras.TIME_FORMAT), fod.Rating, fod.FinalVerdict, fod.Md5, fod.SHA, fod.SHA256, fod.Score)
		}

		fodRepo = dao.DatabaseOperationsRepo{
			QueryExecSet: []string{queryString},
			Result:       &fod,
		}

		err = dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
		if err != nil {
			// slog.Println("ERROR WHILE INSERTING FOD: ", err)
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
		}
		return model.NewSuccessResponse(extras.ERR_SUCCESS, respMes)
	}

	isMalware, err := dao.IsMalwareHashFromDb(md5, sha1, sha256)
	if err != nil {
		// slog.Println("ERROR WHILE CHECKING FOR MALWARE HASH: ", err)
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}
	if isMalware {
		// slog.Println("ALREADY ANALYSED")
		fod.Status = extras.PREVIOUSLY_SCANNED_FILE
		fod.FinishedTime = sql.NullTime{Time: time.Now(), Valid: true}

		type VerdictAndRating struct {
			FinalVerdict string
			Rating       string
			Score        float32
		}
		var verdictAndRating VerdictAndRating

		queryString = fmt.Sprintf("SELECT final_verdict, rating, score FROM %s WHERE (md5 = '%s' OR sha = '%s' OR sha256 = '%s')", extras.FileOnDemandTable, md5, sha1, sha256)

		fodRepo := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{queryString},
			Result:       &verdictAndRating,
		}

		err := dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
		if err != nil || verdictAndRating.Rating == "" || verdictAndRating.FinalVerdict == "" {
			// slog.Println("ERROR FROM DATABASE: ", err)
			verdictAndRating.Rating = string(model.Critical)
			verdictAndRating.FinalVerdict = extras.BLOCK
			verdictAndRating.Score = 7
		}

		fod.Rating = verdictAndRating.Rating
		fod.FinalVerdict = verdictAndRating.FinalVerdict
		fod.Score = verdictAndRating.Score

		queryString = fmt.Sprintf("INSERT INTO %s (id, file_name, content_type, submitted_time, submitted_by, comments, file_count, from_device, status, finished_time, rating, final_verdict, md5, sha, sha256, score) VALUES (%d, '%s', '%s', '%s', '%s', '%s', %d, %t, '%s', '%s', '%s', '%s', '%s', '%s', '%s', %f)", extras.FileOnDemandTable, fod.Id, fod.FileName, fod.ContentType, fod.SubmittedTime.Format(extras.TIME_FORMAT), fod.SubmittedBy, fod.Comments, fod.FileCount, fod.FromDevice, fod.Status, fod.FinishedTime.Time.Format(extras.TIME_FORMAT), fod.Rating, fod.FinalVerdict, fod.Md5, fod.SHA, fod.SHA256, fod.Score)
		if fod.FromDevice {
			fod.ClientIp = ip
			queryString = fmt.Sprintf("INSERT INTO %s (id, file_name, content_type, submitted_time, submitted_by, comments, client_ip, file_count, from_device, status, finished_time, rating, final_verdict, md5, sha, sha256, score) VALUES (%d, '%s', '%s', '%s', '%s', '%s', '%s', %d, %t, '%s', '%s', '%s', '%s', '%s', '%s', '%s', %f)", extras.FileOnDemandTable, fod.Id, fod.FileName, fod.ContentType, fod.SubmittedTime.Format(extras.TIME_FORMAT), fod.SubmittedBy, fod.Comments, fod.ClientIp, fod.FileCount, fod.FromDevice, fod.Status, fod.FinishedTime.Time.Format(extras.TIME_FORMAT), fod.Rating, fod.FinalVerdict, fod.Md5, fod.SHA, fod.SHA256, fod.Score)
		}

		fodRepo = dao.DatabaseOperationsRepo{
			QueryExecSet: []string{queryString},
		}

		err = dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
		if err != nil {
			// slog.Println("ERROR WHILE INSERTING FOD: ", err)
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
		}
		return model.NewSuccessResponse(extras.ERR_SUCCESS, respMes)
	}

	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FILE_NOT_FOUND, extras.ErrFileNotFound)
}

func deleteLocalTask(id int) {
	fp := extras.SANDBOX_FILE_PATHS + fmt.Sprintf("%d", id)

	err := os.Remove(fp)
	if err != nil {
		// slog.Println("ERROR WHILE DELETING TASK FILE: ", id, err)
	}
}

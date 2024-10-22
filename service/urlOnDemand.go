package service

import (
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"anti-apt-backend/model"
	"database/sql"
	"encoding/csv"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func CreateUrlOnDemand(urlRequest model.UrlOnDemand, adminName, organizationKey string) model.APIResponse {
	var err error
	var resp model.APIResponse
	respMes := "Url: " + urlRequest.UrlName + " successfully uploaded"

	if len(urlRequest.UrlName) == 0 {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_REQUIRED_FIELD_EMPTY, extras.ErrRequiredFieldEmpty)
	}

	urlBody := urlRequest.UrlName
	// Parse Valid URL only else it will give error
	parsedURL, err := url.Parse(urlBody)
	if err != nil {
		// slog.Println("Error parsing URL:", err)
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_WHILE_PARSING_URL, err)
	}

	// if parseed successfully check for schema and host
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		// slog.Println(urlBody + " is not a valid URL.")
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_URL_FOUND, extras.ErrInvalidUrlFound)
	}

	var fromDevice bool
	if adminName == "DEVICE" && organizationKey == "" {
		fromDevice = true
	}

	// var platForm = "Windows 7"
	// platFormInBytes, _ := os.ReadFile(extras.PLATFORM_FILE_NAME)
	// if strings.Contains(strings.ToLower(string(platFormInBytes)), "ubuntu") {
	// 	platForm = "Ubuntu 20"
	// }

	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(3) + 1

	uod := model.UrlOnDemand{
		UrlName:       urlRequest.UrlName,
		SubmittedTime: time.Now(),
		FinishedTime:  sql.NullTime{Time: time.Now().Add(time.Duration(randomNumber) * time.Second), Valid: true},
		SubmittedBy:   adminName,
		Comments:      urlRequest.Comments,
		Status:        extras.REPORTED,
		// OsSupported:   platForm,
		UrlCount:   1,
		FromDevice: fromDevice,
	}

	rand.Seed(time.Now().UnixNano())
	randomNumber = rand.Intn(5) + 1

	ratings := []string{
		string(model.Critical),
		string(model.HighRisk),
		string(model.MediumRisk),
		string(model.LowRisk),
		string(model.Clean),
	}

	// found in trusted_urls
	malicipusUrls := FetchUrlsFromFile(extras.TEMP_MALICIOUS_URLS_FILE)
	if maliciousUrlContains(malicipusUrls, uod.UrlName) {
		uod.Rating = ratings[randomNumber-1]
		uod.FinalVerdict = extras.BLOCK
	} else {
		uod.Rating = string(model.Clean)
		uod.FinalVerdict = extras.ALLOW
	}

	queryString := fmt.Sprintf("INSERT INTO url_on_demands (url_name, submitted_time, finished_time, submitted_by, comments, status, url_count, from_device, rating, final_verdict) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', %d, %t, '%s', '%s')", uod.UrlName, uod.SubmittedTime.Format(extras.TIME_FORMAT), uod.FinishedTime.Time.Format(extras.TIME_FORMAT), uod.SubmittedBy, uod.Comments, uod.Status, uod.UrlCount, uod.FromDevice, uod.Rating, uod.FinalVerdict)
	uodRepo := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
	}
	err = dao.GormOperations(&uodRepo, config.Db, dao.EXEC)
	if err != nil {
		// slog.Println("ERROR IN SAVING TASK: ", err)
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
		return resp
	}

	if fromDevice {
		respMes = fmt.Sprintf("%d", uod.Id)
	}

	logger.LoggerFunc("info", logger.LoggerMessage("taskLog:Final report generated for "+uod.UrlName))

	return model.NewSuccessResponse(extras.ERR_SUCCESS, respMes)
}

func maliciousUrlContains(maliciousUrls []string, url string) bool {
	for _, maliciousUrl := range maliciousUrls {
		if strings.Contains(url, maliciousUrl) {
			return true
		}
	}
	return false
}

func FetchUrlsFromFile(urlFilePath string) []string {
	file, err := os.Open(urlFilePath)
	if err != nil {
		// slog.Println("error in opening malicious urls file: ", err)
		return []string{}
	}
	defer file.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(file)
	fields, _ := csvReader.ReadAll()

	var maliciousUrls []string
	for _, field := range fields {
		maliciousUrls = append(maliciousUrls, field[0])
	}

	return maliciousUrls
}

package cuckoo

import (
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"anti-apt-backend/model"
	"anti-apt-backend/validation"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var (
	BASE_URL = "http://127.0.0.1:1337"
)

func (c *Client) CreateTaskFile(ctx context.Context, id int, fp string) (sandboxId int, err error) {

	URL := fmt.Sprintf("%s/tasks/create/file", BASE_URL)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fp = strings.TrimSpace(fp)

	err = validation.ValidateFile([]string{fp})
	if err != nil {
		// slog.Println("ERROR WHILE VALIDATING FILE: ", err)
		return -1, err
	}

	// slog.Println("VALIDATED FILE IN CREATE TASK FILE")

	file, err := os.Open(fp)
	if err != nil {
		// slog.Println("ERROR WHILE OPENING FILE: ", err)
		return -1, err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(fp))
	if err != nil {
		// slog.Println("ERROR WHILE CREATING FORM FILE: ", err)
		return -1, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		// slog.Println("ERROR WHILE COPYING FILE: ", err)
		return -1, err
	}

	err = writer.Close()
	if err != nil {
		// slog.Println("ERROR WHILE CLOSING WRITER: ", err)
		return -1, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", URL, body)
	if err != nil {
		// slog.Println("ERROR WHILE CREATING REQUEST: ", err)
		return -1, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	var taskResp model.CreateTaskFileResponse
	httpResp, err := c.MakeRequest(req)
	if err != nil {
		// slog.Println("ERROR WHILE MAKING REQUEST: ", err)
		return -1, err
	}

	// slog.Println("PRINTING HTTP RESPONSE IN CREATE FILE TASK", httpResp)

	if httpResp != nil {
		if httpResp.StatusCode != http.StatusOK {
			// slog.Println("BAD RESPONSE CODE: ", httpResp.StatusCode)
			return -1, fmt.Errorf("bad response code: %d", httpResp.StatusCode)
		}

		err = json.NewDecoder(httpResp.Body).Decode(&taskResp)
		if err != nil {
			// slog.Println("ERROR WHILE DECODING RESPONSE: ", err)
			return -1, err
		}

		if taskResp.TaskId == 0 {
			// slog.Println("TASK ID IS 0")
			return 0, fmt.Errorf("task id is 0")
		}

		// slog.Println("TASK ID GENERATED: ", taskResp.TaskId)

	} else {
		// slog.Println("HTTP RESPONSE IS NIL")
		return -1, fmt.Errorf("httpResp is nil")
	}

	return taskResp.TaskId, nil
}

func (c *Client) CreateTaskUrl(ctx context.Context, id int, urlToSubmit string) (sandboxId int, err error) {

	URL := fmt.Sprintf("%s/tasks/create/url", BASE_URL)

	urlToSubmit = strings.TrimSpace(urlToSubmit)

	body := url.Values{}
	body.Set("url", urlToSubmit)

	req, err := http.NewRequestWithContext(ctx, "POST", URL, strings.NewReader(body.Encode()))
	if err != nil {
		// slog.Println("ERROR WHILE CREATING REQUEST: ", err)
		return -1, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var taskResp model.CreateTaskFileResponse
	httpResp, err := c.MakeRequest(req)
	if err != nil {
		// slog.Println("ERROR WHILE MAKING REQUEST: ", err)
		return -1, err
	}

	if httpResp != nil {

		if httpResp.StatusCode != http.StatusOK {
			// slog.Println("BAD RESPONSE CODE: ", httpResp.StatusCode)
			return -1, fmt.Errorf("bad response code: %d", httpResp.StatusCode)
		}

		err = json.NewDecoder(httpResp.Body).Decode(&taskResp)
		if err != nil {
			// slog.Println("ERROR WHILE DECODING RESPONSE: ", err)
			return -1, err
		}

		if taskResp.TaskId == 0 {
			// slog.Println("TASK ID IS 0")
			return 0, fmt.Errorf("task id is 0")
		}

		// slog.Println("TASK ID GENERATED: ", taskResp.TaskId)

	} else {
		// slog.Println("HTTP RESPONSE IS NIL FOR TASK Id : ", id)
		return -1, fmt.Errorf("httpResp is nil")
	}

	return taskResp.TaskId, nil
}

func (c *Client) TasksView(ctx context.Context, taskID int) (*model.Task, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/tasks/view/%d", BASE_URL, taskID), nil)
	if err != nil {
		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:couldn't create request"))
		return nil, err
	}

	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}

	// log.Println("resp------------------>", resp)
	// logger.LoggerFunc("error", logger.LoggerMessage(resp))

	switch resp.StatusCode {
	case 404:
		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:"+extras.ErrTaskNotFound.Error()))
		return nil, extras.ErrTaskNotFound
	case 200:
		break
	default:
		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:bad response"))
		return nil, fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	task := struct {
		Task model.Task `json:"task"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:couldn't decoding task"))
		return nil, err
	}

	logger.LoggerFunc("info", logger.LoggerMessage("taskLog:Final task generated"))
	return &task.Task, nil
}

func (c *Client) TasksReport(ctx context.Context, taskID int) (*model.Report, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/tasks/report/%d", BASE_URL, taskID), nil)
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:couldn't create request"))
		return nil, err
	}

	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}

	// log.Println("resp------------------>", resp, "taskID:", taskID)
	// logger.LoggerFunc("error", logger.LoggerMessage(resp))

	var report model.Report

	err = json.NewDecoder(resp.Body).Decode(&report)
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:couldn't decode report"))
		return nil, err
	}

	switch resp.StatusCode {
	case 404:
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:"+extras.ErrReportNotFound.Error()))
		return nil, extras.ErrReportNotFound
	case 400:
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:invalid report format"))
		return nil, fmt.Errorf("invalid report format")
	case 200:
		break
	default:
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:bad response code"))
		return nil, fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	queryString := fmt.Sprintf("SELECT * FROM %s WHERE id = '%s'", extras.FileOnDemandTable, report.Target.File.Name)

	var fod model.FileOnDemand
	fodRepo := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &fod,
	}

	err = dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
	if err != nil {
		return nil, err
	}

	// slog.Println("report file name: ", report.Target.File.Name)
	// slog.Println("file name from DB: ", fod.FileName)

	logger.LoggerFunc("info", logger.LoggerMessage("taskLog:Final report generated for "+fod.FileName))
	return &report, nil
}

func (c *Client) ListMachines(ctx context.Context) ([]*model.Machine, error) {

	c.BaseURL = "http://127.0.0.1:1337"

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/machines/list", BASE_URL), nil)
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:couldn't create request"))
		// fmt.Println("error in creating request: ", err)
		return nil, err
	}

	resp, err := c.MakeRequest(req)
	if err != nil {
		// fmt.Println("error in making request: ", err)
		return nil, err
	}
	switch resp.StatusCode {
	case 200:
		break
	default:
		// fmt.Println("bad response code: ", resp.StatusCode)
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:bad response"))
		return nil, fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	machines := struct {
		Machines []*model.Machine `json:"machines"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&machines)
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:couldn't decode machines"))
		// fmt.Println("error in decoding response: ", err)
		return nil, fmt.Errorf("response marshalling error: %w", err)
	}

	return machines.Machines, nil
}

func (c *Client) ViewMachine(ctx context.Context, machineName string) (*model.Machine, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/machines/view/%s", BASE_URL, machineName), nil)
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:couldn't create request"))
		return nil, err
	}

	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case 200:
		break
	case 404:
		return nil, extras.ErrMachineNotfound
	default:
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:bad response"))
		return nil, fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	machine := struct {
		Machine *model.Machine `json:"machine"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&machine)
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:couldn't decode machines"))
		return nil, fmt.Errorf("response marshalling error: %w", err)
	}

	return machine.Machine, nil
}

func (c *Client) TasksDelete(ctx context.Context, taskID int) (err error) {

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/tasks/delete/%d", BASE_URL, taskID), nil)
	if err != nil {
		return err
	}

	resp, err := c.MakeRequest(req)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case 404:
		return extras.ErrTaskNotFound
	case 500:
		body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		if err != nil {
			body = []byte{}
		}
		return fmt.Errorf("unable to delete the task, body: %s", body)
	case 200:
		break
	default:
		return fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) ListAllTasks(ctx context.Context) ([]*model.Task, error) {

	URL := fmt.Sprintf("%s/tasks/list", BASE_URL)
	req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad response code: %d", resp.StatusCode)
	}

	tasks := struct {
		Tasks []*model.Task `json:"tasks"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return nil, err
	}

	return tasks.Tasks, nil
}

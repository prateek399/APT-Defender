package queues

import (
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service/cuckoo"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

type Sandbox struct {
	SandboxId   int
	Status      string
	CompletedOn time.Time
}

const (
	SANDBOX_API_TIMEOUT = 30 * time.Second
	SANDBOX_MAX_RETRIES = 5
)

// Add timeout for all cuckoo APIS
func fetchAllTasksFromSandbox() ([]Sandbox, error) {

	client := cuckoo.New(&cuckoo.Config{
		Client: &http.Client{Timeout: SANDBOX_API_TIMEOUT},
	})
	allTasks, err := client.ListAllTasks(context.Background())
	if err != nil {
		return nil, err
	}

	var resp []Sandbox
	var t time.Time
	for _, task := range allTasks {
		if task.CompletedOn != nil {
			t, _ = time.Parse(time.RFC1123, fmt.Sprintf("%v", task.CompletedOn))
		}
		resp = append(resp, Sandbox{
			SandboxId:   task.ID,
			Status:      string(task.Status),
			CompletedOn: t,
		})
	}

	return resp, nil
}

func countOfFreeVm() (int, error) {
	client := cuckoo.New(&cuckoo.Config{
		Client: &http.Client{Timeout: SANDBOX_API_TIMEOUT},
	})
	allMachines, err := client.ListMachines(context.Background())
	count := 0
	if err != nil {
		return count, err
	}
	for _, machine := range allMachines {
		if !machine.Locked {
			count++
		}
	}
	return count, nil
}

func sendToSandbox(task Task) (int, error) {

	fp := extras.SANDBOX_FILE_PATHS + fmt.Sprintf("%d", task.Id)

	client := cuckoo.New(&cuckoo.Config{})
	sandboxId, err := client.CreateTaskFile(context.Background(), task.Id, fp)
	if err != nil {
		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE CREATING TASK IN SANDBOX: %v", err))
		return -1, err
	}

	if sandboxId == 0 {
		// logger.LogAccToTaskId(task.Id, "SANDBOXID IS ZERO")
	}

	return sandboxId, nil
}

func fetchScoreFromSandBox(taskId int, sandboxId int) (float32, error) {

	client := cuckoo.New(&cuckoo.Config{
		Client: &http.Client{Timeout: SANDBOX_API_TIMEOUT},
	})
	report, err := client.TasksReport(context.Background(), sandboxId)
	if err != nil {
		// logger.LogAccToTaskId(taskId, fmt.Sprintf("ERROR WHILE FETCHING REPORT: %v", err))
		return 0, err
	}

	if report != nil {
		return report.Info.Score, nil
	}

	return 0, nil
}

func deleteSandboxData(taskId int, sandboxId int) error {

	if sandboxId == 0 {
		return nil
	}

	client := cuckoo.New(&cuckoo.Config{})
	err := client.TasksDelete(context.Background(), sandboxId)
	if err != nil {
		// logger.LogAccToTaskId(taskId, fmt.Sprintf("ERROR WHILE DELETING SANDBOX DATA: %v", err))
		return err
	} else {
		// logger.LogAccToTaskId(taskId, fmt.Sprintf("DELETED SANDBOX DATA: %d", sandboxId))
		// slog.Println("DELETED SANDBOX DATA: ", sandboxId)
	}

	return nil
}

type Report struct {
	Score       float32
	Completedon time.Time
}

func fetchReportInfoFromSandBox(sandboxId int) (Report, error) {

	client := cuckoo.New(&cuckoo.Config{})
	report, err := client.TasksReport(context.Background(), sandboxId)
	if err != nil {
		return Report{}, err
	}

	completedOn, _ := time.Parse(time.RFC1123, fmt.Sprintf("%v", report.Info.Ended))

	return Report{Score: report.Info.Score, Completedon: completedOn}, nil
}

func FlushSandboxData() error {
	allTasks, err := fetchAllTasksFromSandbox()
	if err != nil {
		return err
	}

	for _, task := range allTasks {
		err := deleteSandboxData(task.SandboxId, task.SandboxId)
		if err != nil {
			// // slog.Println("ERROR WHILE DELETING SANDBOX DATA FOR ID: ", task.SandboxId, err)
		}
	}
	return nil
}

func RemoveSandboxTasksWhichAreNotInTasksAnReachedTimeOut(allSandboxTasks []Sandbox, tasks []Task) error {

	taskMap := make(map[int]bool)

	for _, task := range tasks {
		if task.SandboxId > 0 {
			taskMap[task.SandboxId] = true
		}
	}

	var resp []Sandbox

	for _, sandboxTask := range allSandboxTasks {
		if _, ok := taskMap[sandboxTask.SandboxId]; !ok && time.Since(sandboxTask.CompletedOn) > 30*time.Minute {
			resp = append(resp, sandboxTask)
		}
	}

	for _, task := range resp {
		err := pushToQueue(LogQueue, Task{SandboxId: task.SandboxId}, SandboxApiTimeout)
		if err != nil {
			go deleteSandboxData(task.SandboxId, task.SandboxId)
		}
	}
	return nil
}

func FetchSandboxTaskCountWhichAreNotReported() int {
	allSandboxTasks, err := fetchAllTasksFromSandbox()
	if err != nil {
		// slog.Println("ERROR WHILE FETCHING ALL TASKS FROM SANDBOX: ", err)
	}

	count := 0

	for _, sandboxTask := range allSandboxTasks {
		if sandboxTask.Status != extras.Reported && sandboxTask.Status != "failed_analysis" {
			count++
		}
	}
	return count
}

const (
	ANALYSING = "ANALYSING"
)

func SendAcknowledgementToClientIp(taskId int) {

	// slog.Println("SENDING ACKNOWLEDGEMENT TO CLIENT IP: ", taskId)

	ip, verdict := sendAcknowledgement(taskId)

	if ip == "" || verdict == "" {
		return
	}

	URL := "https://" + ip + ":8085/verdict"

	payload := fmt.Sprintf(`{"verdict": "%s", "taskID": %d}`, verdict, taskId)
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		// slog.Println("ERROR WHILE CREATING REQUEST: ", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	transport := http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := http.Client{
		Timeout:   10 * time.Second,
		Transport: &transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		// slog.Println("ERROR WHILE MAKING REQUEST: ", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// slog.Printf("ERROR IN RESPONSE CODE: %d", resp.StatusCode)
		return
	}

	// slog.Println("RESPONSE: ", resp)
}

func sendAcknowledgement(taskId int) (string, string) {
	queryString := fmt.Sprintf("SELECT * FROM %s WHERE id = %d", extras.FileOnDemandTable, taskId)

	var fod model.FileOnDemand
	fodRepo := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &fod,
	}

	err := dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
	if err != nil {
		return fod.ClientIp, fod.FinalVerdict
	}

	if fod.FinalVerdict == "" {
		return fod.ClientIp, ANALYSING // Not sure what to do here
	} else {
		return fod.ClientIp, fod.FinalVerdict
	}
}

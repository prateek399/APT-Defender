package queues

import (
	"anti-apt-backend/extras"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type SandboxTest struct {
	SandboxId   int
	Status      string
	AddedOn     time.Time
	CompletedOn time.Time
	Score       float32
}

type Machine struct {
	Name   string
	Locked bool
}

type MockCuckooClient struct {
	mu        sync.Mutex
	tasks     map[int]*SandboxTest
	nextID    int
	createErr error
	deleteErr error
	reportErr error
}

var (
	mockClientInstance *MockCuckooClient
	once               sync.Once
)

func NewMockCuckooClient() *MockCuckooClient {
	once.Do(func() {
		mockClientInstance = &MockCuckooClient{
			tasks:  make(map[int]*SandboxTest),
			nextID: 1,
		}
		go mockClientInstance.updateTaskStatuses()
	})
	return mockClientInstance
}

func (m *MockCuckooClient) updateTaskStatuses() {
	for {
		time.Sleep(1 * time.Second)
		m.mu.Lock()
		now := time.Now()
		for _, task := range m.tasks {
			switch task.Status {
			case "pending":
				if now.Sub(task.AddedOn) > 2*time.Second {
					task.Status = "running"
				}
			case "running":
				if now.Sub(task.AddedOn) > 4*time.Second {
					task.Status = "completed"
					task.CompletedOn = now
				}
			case "completed":
				if now.Sub(task.AddedOn) > 6*time.Second {
					task.Status = "reported"
					task.Score = rand.Float32() * 10
				}
			}
		}
		m.mu.Unlock()
	}
}

func (m *MockCuckooClient) CreateTaskFileMock(ctx context.Context, taskId int, filePath string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createErr != nil {
		return 0, m.createErr
	}
	id := m.nextID
	m.nextID++
	m.tasks[id] = &SandboxTest{
		SandboxId: id,
		Status:    "pending",
		AddedOn:   time.Now(),
	}
	return id, nil
}

func (m *MockCuckooClient) ListAllTasksMock(ctx context.Context) ([]SandboxTest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var tasks []SandboxTest
	for _, task := range m.tasks {
		tasks = append(tasks, *task)
	}
	// slog.Println("LIST ALL TASKS: ", tasks)
	return tasks, nil
}

func (m *MockCuckooClient) ListMachinesMock(ctx context.Context) ([]Machine, error) {
	var machines []Machine
	for i := 0; i < 8; i++ {
		machines = append(machines, Machine{Locked: rand.Intn(2) == 0, Name: fmt.Sprintf("machine-%d", i+1)})
	}
	return machines, nil
}

func (m *MockCuckooClient) TasksReportMock(ctx context.Context, sandboxId int) (*Report, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.reportErr != nil {
		return nil, m.reportErr
	}
	task, ok := m.tasks[sandboxId]
	if !ok {
		return nil, nil
	}
	return &Report{
		Score:       task.Score,
		Completedon: task.CompletedOn,
	}, nil
}

func (m *MockCuckooClient) TasksDeleteMock(ctx context.Context, sandboxId int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.tasks, sandboxId)
	return nil
}

func fetchAllTasksFromSandboxMock() ([]Sandbox, error) {
	client := NewMockCuckooClient()
	allTasks, err := client.ListAllTasksMock(context.Background())
	if err != nil {
		return nil, err
	}

	var resp []Sandbox
	for _, task := range allTasks {
		resp = append(resp, Sandbox{
			SandboxId:   task.SandboxId,
			Status:      task.Status,
			CompletedOn: task.CompletedOn,
		})
	}

	// slog.Println("ALL SANDBOX TASKS: ", resp)
	return resp, nil
}

func countOfFreeVmMock() (int, error) {
	client := NewMockCuckooClient()
	allMachines, err := client.ListMachinesMock(context.Background())
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

func sendToSandboxMock(task Task) (int, error) {
	client := NewMockCuckooClient()
	fp := extras.SANDBOX_FILE_PATHS + fmt.Sprintf("%d", task.Id)
	sandboxId, err := client.CreateTaskFileMock(context.Background(), task.Id, fp)
	if err != nil {
		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE CREATING TASK IN SANDBOX: %v", err))
		return -1, err
	}

	if sandboxId == 0 {
		// logger.LogAccToTaskId(task.Id, "SANDBOXID IS ZERO")
	}

	return sandboxId, nil
}

func fetchScoreFromSandboxMock(taskId int, sandboxId int) (float32, error) {
	client := NewMockCuckooClient()
	report, err := client.TasksReportMock(context.Background(), sandboxId)
	if err != nil {
		// logger.LogAccToTaskId(taskId, fmt.Sprintf("ERROR WHILE FETCHING REPORT: %v", err))
		return 0, err
	}

	if report != nil {
		return report.Score, nil
	}

	return 0, nil
}

func deleteSandboxDataMock(taskId int, sandboxId int) error {
	if sandboxId == 0 {
		return nil
	}

	client := NewMockCuckooClient()
	err := client.TasksDeleteMock(context.Background(), sandboxId)
	if err != nil {
		// logger.LogAccToTaskId(taskId, fmt.Sprintf("ERROR WHILE DELETING SANDBOX DATA: %v", err))
		return err
	} else {
		// logger.LogAccToTaskId(taskId, fmt.Sprintf("DELETED SANDBOX DATA: %d", sandboxId))
		// slog.Println("DELETED SANDBOX DATA: ", sandboxId)
	}

	return nil
}

func FlushSandboxDataMock() error {
	allTasks, err := fetchAllTasksFromSandbox()
	if err != nil {
		return err
	}

	for _, task := range allTasks {
		err := deleteSandboxData(task.SandboxId, task.SandboxId)
		if err != nil {
			// slog.Println("ERROR WHILE DELETING SANDBOX DATA FOR ID: ", task.SandboxId, err)
		}
	}
	return nil
}

func RemoveSandboxTasksWhichAreNotInTasksAndReachedTimeOutMock(allSandboxTasks []Sandbox, tasks []Task) error {
	taskMap := make(map[int]bool)
	for _, task := range tasks {
		if task.SandboxId > 0 {
			taskMap[task.SandboxId] = true
		}
	}

	for _, sandboxTask := range allSandboxTasks {
		if _, ok := taskMap[sandboxTask.SandboxId]; !ok && !sandboxTask.CompletedOn.IsZero() && time.Since(sandboxTask.CompletedOn) > 30*time.Minute {
			// logger.LogAccToTaskId(sandboxTask.SandboxId, fmt.Sprintf("DELETING SANDBOX DATA FOR TASK: %v", sandboxTask))
			// slog.Println("DELETING SANDBOX DATA FOR TASK: ", sandboxTask)
			err := pushToQueue(LogQueue, Task{SandboxId: sandboxTask.SandboxId}, SandboxApiTimeout)
			if err != nil {
				// logger.LogAccToTaskId(sandboxTask.SandboxId, fmt.Sprintf("DELETING SANDBOX DATA FOR TASK: %v", sandboxTask))
				// slog.Println("DELETING SANDBOX DATA FOR TASK: ", sandboxTask)
				go deleteSandboxData(sandboxTask.SandboxId, sandboxTask.SandboxId)
			}
		}
	}

	return nil
}

func SendAcknowledgementToClientIpMock(taskId int) {

	// slog.Println("SENDING ACKNOWLEDGEMENT TO CLIENT IP: ", taskId)

	ip, verdict := sendAcknowledgement(taskId)

	if ip == "" || verdict == "" {
		return
	}

	URL := "https://192.168.199.163:8082/api/v1/checkhash"

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
		// slog.Printf("ERROR IN RESPONSE CODE: %d", resp.Body)
		return
	}

	// slog.Println("RESPONSE: ", resp)
}

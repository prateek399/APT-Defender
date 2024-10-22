package queues

import (
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gookit/slog"
	"golang.org/x/sync/semaphore"
)

const (
	PendingNotInQueue         = "pending not in queue"
	PendingInQueue            = "pending in queue"
	Pending                   = "pending"
	RunningNotInQueue         = "running not in queue"
	RunningInQueue            = "running in queue"
	Running                   = "running" // in log_queue
	Reported                  = "reported"
	ReportedThroughQuickScope = "reported through quickscope"
	ReportedThroughClamd      = "reported through clamd"
	Aborted                   = "aborted"
	Completed                 = "completed"
	SandboxApiTimeout         = "sandbox api timeout"
)

type Task struct {
	Id                int
	Status            string
	SandboxId         int
	QueueRetryCount   int
	RunningRetryCount int
	SandboxRetryCount int
	Md5               string
	SHA               string
	SHA256            string
	LogQueueFailed    bool
	SubmittedTime     time.Time
	RunningStartedAt  time.Time
}

type FinishedTask struct {
	Id        int
	SandboxId int
	Aborted   bool
}

var (
	PendingQueue = make(chan Task, 5000)
	RunningQueue = make(chan Task, 5000)
	LogQueue     = make(chan Task, 5000)
)

type DuplicateTask struct {
	Id     int
	Md5    string
	SHA    string
	SHA256 string
}

const (
	QUEUE_MAX_RETRIES   = 3
	RUNNING_MAX_RETRIES = 5
	MAX_FREE_VMS        = 5
)

var (
	MAX_SANDBOX_TASKS     = 300
	TIMEOUT               = 15
	PENDING_QUEUE_TIMEOUT = 15 * time.Minute
	SANDBOX_TIME_OUT      = 15 * time.Minute
)

// push to respective queue
func pushToQueue(queue chan Task, task Task, status string) error {
	task.Status = status
	select {
	case queue <- task:
		return nil
	default:
		// // slog.Println("default case while pusing into queue : ", task.Id)
		return fmt.Errorf("failed to push into queue")
	}
}

func fetchTasksFromQueue(queue chan Task, count int) []Task {
	var tasks []Task

	if count == -1 {
		for {
			select {
			case task := <-queue:
				tasks = append(tasks, task)
			default:
				return tasks
			}
		}
	}

	for i := 0; i < count && len(queue) > 0; i++ {
		tasks = append(tasks, <-queue)
	}
	return tasks
}

func getMaxSandboxTasks() int {
	content, err := os.ReadFile(extras.MAX_SANDBOX_TASKS_FILE_PATH)
	if err != nil {
		return MAX_SANDBOX_TASKS
	}

	valueStr := strings.TrimSpace(string(content))
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return MAX_SANDBOX_TASKS
	}
	return value
}

func getTimeOut() int {
	content, err := os.ReadFile(extras.TIMEOUT_FILE_PATH)
	if err != nil {
		return TIMEOUT
	}

	valueStr := strings.TrimSpace(string(content))
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return TIMEOUT
	}
	return value
}

// Step 3
func NotInQueueHandler() {

	// ignoreExtensions, _ := extensionsToIgnore()

	var targetQueue chan Task
	var newStatus string
	for {
		// logger.LogAccToTaskId(0, "GO ROUTINES: "+fmt.Sprintf("%d", runtime.NumGoroutine()))

		tasks, err := fetchNotInQueueTasks()
		if err != nil {
			// // slog.Println("ERROR WHILE FETCHING NOT IN QUEUE TASKS: ", err)
			continue
		}

		// // slog.Println("NOT IN QUEUE TASKS: ", tasks)

		for _, task := range tasks {

			if task.LogQueueFailed {
				// logger.LogAccToTaskId(task.Id, fmt.Sprintf("LOG QUEUE FAILED : %v", task))
				targetQueue = LogQueue
				newStatus = task.Status
				if strings.Contains(task.Status, PendingInQueue) {
					newStatus = PendingNotInQueue
				} else if strings.Contains(task.Status, RunningInQueue) {
					newStatus = RunningNotInQueue
				} else {
					// logger.LogAccToTaskId(task.Id, fmt.Sprintf("UNKNOWN TASK STATUS: %v", task.Status))
				}
			} else if task.Status == PendingNotInQueue {

				var foundMalicious = false

				maliciousFromClamd, err := ScanFileThroughClamd(task)
				if maliciousFromClamd {
					foundMalicious = true
					targetQueue = LogQueue
					newStatus = ReportedThroughClamd
				} else if err != nil || !maliciousFromClamd {
					maliciousFromQuickScope, _ := ScanFileThroughQuickScope(task)
					if maliciousFromQuickScope {
						foundMalicious = true
						targetQueue = LogQueue
						newStatus = ReportedThroughQuickScope
					}
				}

				if !foundMalicious {
					liveTaskCount, _ := fetchLiveTaskCount()
					MAX_SANDBOX_TASKS = getMaxSandboxTasks()
					if liveTaskCount >= MAX_FREE_VMS {
						targetQueue = LogQueue
						newStatus = Aborted
					} else {
						targetQueue = PendingQueue
						newStatus = PendingInQueue
					}
				}

				// extension := strings.ToLower(filepath.Ext(filePath))
				// if _, found := ignoreExtensions[extension]; found {
				// 	targetQueue = PendingQueue
				// 	newStatus = PendingInQueue
				// }
			} else if task.Status == RunningNotInQueue {
				// if this case repeasts every time, it will lead to deadlock as it will replace time with current time
				task.RunningStartedAt = time.Now()
				targetQueue = RunningQueue
				newStatus = RunningInQueue
			} else {
				// // slog.Println("UNKNOWN TASK STATUS: ", task.Status, task.Id)
				continue
			}
			err := pushToQueue(targetQueue, task, newStatus)
			if err != nil {
				// logger.LogAccToTaskId(task.Id, fmt.Sprintf("QUEUE COUNT %d, RUNNING COUNT %d, SANDBOX COUNT %d, ERROR WHILE PUSHING TASK INTO QUEUE: %v", task.QueueRetryCount, task.RunningRetryCount, task.SandboxRetryCount, err))
				// slog.Println("ERROR WHILE PUSHING TASK INTO QUEUE: ", task.Id)
			} else {
				// direct update through DB, rest should all be through loq queue
				// update log_queue_failed to 0, so that it will not be retried
				if targetQueue == LogQueue {
					err = updateLogQueueFail(task.Id, false)
					if err != nil {
						// logger.LogAccToTaskId(task.Id, "Failed to set log_queue_failed for task "+err.Error())
						// slog.Println("Failed to set log_queue_failed for task %d : %v", task.Id, err)
					}
				} else {
					err := updateTaskStatus(task.Id, newStatus)
					if err != nil {
						// logger.LogAccToTaskId(task.Id, "Failed to update task status "+err.Error())
						// slog.Println("Failed to update task %d status to %s: %v", task.Id, newStatus, err)
					}
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

var pendingSem = semaphore.NewWeighted(int64(MAX_CONCURRENT_LOG_QUEUE_TASKS))

func PendingQueueHandler() {
	for {
		// freeVms, err := countOfFreeVm()
		// if err != nil {
		// 	// // slog.Println("ERROR WHILE FETCHING FREE VM'S: ", err)
		// 	time.Sleep(2 * time.Second)
		// 	continue
		// }

		freeVms := MAX_FREE_VMS

		notReported := FetchSandboxTaskCountWhichAreNotReported()

		if notReported >= freeVms {
			time.Sleep(2 * time.Second)
			continue
		}

		freeVms = freeVms - notReported
		// slog.Println("FREE VM'S: ", freeVms)

		if freeVms <= 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		tasks := fetchTasksFromQueue(PendingQueue, freeVms)

		if len(tasks) == 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		// // slog.Println("PENDING QUEUE TASKS: ", tasks)
		// var wg sync.WaitGroup
		PENDING_QUEUE_TIMEOUT = time.Duration(getTimeOut()) * time.Minute
		for _, task := range tasks {
			if task.QueueRetryCount >= QUEUE_MAX_RETRIES || task.SubmittedTime.Add(PENDING_QUEUE_TIMEOUT).Before(time.Now()) {
				// logger.LogAccToTaskId(task.Id, fmt.Sprintf("PENDING QUEUE TIMEOUT: %v", task.SubmittedTime))
				changeStatusAndPushToLogQueue(task, Aborted)
				// no need for further processing, decreased wg count by 1 so that deadlock will not happen
				continue
			}

			// added tasks in wg to avoid deadlock, as tasks can be lesser than freevms
			// wg.Add(1)

			if err := pendingSem.Acquire(context.Background(), 1); err != nil {
				// slog.Println("Failed to acquire semaphore:", err)
				// wg.Done()
				continue
			}

			go func(task Task) {
				// defer wg.Done()
				defer pendingSem.Release(1)
				sandboxId, err := sendToSandbox(task)
				// logger.LogAccToTaskId(task.Id, fmt.Sprintf("SANDBOX ID IN PENDING QUEUE: %d", sandboxId))
				// slog.Println("SANDBOX ID IN PENDING QUEUE: ", sandboxId)
				if err != nil || sandboxId <= 0 {
					task.SandboxRetryCount++
					if task.SandboxRetryCount >= SANDBOX_MAX_RETRIES {
						// logger.LogAccToTaskId(task.Id, fmt.Sprintf("Sandbox retry count exceeded for task: %d", task.Id))
						// slog.Println("Sandbox retry count exceeded for task: ", task.Id)
						changeStatusAndPushToLogQueue(task, Aborted)
					} else {
						err := pushToQueue(PendingQueue, task, task.Status)
						if err != nil {
							// logger.LogAccToTaskId(task.Id, fmt.Sprintf("QUEUE COUNT %d, RUNNING COUNT %d, SANDBOX COUNT %d, ERROR WHILE PUSHING TASK INTO QUEUE: %v", task.QueueRetryCount, task.RunningRetryCount, task.SandboxRetryCount, err))
							// slog.Println("ERROR WHILE PUSHING TASK INTO PENDING QUEUE: ", err)
							task.QueueRetryCount++
							if task.QueueRetryCount >= QUEUE_MAX_RETRIES {
								// logger.LogAccToTaskId(task.Id, fmt.Sprintf("Queue retry count exceeded for task: %d", task.Id))
								// slog.Println("Queue retry count exceeded for task: ", task.Id)
								changeStatusAndPushToLogQueue(task, Aborted)
							} else {
								// push back into pending queue instead of log queue, same task turn may get longer wait (PG)
								err := pushToQueue(LogQueue, task, PendingNotInQueue)
								if err != nil {
									// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE PUSHING TASK INTO LOG QUEUE: %v", err))
									// slog.Println("ERROR WHILE PUSHING TASK INTO LOG QUEUE: ", err)
									task.Status = PendingNotInQueue
									task.LogQueueFailed = true
									err = updateLiveTaskDirectlyThroughDb(task)
									if err != nil {
										// logger.LogAccToTaskId(task.Id, fmt.Sprintf("Failed to update loq queue failed task %d status to %s: %v", task.Id, PendingNotInQueue, err))
										// slog.Println("Failed to update loq queue failed task %d status to %s: %v", task.Id, PendingNotInQueue, err)
									}
								} else {
									// logger.LogAccToTaskId(task.Id, fmt.Sprintf("QUEUE RETRY COUNT %d ", task.QueueRetryCount))
									// logger.LogAccToTaskId(task.Id, fmt.Sprintf("PUSHED TASK INTO LOG QUEUE: %s", task.Status))
									// slog.Println("PUSHED TASK INTO LOG QUEUE: ", task.Id, task.Status)
								}
							}

						} else {
							// logger.LogAccToTaskId(task.Id, fmt.Sprintf("PUSHED TASK BACK INTO PENDING QUEUE: %s", task.Status))
							// slog.Println("PUSHED TASK BACK INTO PENDING QUEUE: ", task.Id, task.Status)
						}
					}
				} else {
					task.SandboxId = sandboxId
					task.RunningStartedAt = time.Now()
					// logger.LogAccToTaskId(task.Id, fmt.Sprintf("TASKID %d- SANDBOXID %d- RUNNING_STARTED_AT %v : ", task.Id, task.SandboxId, task.RunningStartedAt))
					// slog.Println("TASKID %d- SANDBOXID %d- RUNNING_STARTED_AT %v : ", task.Id, task.SandboxId, task.RunningStartedAt)
					err := pushToQueue(RunningQueue, task, task.Status)
					if err != nil {
						// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE PUSHING TASK INTO RUNNING QUEUE: %v", err))
						changeStatusAndPushToLogQueue(task, RunningNotInQueue)
					} else {
						// logger.LogAccToTaskId(task.Id, fmt.Sprintf("PUSHED TASK INTO RUNNING QUEUE: %s", task.Status))
						// slog.Println("PUSHED TASK INTO RUNNING QUEUE: ", task.Id, task.Status)
						changeStatusAndPushToLogQueue(task, RunningInQueue)
					}
				}
			}(task)
		}
		// wg.Wait()
		// time.Sleep(2 * time.Second)
	}
}

func RunningQueueHandler() {
	for {

		allSandboxTasks, err := fetchAllTasksFromSandbox()
		if err != nil {
			// slog.Println("ERROR WHILE FETCHING ALL TASKS FROM SANDBOX: ", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var sandboxTaskMap = make(map[int]string)

		for _, sandboxTask := range allSandboxTasks {
			sandboxTaskMap[sandboxTask.SandboxId] = sandboxTask.Status
		}

		tasks := fetchTasksFromQueue(RunningQueue, -1)

		// RemoveSandboxTasksWhichAreNotInTasksAnReachedTimeOut(allSandboxTasks, tasks)

		if len(tasks) > 0 {
			// slog.Println("RUNNING QUEUE TASKS: ", tasks)
		}

		SANDBOX_TIME_OUT = time.Duration(getTimeOut()) * time.Minute

		for _, task := range tasks {
			if task.RunningRetryCount >= RUNNING_MAX_RETRIES {
				changeStatusAndPushToLogQueue(task, Aborted)
			} else {
				sandboxStatus, ok := sandboxTaskMap[task.SandboxId]
				if !ok {
					// logger.LogAccToTaskId(task.Id, fmt.Sprintf("SANDBOX NOT FOUND IN MAP: %d", task.SandboxId))
					// slog.Println("SANDBOX NOT FOUND IN MAP: ", task.SandboxId)
					task.RunningRetryCount++
					if task.RunningRetryCount >= RUNNING_MAX_RETRIES {
						// logger.LogAccToTaskId(task.Id, fmt.Sprintf("SANDBOX NOT FOUND IN MAP: %d", task.SandboxId))
						// slog.Println("SANDBOX NOT FOUND IN MAP: ", task.SandboxId)
						changeStatusAndPushToLogQueue(task, Aborted)
					}
					changeStatusAndPushToLogQueue(task, RunningNotInQueue)
				} else {
					// logger.LogAccToTaskId(task.Id, fmt.Sprintf("TASKID - SANDBOXID - STATUS : %d - %d - %s", task.Id, task.SandboxId, sandboxStatus))
					// slog.Println("TASKID - SANDBOXID - STATUS : ", task.Id, task.SandboxId, sandboxStatus)
					if sandboxStatus == Reported {
						changeStatusAndPushToLogQueue(task, Reported)
					} else if sandboxStatus == Running || sandboxStatus == Completed || sandboxStatus == Pending {
						if time.Since(task.RunningStartedAt) > SANDBOX_TIME_OUT {
							// logger.LogAccToTaskId(task.Id, fmt.Sprintf("SANDBOX TIMEOUT FOR : %d", task.SandboxId))
							// slog.Println("SANDBOX TIMEOUT FOR : ", task.SandboxId)
							changeStatusAndPushToLogQueue(task, Aborted)
						} else {
							// push to runningQueue
							err := pushToQueue(RunningQueue, task, task.Status)
							if err != nil {
								// logger.LogAccToTaskId(task.Id, fmt.Sprintf("PUSHING TASK BACK INTO RUNNING QUEUE: %v", err))
								// slog.Println("PUSHING TASK BACK INTO RUNNING QUEUE:", task.SandboxId, err)
								changeStatusAndPushToLogQueue(task, RunningNotInQueue)
							}
						}
					} else {
						// logger.LogAccToTaskId(task.Id, fmt.Sprintf("SANDBOX STATUS: %d", task.SandboxId))
						// slog.Println("SANDBOX STATUS: ", task.SandboxId)
						changeStatusAndPushToLogQueue(task, Aborted)
					}
				}

			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

const (
	MAX_CONCURRENT_LOG_QUEUE_TASKS = 50
)

func LogQueueHandler() {
	for {

		batchSize := 20

		tasks := fetchTasksFromQueue(LogQueue, batchSize)

		if len(tasks) > 0 {
			logger.Print("LOG QUEUE TASKS: ", len(tasks))
		} else {
			continue
		}

		// var wg sync.WaitGroup
		// sem := semaphore.NewWeighted(int64(MAX_CONCURRENT_LOG_QUEUE_TASKS))

		for _, task := range tasks {
			// wg.Add(1)
			processLogQueueTasks(task)
		}
		// wg.Wait()
	}
}

func processLogQueueTasks(task Task) {
	// defer wg.Done()
	// defer sem.Release(1)

	if task.Status == SandboxApiTimeout {
		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("%d - %d API TIMEOUT", task.Id, task.SandboxId))
		// slog.Printf("%d - %d API TIMEOUT\n", task.Id, task.SandboxId)
		// go deleteSandboxData(task.Id, task.SandboxId)
	} else if task.Status == Reported {
		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("%d - %d REPORTED", task.Id, task.SandboxId))
		// slog.Printf("%d - %d REPORTED\n", task.Id, task.SandboxId)

		score, err := fetchScoreFromSandBox(task.Id, task.SandboxId)
		if err != nil {
			score = 0
		}

		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("SCORE: %f", score))
		// slog.Printf("TASKID %d - SANDBOXID %d - SCORE %f\n", task.Id, task.SandboxId, score)

		dbOprs := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{},
		}
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, updateFOD(task, score))
		// saveVerdictInHash(task, score)
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, moveTaskToFinishedTable(task, false)...)
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, processDuplicateTasksForReported(task, score)...)
		err = dao.GormOperations(&dbOprs, config.Db, dao.EXEC)
		if err != nil {
			// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE CREATING TASK IN FINISHED TABLE: %v", err))
			// slog.Println("ERROR WHILE CREATING TASK IN FINISHED TABLE: ", task.SandboxId, err)
		}
		SendAcknowledgementToClientIp(task.Id)
		go deleteSandboxData(task.Id, task.SandboxId)
		go deleteLocalTask(task.Id)

	} else if task.Status == Aborted {
		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("%d - %d ABORTED", task.Id, task.SandboxId))
		// slog.Printf("%d - %d ABORTED\n", task.Id, task.SandboxId)
		dbOprs := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{},
		}
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, updateFOD(task, 0))
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, moveTaskToFinishedTable(task, true)...)
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, processDuplicateTasksForAborted(task)...)
		err := dao.GormOperations(&dbOprs, config.Db, dao.EXEC)
		if err != nil {
			// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE CREATING TASK IN FINISHED TABLE: %v", err))
			// slog.Println("ERROR WHILE CREATING TASK IN FINISHED TABLE: ", task.SandboxId, err)
		}
		SendAcknowledgementToClientIp(task.Id)

		// go deleteSandboxData(task.Id, task.SandboxId)
		go deleteLocalTask(task.Id)

	} else if task.Status == ReportedThroughClamd {
		slog.Println("REPORTED THROUGH CLAMD")
		score := float32(3.5)
		dbOprs := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{},
		}
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, updateFOD(task, score))
		// saveVerdictInHash(task, score)
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, moveTaskToFinishedTable(task, false)...)
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, processDuplicateTasksForReported(task, score)...)
		err := dao.GormOperations(&dbOprs, config.Db, dao.EXEC)
		if err != nil {
			// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE CREATING TASK IN FINISHED TABLE: %v", err))
			// slog.Println("ERROR WHILE CREATING TASK IN FINISHED TABLE: ", task.SandboxId, err)
		}
		SendAcknowledgementToClientIp(task.Id)
		go deleteLocalTask(task.Id)
	} else if task.Status == ReportedThroughQuickScope {
		slog.Println("REPORTED THROUGH QUICK SCOPE")
		score := float32(3.5)
		dbOprs := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{},
		}
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, updateFOD(task, score))
		// saveVerdictInHash(task, score)
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, moveTaskToFinishedTable(task, false)...)
		dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, processDuplicateTasksForReported(task, score)...)
		err := dao.GormOperations(&dbOprs, config.Db, dao.EXEC)
		if err != nil {
			// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE CREATING TASK IN FINISHED TABLE: %v", err))
			// slog.Println("ERROR WHILE CREATING TASK IN FINISHED TABLE: ", task.SandboxId, err)
		}
		SendAcknowledgementToClientIp(task.Id)
		go deleteLocalTask(task.Id)
	} else {
		err := updateStatusChangeDirectlyThroughDb(task)
		if err != nil {
			// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE UPDATING STATUS: %v", err))
			// slog.Println("ERROR WHILE UPDATING STATUS: ", task.Id, err)
		}
	}
}

func changeStatusAndPushToLogQueue(task Task, newStatus string) {

	err := pushToQueue(LogQueue, task, newStatus)
	if err != nil {
		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE PUSHING TASK INTO LOG QUEUE: %v", err))
		// slog.Println("ERROR WHILE PUSHING TASK INTO LOG QUEUE: ", task.Id, err)
		task.LogQueueFailed = true
		err = updateLiveTaskDirectlyThroughDb(task)
		if err != nil {
			// logger.LogAccToTaskId(task.Id, fmt.Sprintf("FAILED TO UPDATE LOG_QUEUE_FAILED %d STATUS TO %s: %v", task.Id, Aborted, err))
			// slog.Println("FAILED TO UPDATE LOG_QUEUE_FAILED %d STATUS TO %s: %v", task.Id, Aborted, err)
		}
	} else {
		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("PUSHED TASK INTO LOG QUEUE: %s", newStatus))
		// slog.Println("PUSHED TASK INTO LOG QUEUE: ", task.Id, newStatus)
	}
}

func deleteLocalTask(id int) {
	fp := extras.SANDBOX_FILE_PATHS + fmt.Sprintf("%d", id)

	err := os.Remove(fp)
	if err != nil {
		slog.Println("ERROR WHILE DELETING TASK FILE: ", id, err)
	}
}

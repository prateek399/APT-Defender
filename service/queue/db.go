package queues

import (
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/hash"
	"anti-apt-backend/util"
	"fmt"
	"time"
)

const (
	TaskLiveAnalysingTable = "task_live_analysis_tables"
	TaskFinishedTable      = "task_finished_tables"
	TaskDuplicateTable     = "task_duplicate_tables"
	FileOnDemandTable      = "file_on_demands"
)

func fetchNotInQueueTasks() ([]Task, error) {

	var tasks []Task

	queryString := fmt.Sprintf("SELECT task_live_analysis_tables.*, file_on_demands.submitted_time FROM %s INNER JOIN %s ON %s.id = %s.id WHERE task_live_analysis_tables.status IN ('%s', '%s') or task_live_analysis_tables.log_queue_failed = true", TaskLiveAnalysingTable, FileOnDemandTable, TaskLiveAnalysingTable, FileOnDemandTable, PendingNotInQueue, RunningNotInQueue)
	liveTask := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &tasks,
	}
	err := dao.GormOperations(&liveTask, config.Db, dao.EXEC)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func updateTaskStatus(taskId int, status string) error {

	queryString := fmt.Sprintf("UPDATE %s SET status = '%s' WHERE id = %d", TaskLiveAnalysingTable, status, taskId)
	liveTask := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
	}
	err := dao.GormOperations(&liveTask, config.Db, dao.EXEC)
	if err != nil {
		return err
	}
	return nil
}

func moveTaskToFinishedTable(task Task, aborted bool) []string {
	var queryStringArr []string
	// logger.LogAccToTaskId(task.Id, fmt.Sprintf("MOVING TASK TO FINISHED TABLE FOR TASK %d", task.Id))
	// slog.Println("MOVING TASK TO FINISHED TABLE FOR TASK %d", task.Id)

	queryString := fmt.Sprintf("DELETE FROM %s WHERE id = %d", TaskLiveAnalysingTable, task.Id)
	queryStringArr = append(queryStringArr, queryString)
	queryString = fmt.Sprintf("INSERT INTO %s (id, sandbox_id, aborted) VALUES (%d, %d, %t)", TaskFinishedTable, task.Id, task.SandboxId, aborted)
	// dbOprs.QueryExecSet = append(dbOprs.QueryExecSet, queryString)
	queryStringArr = append(queryStringArr, queryString)

	// err := dao.GormOperations(&dbOprs, config.Db, dao.EXEC)
	// if err != nil {
	// 	logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE CREATING TASK IN FINISHED TABLE: %v", err))
	// 	slog.Println("ERROR WHILE CREATING TASK IN FINISHED TABLE: ", task.SandboxId, err)
	// }
	return queryStringArr
}

func processDuplicateTasksForReported(task Task, score float32) []string {

	// logger.LogAccToTaskId(task.Id, fmt.Sprintf("PROCESSING DUPLICATE TASKS FOR REPORTED TASK %d", task.Id))
	// slog.Println("PROCESSING DUPLICATE TASKS FOR REPORTED TASK %d", task.Id)

	var duplicateTasks []int

	queryString := fmt.Sprintf("SELECT id FROM %s WHERE md5 = '%s' or sha = '%s' or sha256 = '%s'", TaskDuplicateTable, task.Md5, task.SHA, task.SHA256)
	duplicateTask := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &duplicateTasks,
	}
	err := dao.GormOperations(&duplicateTask, config.Db, dao.EXEC)
	if err != nil {
		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("SHOULD NOT COME(MYSQL) FETCHING DUPLICATE: %v", err))
		// slog.Println("SHOULD NOT COME(MYSQL) FETCHING DUPLICATE:", task.Id, err)
	}

	if len(duplicateTasks) == 0 {
		return []string{}
	}

	for _, duplicate := range duplicateTasks {
		go deleteLocalTask(duplicate)
	}

	var queryStringArr []string
	queryString = fmt.Sprintf("INSERT INTO %s (id, sandbox_id, aborted) VALUES", TaskFinishedTable)

	for _, duplicate := range duplicateTasks {
		queryString += fmt.Sprintf(" (%d, %d, 0),", duplicate, task.SandboxId)
		queryStringArr = append(queryStringArr, updateFOD(Task{Id: duplicate}, score))
	}

	queryString = queryString[:len(queryString)-1]
	queryStringArr = append(queryStringArr, queryString)

	idsToRemove := ""

	for _, duplicate := range duplicateTasks {
		idsToRemove += fmt.Sprintf("%d,", duplicate)
	}
	idsToRemove = idsToRemove[:len(idsToRemove)-1]

	queryString = fmt.Sprintf("DELETE FROM %s WHERE id IN (%s)", TaskDuplicateTable, idsToRemove)
	queryStringArr = append(queryStringArr, queryString)

	return queryStringArr
}

func processDuplicateTasksForAborted(task Task) []string {
	var queryStringArr []string
	// logger.LogAccToTaskId(task.Id, fmt.Sprintf("PROCESSING DUPLICATE TASKS FOR ABORTED TASK %d", task.Id))
	// slog.Println("PROCESSING DUPLICATE TASKS FOR ABORTED TASK %d", task.Id)

	var duplicateTask DuplicateTask

	var count int64 = 0

	queryString := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE md5 = '%s' or sha = '%s' or sha256 = '%s'", TaskDuplicateTable, task.Md5, task.SHA, task.SHA256)
	duplicateTaskData := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &count,
	}
	err := dao.GormOperations(&duplicateTaskData, config.Db, dao.EXEC)
	if err != nil {
		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE COUNTING DUPLICATES FOR TASK: %v", err))
		// slog.Println("ERROR WHILE COUNTING DUPLICATES FOR TASK:", task.Id, err)
	}

	if count > 0 {

		queryString = fmt.Sprintf("SELECT id FROM %s WHERE md5 = '%s' or sha = '%s' or sha256 = '%s' ORDER BY id LIMIT 1", TaskDuplicateTable, task.Md5, task.SHA, task.SHA256)
		dbOprs := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{queryString},
			Result:       &duplicateTask,
		}
		err = dao.GormOperations(&dbOprs, config.Db, dao.EXEC)
		if err != nil {
			// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE FETCHING DUPLICATE: %v", err))
			// slog.Println("ERROR WHILE FETCHING DUPLICATE:", task.Id, err)
		}

		queryString = fmt.Sprintf("INSERT INTO %s (id, status) VALUES (%d, '%s')", TaskLiveAnalysingTable, duplicateTask.Id, PendingNotInQueue)
		queryStringArr = append(queryStringArr, queryString)

		queryString = fmt.Sprintf("DELETE FROM %s WHERE id = %d", TaskDuplicateTable, duplicateTask.Id)
		queryStringArr = append(queryStringArr, queryString)
	}

	return queryStringArr
}

func updateStatusChangeDirectlyThroughDb(task Task) error {

	queryString := fmt.Sprintf("UPDATE %s SET status = '%s', sandbox_id = %d, queue_retry_count = %d, log_queue_failed = %t, sandbox_retry_count = %d, running_retry_count = %d  WHERE id = %d", TaskLiveAnalysingTable, task.Status, task.SandboxId, task.QueueRetryCount, task.LogQueueFailed, task.SandboxRetryCount, task.RunningRetryCount, task.Id)
	liveTask := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
	}
	err := dao.GormOperations(&liveTask, config.Db, dao.EXEC)
	if err != nil {
		return err
	}
	return nil
}

func updateFOD(task Task, score float32) string {
	rating := util.GetVerdict(score)

	var final_verdict string
	if score == 0 {
		final_verdict = extras.ALLOW
	} else {
		final_verdict = extras.BLOCK
	}

	queryString := fmt.Sprintf("UPDATE %s SET score = %f, rating = '%s', final_verdict = '%s', finished_time = '%s' WHERE id = %d", FileOnDemandTable, score, rating, final_verdict, time.Now().Format(extras.TIME_FORMAT), task.Id)
	// fileOnDemand := dao.DatabaseOperationsRepo{
	// 	QueryExecSet: []string{queryString},
	// }
	// err := dao.GormOperations(&fileOnDemand, config.Db, dao.EXEC)
	// if err != nil {
	// 	logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE UPDATING SCORE & RATING FOR TASKS %d, ERROR: %v", task.Id, err))
	// 	slog.Println("ERROR WHILE UPDATING SCORE & RATING FOR TASKS %d, ERROR: %v", task.Id, err)
	// }
	return queryString
}

func saveVerdictInHash(task Task, score float32) {

	queryString := fmt.Sprintf("SELECT id FROM %s WHERE id = %d", FileOnDemandTable, task.Id)
	fileOnDemand := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &task,
	}
	err := dao.GormOperations(&fileOnDemand, config.Db, dao.EXEC)
	if err != nil {
		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE FETCHING FOD FOR TASK %d, ERROR: %v", task.Id, err))
		// slog.Println("ERROR WHILE FETCHING FOD FOR TASK %d, ERROR: %v", task.Id, err)
	}

	if task.Md5 == "" {
		return
	}

	var verdict string
	if score > 0 {
		verdict = extras.BLOCK
	} else {
		verdict = extras.ALLOW
	}

	err = hash.SaveVerdict(task.Md5, verdict)
	if err != nil {
		// logger.LogAccToTaskId(task.Id, fmt.Sprintf("ERROR WHILE SAVING VERDICT IN HASH FOR TASK %d, ERROR: %v", task.Id, err))
		// slog.Println("ERROR WHILE SAVING VERDICT FOR TASK %d, ERROR: %v", task.Id, err)
	}

}

func updateLiveTaskDirectlyThroughDb(task Task) error {

	// logger.LogAccToTaskId(task.Id, fmt.Sprintf("LOG QUEUE FAILED: %t, SETTING STATUS TO %d", task.LogQueueFailed, task.Id))

	queryString := fmt.Sprintf("UPDATE %s SET status = '%s', sandbox_id = %d, queue_retry_count = %d, log_queue_failed = %t, sandbox_retry_count = %d, running_retry_count = %d  WHERE id = %d", TaskLiveAnalysingTable, task.Status, task.SandboxId, task.QueueRetryCount, task.LogQueueFailed, task.SandboxRetryCount, task.RunningRetryCount, task.Id)
	liveTask := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
	}
	err := dao.GormOperations(&liveTask, config.Db, dao.EXEC)
	if err != nil {
		return err
	}
	return nil
}

func fetchLiveTaskCount() (int, error) {

	queryString := fmt.Sprintf("SELECT COUNT(*) FROM %s", TaskLiveAnalysingTable)
	liveTaskCount := 0
	liveTask := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &liveTaskCount,
	}

	err := dao.GormOperations(&liveTask, config.Db, dao.EXEC)
	if err != nil {
		return liveTaskCount, err
	}

	return liveTaskCount, nil
}

func updateLogQueueFail(taskId int, flag bool) error {

	queryString := fmt.Sprintf("UPDATE %s SET log_queue_failed = %t WHERE id = %d", TaskLiveAnalysingTable, flag, taskId)
	liveTask := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
	}
	err := dao.GormOperations(&liveTask, config.Db, dao.EXEC)
	if err != nil {
		return err
	}
	return nil
}

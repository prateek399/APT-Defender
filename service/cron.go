package service

// import (
// 	"anti-apt-backend/dao"
// 	"anti-apt-backend/extras"
// 	"anti-apt-backend/logger"
// 	"anti-apt-backend/model"
// 	"anti-apt-backend/service/cuckoo"
// 	"anti-apt-backend/service/hash"
// 	"anti-apt-backend/util"
// 	"context"
// 	"encoding/csv"
// 	"fmt"
// 	"os"
// 	"strings"
// 	"sync"
// 	"sync/atomic"
// 	"time"

// 	"github.com/gookit/slog"
// 	"github.com/robfig/cron"
// )

// var (
// 	pendingFileTaskRunning int32
// 	pendingUrlTaskRunning  int32
// 	runningFileTaskRunning int32
// 	runningUrlTaskRunning  int32
// )

// func CountOfFreeVm() (int, error) {
// 	client := cuckoo.New(&cuckoo.Config{})
// 	allMachines, err := client.ListMachines(context.Background())
// 	count := 0
// 	if err != nil {
// 		return count, err
// 	}
// 	for _, machine := range allMachines {
// 		if !machine.Locked {
// 			count++
// 		}
// 	}
// 	return count, nil
// }

// type fodStruct struct {
// 	jobID        string
// 	fileOnDemand model.FileOnDemand
// }

// type uodStruct struct {
// 	jobID       string
// 	urlOnDemand model.UrlOnDemand
// }

// type WorkerPool struct {
// 	fileTasks chan fodStruct
// 	urlTasks  chan uodStruct
// 	workerWG  sync.WaitGroup
// }

// func NewWorkerPool() {
// 	fods, err := dao.FetchFileOnDemandProfile(map[string]any{})
// 	if err != nil {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching file on demand"))
// 		// fmt.Println("error in fetching file on demand: ", err)
// 	}

// 	uods, err := dao.FetchUrlOnDemandProfile(map[string]any{})
// 	if err != nil {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching url on demand"))
// 		// fmt.Println("error in fetching file on demand: ", err)
// 	}

// 	wp := &WorkerPool{
// 		fileTasks: make(chan fodStruct, len(fods)),
// 		urlTasks:  make(chan uodStruct, len(uods)),
// 	}

// 	wp.sendTasksToCuckoo()
// 	// wp.Close()
// }

// // func (wp *WorkerPool) Close() {
// // 	close(wp.fileTasks)
// // 	close(wp.urlTasks)
// // 	wp.workerWG.Wait()
// // }

// func (wp *WorkerPool) sendTasksToCuckoo() {
// 	go func() {
// 		ticker := time.NewTicker(10 * time.Second)
// 		defer ticker.Stop()
// 		for {
// 			select {
// 			case <-ticker.C:
// 				wp.fetchPendingTasks()
// 				// default:
// 				// 	return
// 			}
// 		}
// 	}()

// 	go func() {
// 		ticker := time.NewTicker(15 * time.Second)
// 		defer ticker.Stop()
// 		for {
// 			select {
// 			case <-ticker.C:
// 				wp.fetchRunningTasks()
// 				// default:
// 				// 	return
// 			}
// 		}
// 	}()
// }

// func (wp *WorkerPool) fetchPendingTasks() {
// 	fods, err := dao.FetchFileOnDemandProfile(map[string]any{})
// 	if err != nil {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching file on demand"))
// 		return
// 	}

// 	uods, err := dao.FetchUrlOnDemandProfile(map[string]any{})
// 	if err != nil {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching url on demand"))
// 		return
// 	}

// 	wp.adjustChannelBufferSizes(len(fods), len(uods))

// 	for jobID := range fods {
// 		wp.fileTasks <- fodStruct{
// 			jobID:        jobID,
// 			fileOnDemand: fods[jobID],
// 		}
// 	}

// 	for jobID := range uods {
// 		wp.urlTasks <- uodStruct{
// 			jobID:       jobID,
// 			urlOnDemand: uods[jobID],
// 		}
// 	}

// 	freeVMs, err := CountOfFreeVm()
// 	if err != nil {
// 		// fmt.Println("Error in getting free VMs: ", err)
// 		return
// 	}
// 	wp.workerWG.Add(freeVMs)
// 	for i := 0; i < freeVMs; i++ {
// 		go wp.processPendingTasks()
// 	}
// }

// func (wp *WorkerPool) fetchRunningTasks() {
// 	fods, err := dao.FetchFileOnDemandProfile(map[string]any{})
// 	if err != nil {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching file on demand"))
// 		return
// 	}

// 	uods, err := dao.FetchUrlOnDemandProfile(map[string]any{})
// 	if err != nil {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching url on demand"))
// 		return
// 	}

// 	wp.adjustChannelBufferSizes(len(fods), len(uods))

// 	for jobID := range fods {
// 		wp.fileTasks <- fodStruct{
// 			jobID:        jobID,
// 			fileOnDemand: fods[jobID],
// 		}
// 	}

// 	for jobID := range uods {
// 		wp.urlTasks <- uodStruct{
// 			jobID:       jobID,
// 			urlOnDemand: uods[jobID],
// 		}
// 	}

// 	freeVMs, err := CountOfFreeVm()
// 	if err != nil {
// 		// fmt.Println("Error in getting free VMs: ", err)
// 		return
// 	}
// 	wp.workerWG.Add(freeVMs)
// 	for i := 0; i < freeVMs; i++ {
// 		go wp.processRunningTasks()
// 	}
// }

// func (wp *WorkerPool) adjustChannelBufferSizes(fileTasksSize, urlTasksSize int) {
// 	// Resize fileTasks channel
// 	if fileTasksSize > 0 {
// 		newFileTasks := make(chan fodStruct, fileTasksSize)
// 		wp.fileTasks = newFileTasks
// 	}

// 	// Resize urlTasks channel
// 	if urlTasksSize > 0 {
// 		newUrlTasks := make(chan uodStruct, urlTasksSize)
// 		wp.urlTasks = newUrlTasks
// 	}
// }

// func (wp *WorkerPool) processPendingTasks() {
// 	defer wp.workerWG.Done()
// 	// fmt.Println("AA rha hu andar 3")
// 	select {
// 	case fod := <-wp.fileTasks:
// 		if fod.fileOnDemand.Status == extras.PENDING {
// 			newFileLocation := fmt.Sprintf(extras.DATABASE_PATH+"files/%s/%s", fod.jobID, fod.fileOnDemand.FileName.Filename)
// 			if hash.IsAllowedHash(newFileLocation) {
// 				newFileOnDemand := fod.fileOnDemand
// 				newFileOnDemand.Rating = model.Clean
// 				newFileOnDemand.Status = extras.CLEAN_FOUND_FROM_HASH
// 				newFileOnDemand.FinalVerdict = extras.ALLOW
// 				newFileOnDemand.Comments = "File is allowed by the system, as it is previously scanned and found clean."
// 				if err := dao.SaveProfile([]interface{}{map[string]model.FileOnDemand{fod.jobID: newFileOnDemand}}, extras.PATCH); err != nil {
// 					logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 					// fmt.Println("error in saving profile: ", err)
// 				}
// 				return
// 			}

// 			if hash.IsBlockedHash(newFileLocation) {
// 				newFileOnDemand := fod.fileOnDemand
// 				newFileOnDemand.Status = extras.MALWARE_FOUND_FROM_HASH
// 				newFileOnDemand.FinalVerdict = extras.BLOCK
// 				newFileOnDemand.Rating = model.Malicious
// 				newFileOnDemand.Comments = "File is blocked by the system, as it is previously scanned and found malware from internet source."
// 				if err := dao.SaveProfile([]interface{}{map[string]model.FileOnDemand{fod.jobID: newFileOnDemand}}, extras.PATCH); err != nil {
// 					logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 					// fmt.Println("error in saving profile: ", err)
// 				}
// 				return
// 			}

// 			if !hash.IsFileBeingProcessed(newFileLocation) {
// 				// fmt.Println("AA rha hu andar 2")
// 				client := cuckoo.New(&cuckoo.Config{})
// 				if err := client.CreateTaskFile(context.Background(), fod.jobID, newFileLocation); err != nil {
// 					// fmt.Println("error in creating task file: ", err)
// 					return
// 				}
// 				hash.AddFileToQueue(newFileLocation)
// 			}
// 		}

// 	case uod := <-wp.urlTasks:
// 		if uod.urlOnDemand.Status == extras.PENDING {
// 			verdict := CheckUrlCache(uod.urlOnDemand.UrlName)
// 			// fmt.Println("Verdict : ", verdict)

// 			if verdict == extras.CLEAN || verdict == extras.MALICIOUS {
// 				newUrlOnDemand := uod.urlOnDemand
// 				if verdict == extras.CLEAN {
// 					newUrlOnDemand.Rating = model.Clean
// 					newUrlOnDemand.FinalVerdict = extras.ALLOW
// 					newUrlOnDemand.Comments = "Url is allowed by the system, as it is previously scanned and found clean."
// 					newUrlOnDemand.Status = extras.CLEAN_FOUND_FROM_CACHE
// 				} else {
// 					newUrlOnDemand.Rating = model.Malicious
// 					newUrlOnDemand.FinalVerdict = extras.BLOCK
// 					newUrlOnDemand.Comments = "Url is blocked by the system, as it is previously scanned and found malicious."
// 					newUrlOnDemand.Status = extras.MALWARE_FOUND_FROM_CACHE
// 				}
// 				if err := dao.SaveProfile([]interface{}{map[string]model.UrlOnDemand{uod.jobID: newUrlOnDemand}}, extras.PATCH); err != nil {
// 					logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 					// fmt.Println("error in saving profile: ", err)
// 				}
// 				return
// 			}

// 			if verdict == extras.ANALYSING {
// 				// fmt.Println("Still analysing url: ", uod.urlOnDemand.UrlName)
// 				return
// 			}

// 			client := cuckoo.New(&cuckoo.Config{})
// 			if err := client.CreateTaskUrl(context.Background(), uod.jobID, uod.urlOnDemand.UrlName); err != nil {
// 				// fmt.Println("error in creating task file: ", err)
// 				return
// 			}
// 		}

// 	default:
// 		return
// 	}
// }

// func (wp *WorkerPool) processRunningTasks() {
// 	defer wp.workerWG.Done()

// 	// fmt.Println("AA rha hu andar 4")
// 	select {
// 	case fod := <-wp.fileTasks:
// 		if fod.fileOnDemand.Status == extras.RUNNING {
// 			client := cuckoo.New(&cuckoo.Config{})
// 			taskID := -1
// 			for id := range fod.fileOnDemand.TaskID {
// 				taskID = id
// 			}
// 			if taskID <= 0 {
// 				return
// 			}

// 			task, err := client.TasksView(context.Background(), taskID)
// 			if err != nil {
// 				// fmt.Println("error in viewing task: ", err)
// 				return
// 			}

// 			if task == nil {
// 				// fmt.Println("task not found")
// 				return
// 			}
// 			if task.Status == model.StatusRunning {
// 				newFileInDemand := fod.fileOnDemand
// 				newFileInDemand.Status = string(task.Status)
// 				newFileInDemand.TaskID[taskID] = *task
// 				if err := dao.SaveProfile([]interface{}{map[string]model.FileOnDemand{fod.jobID: newFileInDemand}}, extras.PATCH); err != nil {
// 					logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 					// fmt.Println("error in saving profile: ", err)
// 				}
// 			}
// 			if task.Status == model.StatusCompleted || task.Status == model.StatusReported {
// 				newFileInDemand := fod.fileOnDemand
// 				newFileInDemand.Status = string(task.Status)
// 				newFileInDemand.TaskID = make(map[int]model.Task)
// 				newFileInDemand.TaskID[taskID] = *task
// 				if fod.fileOnDemand.TaskReport == nil {
// 					report, err := client.TasksReport(context.Background(), taskID)
// 					if err != nil {
// 						// fmt.Println("error in getting report: ", err)
// 						return
// 					}
// 					if report == nil {
// 						// fmt.Println("report not found")
// 						return
// 					}
// 					rating := util.GetVerdict(report.Info.Score)
// 					newFileInDemand.Rating = rating
// 					newFileInDemand.TaskReport = report
// 					newFileInDemand.FinalVerdict = extras.BLOCK
// 					if rating == model.Clean {
// 						newFileInDemand.FinalVerdict = extras.ALLOW
// 					}
// 					newFileLocation := fmt.Sprintf(extras.DATABASE_PATH+"files/%s/%s", fod.jobID, fod.fileOnDemand.FileName.Filename)
// 					hash.RemoveFileFromQueue(newFileLocation)
// 					hashErr := hash.SaveVerdict(report.Target.File.Md5, newFileInDemand.FinalVerdict)
// 					if hashErr != nil {
// 						logger.LoggerFunc("error", logger.LoggerMessage("taskLog:couldn't save hash verdict"))
// 						// fmt.Println("hashErr: " + hashErr.Error())
// 					}
// 				}
// 				if err := dao.SaveProfile([]interface{}{map[string]model.FileOnDemand{fod.jobID: newFileInDemand}}, extras.PATCH); err != nil {
// 					logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 					// fmt.Println("error in saving profile: ", err)
// 				}
// 				logger.LoggerFunc("info", logger.LoggerMessage("taskLog:file status is updated to reported"))
// 			}

// 		} else {
// 			if fod.fileOnDemand.Status == extras.REPORTED || fod.fileOnDemand.Status == extras.COMPLETED {
// 				if err := os.RemoveAll(extras.DATABASE_PATH + "files/" + fod.jobID); err != nil {
// 					// fmt.Println("error in removing directory: ", err)
// 				}
// 			}
// 		}

// 	case uod := <-wp.urlTasks:
// 		if uod.urlOnDemand.Status == extras.RUNNING {
// 			client := cuckoo.New(&cuckoo.Config{})
// 			taskID := -1
// 			for id := range uod.urlOnDemand.TaskID {
// 				taskID = id
// 			}
// 			if taskID <= 0 {
// 				// fmt.Println("task not found")
// 				return
// 			}

// 			task, err := client.TasksView(context.Background(), taskID)
// 			if err != nil {
// 				// fmt.Println("error in viewing task: ", err)
// 				return
// 			}

// 			if task == nil {
// 				// fmt.Println("task not found")
// 				return
// 			}

// 			if task.Status == model.StatusRunning {
// 				newUrlInDemand := uod.urlOnDemand
// 				newUrlInDemand.Status = string(task.Status)
// 				newUrlInDemand.TaskID[taskID] = *task
// 				if err := dao.SaveProfile([]interface{}{map[string]model.UrlOnDemand{uod.jobID: newUrlInDemand}}, extras.PATCH); err != nil {
// 					logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 				}
// 			}
// 			if task.Status == model.StatusCompleted || task.Status == model.StatusReported {
// 				newUrlOnDemand := uod.urlOnDemand
// 				newUrlOnDemand.Status = string(task.Status)
// 				newUrlOnDemand.TaskID[taskID] = *task
// 				if uod.urlOnDemand.TaskReport == nil {
// 					report, err := client.TasksReport(context.Background(), taskID)
// 					if err != nil {
// 						// fmt.Println("error in getting report: ", err)
// 						return
// 					}
// 					if report == nil {
// 						// fmt.Println("report not found")
// 						return
// 					}

// 					rating := util.GetVerdict(report.Info.Score)
// 					newUrlOnDemand.Rating = rating
// 					newUrlOnDemand.TaskReport = report
// 					if rating == model.Clean {
// 						hash.ReplaceUrlCache(newUrlOnDemand.UrlName, extras.CLEAN)
// 						newUrlOnDemand.FinalVerdict = extras.ALLOW
// 					} else {
// 						hash.ReplaceUrlCache(newUrlOnDemand.UrlName, extras.MALICIOUS)
// 						newUrlOnDemand.FinalVerdict = extras.BLOCK
// 					}
// 				}
// 				if err := dao.SaveProfile([]interface{}{map[string]model.UrlOnDemand{uod.jobID: newUrlOnDemand}}, extras.PATCH); err != nil {
// 					logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 				}
// 				logger.LoggerFunc("info", logger.LoggerMessage("taskLog:url status is updated to reported"))
// 			}

// 		}
// 	default:
// 		return
// 	}

// }

// func CronTask() error {

// 	var err error

// 	c := cron.New()
// 	err = c.AddFunc("@every 10s", func() {

// 		if atomic.CompareAndSwapInt32(&pendingFileTaskRunning, 0, 1) {
// 			CronPendingFileTask()
// 			atomic.StoreInt32(&pendingFileTaskRunning, 0)
// 		}

// 		if atomic.CompareAndSwapInt32(&pendingUrlTaskRunning, 0, 1) {
// 			CronPendingUrlTask()
// 			atomic.StoreInt32(&pendingUrlTaskRunning, 0)
// 		}
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	err = c.AddFunc("@every 15s", func() {

// 		if atomic.CompareAndSwapInt32(&runningFileTaskRunning, 0, 1) {
// 			CronRunningFileTask()
// 			atomic.StoreInt32(&runningFileTaskRunning, 0)
// 		}

// 		if atomic.CompareAndSwapInt32(&runningUrlTaskRunning, 0, 1) {
// 			CronRunningUrlTask()
// 			atomic.StoreInt32(&runningUrlTaskRunning, 0)
// 		}
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	c.Start()

// 	return nil
// }

// func CronPendingFileTask() {
// 	var err error
// 	var fileOnDemand = make(map[string]model.FileOnDemand)
// 	fileOnDemand, err = dao.FetchFileOnDemandProfile(map[string]any{})
// 	if err == extras.ErrNoRecordForFileOnDemand {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching file on demand"))
// 		// fmt.Println(extras.ErrNoRecordForFileOnDemand)
// 	} else if err != nil {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching file on demand"))
// 		// fmt.Println("error in fetching file on demand: ", err)
// 	}

// 	for jobID := range fileOnDemand {
// 		freeVMs, err := CountOfFreeVm()
// 		if err != nil {
// 			logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching free vm's"))
// 		}

// 		slog.Println("PROCESS PENDING FILE: ", fileOnDemand[jobID].FileName.Filename)
// 		if fileOnDemand[jobID].Status == extras.PENDING {
// 			newFileLocation := fmt.Sprintf(extras.DATABASE_PATH+"files/%s/%s", jobID, fileOnDemand[jobID].FileName.Filename)
// 			if hash.IsAllowedHash(newFileLocation) {
// 				newFileOnDemand := fileOnDemand[jobID]
// 				newFileOnDemand.Rating = model.Clean
// 				newFileOnDemand.Status = extras.CLEAN_FOUND_FROM_HASH
// 				newFileOnDemand.FinalVerdict = extras.ALLOW
// 				newFileOnDemand.Comments = "File is allowed by the system, as it is previously scanned and found clean."
// 				if err := dao.SaveProfile([]interface{}{map[string]model.FileOnDemand{jobID: newFileOnDemand}}, extras.PATCH); err != nil {
// 					logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 					// fmt.Println("error in saving profile: ", err)
// 				}
// 				continue
// 			}

// 			if hash.IsBlockedHash(newFileLocation) {
// 				newFileOnDemand := fileOnDemand[jobID]
// 				newFileOnDemand.Status = extras.MALWARE_FOUND_FROM_HASH
// 				newFileOnDemand.FinalVerdict = extras.BLOCK
// 				newFileOnDemand.Rating = model.Malicious
// 				newFileOnDemand.Comments = "File is blocked by the system, as it is previously scanned and found malware from internet source."
// 				if err := dao.SaveProfile([]interface{}{map[string]model.FileOnDemand{jobID: newFileOnDemand}}, extras.PATCH); err != nil {
// 					logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 					// fmt.Println("error in saving profile: ", err)
// 				}
// 				continue
// 			}

// 			slog.Println("CHECK IF WE CAN SEND THIS FILE TO CUCKOO FURTHER: ", fileOnDemand[jobID].FileName.Filename)
// 			if freeVMs > 0 && !hash.IsFileBeingProcessed(newFileLocation) {
// 				client := cuckoo.New(&cuckoo.Config{})
// 				slog.Println("SEND TO CUCKOO: ", fileOnDemand[jobID].FileName.Filename)
// 				if err := client.CreateTaskFile(context.Background(), jobID, newFileLocation); err != nil {
// 					// fmt.Println("error in creating task file: ", err)
// 					continue
// 				}
// 				freeVMs--
// 			} else {
// 				slog.Println("FILE ALREADY IN PROCESS: ", fileOnDemand[jobID].FileName.Filename)
// 			}
// 		}
// 	}
// }

// func CronPendingUrlTask() {
// 	var err error
// 	var urlOnDemand = make(map[string]model.UrlOnDemand)
// 	urlOnDemand, err = dao.FetchUrlOnDemandProfile(map[string]any{})
// 	if err == extras.ErrNoRecordForUrlOnDemand {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching url on demand"))
// 		// fmt.Println(extras.ErrNoRecordForFileOnDemand)
// 	} else if err != nil {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching url on demand"))
// 		// fmt.Println("error in fetching url on demand: ", err)
// 	}

// 	for jobID := range urlOnDemand {
// 		freeVms, err := CountOfFreeVm()
// 		if err != nil {
// 			logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching free vm's"))
// 		}
// 		if urlOnDemand[jobID].Status == extras.PENDING {
// 			verdict := CheckUrlCache(urlOnDemand[jobID].UrlName)
// 			// fmt.Println("Verdict : ", verdict)

// 			if verdict == extras.CLEAN || verdict == extras.MALICIOUS {
// 				newUrlOnDemand := urlOnDemand[jobID]
// 				if verdict == extras.CLEAN {
// 					newUrlOnDemand.Rating = model.Clean
// 					newUrlOnDemand.FinalVerdict = extras.ALLOW
// 					newUrlOnDemand.Comments = "Url is allowed by the system, as it is previously scanned and found clean."
// 					newUrlOnDemand.Status = extras.CLEAN_FOUND_FROM_CACHE
// 				} else {
// 					newUrlOnDemand.Rating = model.Malicious
// 					newUrlOnDemand.FinalVerdict = extras.BLOCK
// 					newUrlOnDemand.Comments = "Url is blocked by the system, as it is previously scanned and found malicious."
// 					newUrlOnDemand.Status = extras.MALWARE_FOUND_FROM_CACHE
// 				}
// 				if err := dao.SaveProfile([]interface{}{map[string]model.UrlOnDemand{jobID: newUrlOnDemand}}, extras.PATCH); err != nil {
// 					logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 					// fmt.Println("error in saving profile: ", err)
// 				}
// 				continue
// 			}

// 			if verdict == extras.ANALYSING {
// 				continue
// 			}

// 			if freeVms > 0 {
// 				client := cuckoo.New(&cuckoo.Config{})
// 				if err := client.CreateTaskUrl(context.Background(), jobID, urlOnDemand[jobID].UrlName); err != nil {
// 					// fmt.Println("error in creating task file: ", err)
// 					continue
// 				}
// 				freeVms--
// 			}
// 		}
// 	}
// }

// func CheckUrlCache(url string) int {
// 	present, urlCache := hash.GetUrlCaches(url)
// 	if !present {
// 		return extras.NOT_PRESENT
// 	}

// 	if urlCache.Time.Add(24 * time.Hour).Before(time.Now()) {
// 		return extras.NOT_PRESENT
// 	} else {
// 		return urlCache.Verdict
// 	}
// }

// func CronRunningFileTask() {
// 	var err error
// 	var fileOnDemand = make(map[string]model.FileOnDemand)
// 	fileOnDemand, err = dao.FetchFileOnDemandProfile(map[string]any{})
// 	if err == extras.ErrNoRecordForFileOnDemand {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching file on demand"))
// 		// fmt.Println(extras.ErrNoRecordForFileOnDemand)
// 	} else if err != nil {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching file on demand"))
// 		// fmt.Println("error in fetching file on demand: ", err)
// 	}

// 	for jobID := range fileOnDemand {
// 		if fileOnDemand[jobID].Status == extras.RUNNING || fileOnDemand[jobID].Status == extras.COMPLETED {
// 			var task *model.Task
// 			client := cuckoo.New(&cuckoo.Config{})
// 			for taskID := range fileOnDemand[jobID].TaskID {
// 				task, err = client.TasksView(context.Background(), taskID)
// 				slog.Printf("task: %v", task)
// 				slog.Println("Err: ", err)
// 				if err != nil {
// 					// fmt.Println("error in viewing task: ", err)
// 					continue
// 				}
// 				if task == nil {
// 					continue
// 				}
// 				t1, _ := time.Parse(extras.TIME_FORMAT, fileOnDemand[jobID].SubmittedTime)
// 				if task.Status == "" && time.Now().After(t1.Add(time.Second*100)) {
// 					file := fileOnDemand[jobID]
// 					file.Status = extras.ABORTED
// 					file.Comments = "Task aborted. Please re-upload your the task"
// 					if err := dao.SaveProfile([]interface{}{map[string]model.FileOnDemand{jobID: file}}, extras.PATCH); err != nil {
// 						logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 					}
// 				}
// 				// fmt.Println("task: ", task)
// 				if task.Status == model.StatusRunning {
// 					newFileInDemand := fileOnDemand[jobID]
// 					newFileInDemand.Status = string(task.Status)
// 					newFileInDemand.TaskID[taskID] = *task
// 					// fmt.Println("new file on demand: ", newFileInDemand)
// 					if err := dao.SaveProfile([]interface{}{map[string]model.FileOnDemand{jobID: newFileInDemand}}, extras.PATCH); err != nil {
// 						logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 						// fmt.Println("error in saving profile: ", err)
// 					}
// 				}
// 				if task.Status == model.StatusCompleted || task.Status == model.StatusReported {
// 					newFileInDemand := fileOnDemand[jobID]
// 					newFileInDemand.Status = string(task.Status)
// 					newFileInDemand.TaskID = make(map[int]model.Task)
// 					newFileInDemand.TaskID[taskID] = *task
// 					slog.Printf("new file on demand: %v", newFileInDemand)
// 					slog.Printf("task report: %v", fileOnDemand[jobID].TaskReport)
// 					if fileOnDemand[jobID].TaskReport == nil {
// 						report, err := client.TasksReport(context.Background(), taskID)
// 						if err != nil {
// 							// fmt.Println("error in getting report: ", err)
// 							continue
// 						}
// 						if report == nil {
// 							continue
// 						}
// 						rating := util.GetVerdict(report.Info.Score)
// 						newFileInDemand.Rating = rating
// 						newFileInDemand.TaskReport = report
// 						newFileInDemand.FinalVerdict = extras.BLOCK
// 						if rating == model.Clean {
// 							newFileInDemand.FinalVerdict = extras.ALLOW
// 						}
// 						newFileLocation := fmt.Sprintf(extras.DATABASE_PATH+"files/%s/%s", jobID, fileOnDemand[jobID].FileName.Filename)
// 						slog.Printf("Before Removing for quewe")
// 						hash.RemoveFileFromQueue(newFileLocation)
// 						slog.Printf("After Removing for quewe")

// 						hashErr := hash.SaveVerdict(report.Target.File.Md5, newFileInDemand.FinalVerdict)
// 						if hashErr != nil {
// 							logger.LoggerFunc("error", logger.LoggerMessage("taskLog:couldn't save hash verdict"))
// 							// fmt.Println("hashErr: " + hashErr.Error())
// 						}
// 						slog.Printf("After saving verdict")
// 					}
// 					if err := dao.SaveProfile([]interface{}{map[string]model.FileOnDemand{jobID: newFileInDemand}}, extras.PATCH); err != nil {
// 						logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 						// fmt.Println("error in saving profile: ", err)
// 					}
// 					slog.Printf("Reported successfully")
// 					logger.LoggerFunc("info", logger.LoggerMessage("taskLog:file status is updated to reported"))
// 				}
// 			}
// 			// fmt.Println("Running: ", fileOnDemand)
// 		} else {
// 			if fileOnDemand[jobID].Status == extras.REPORTED || fileOnDemand[jobID].Status == extras.COMPLETED {
// 				if err := os.RemoveAll(extras.DATABASE_PATH + "files/" + jobID); err != nil {
// 					// fmt.Println("error in removing directory: ", err)
// 				}
// 			}
// 			// fmt.Println(extras.ERR_STATUS_IS_NOT_PENDING)
// 		}
// 	}
// }

// func CronRunningUrlTask() {
// 	var err error
// 	var UrlOnDemand = make(map[string]model.UrlOnDemand)
// 	UrlOnDemand, err = dao.FetchUrlOnDemandProfile(map[string]any{})
// 	if err == extras.ErrNoRecordForUrlOnDemand {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching url on demand"))
// 		// fmt.Println(extras.ErrNoRecordForUrlOnDemand)
// 	} else if err != nil {
// 		logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in fetching url on demand"))
// 		// fmt.Println("error in fetching url on demand: ", err)
// 	}

// 	for jobID := range UrlOnDemand {
// 		if UrlOnDemand[jobID].Status == extras.RUNNING || UrlOnDemand[jobID].Status == extras.COMPLETED {
// 			var task *model.Task
// 			client := cuckoo.New(&cuckoo.Config{})
// 			for taskID := range UrlOnDemand[jobID].TaskID {
// 				// 	taskID = taskid
// 				task, err = client.TasksView(context.Background(), taskID)
// 				t1, _ := time.Parse(extras.TIME_FORMAT, UrlOnDemand[jobID].SubmittedTime)
// 				slog.Println("submit time: ")
// 				if task.Status == "" && time.Now().After(t1.Add(time.Second*100)) {
// 					url := UrlOnDemand[jobID]
// 					url.Status = extras.ABORTED
// 					url.Comments = "Task aborted. Please re-upload your the task"
// 					if err := dao.SaveProfile([]interface{}{map[string]model.UrlOnDemand{jobID: url}}, extras.PATCH); err != nil {
// 						logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 					}
// 				}
// 				if err != nil {
// 					continue
// 				}
// 				if task == nil {
// 					continue
// 				}

// 				if task.Status == model.StatusRunning {
// 					newUrlInDemand := UrlOnDemand[jobID]
// 					newUrlInDemand.Status = string(task.Status)
// 					newUrlInDemand.TaskID[taskID] = *task
// 					if err := dao.SaveProfile([]interface{}{map[string]model.UrlOnDemand{jobID: newUrlInDemand}}, extras.PATCH); err != nil {
// 						logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 					}
// 				}
// 				if task.Status == model.StatusCompleted || task.Status == model.StatusReported {
// 					newUrlOnDemand := UrlOnDemand[jobID]
// 					newUrlOnDemand.Status = string(task.Status)
// 					newUrlOnDemand.TaskID[taskID] = *task
// 					if UrlOnDemand[jobID].TaskReport == nil {
// 						report, err := client.TasksReport(context.Background(), taskID)
// 						if err != nil {
// 							// fmt.Println("error in getting report: ", err)
// 							continue
// 						}
// 						if report == nil {
// 							continue
// 						}
// 						rating := util.GetVerdict(report.Info.Score)

// 						malicipusUrls := FetchMaliciousUrls()

// 						if maliciousUrlContains(malicipusUrls, newUrlOnDemand.UrlName) {
// 							rating = model.Critical
// 						} else {
// 							rating = model.Clean
// 						}

// 						newUrlOnDemand.Rating = rating
// 						newUrlOnDemand.TaskReport = report
// 						if rating == model.Clean {
// 							hash.ReplaceUrlCache(newUrlOnDemand.UrlName, extras.CLEAN)
// 							newUrlOnDemand.FinalVerdict = extras.ALLOW
// 						} else {
// 							hash.ReplaceUrlCache(newUrlOnDemand.UrlName, extras.MALICIOUS)
// 							newUrlOnDemand.FinalVerdict = extras.BLOCK
// 						}
// 					}
// 					if err := dao.SaveProfile([]interface{}{map[string]model.UrlOnDemand{jobID: newUrlOnDemand}}, extras.PATCH); err != nil {
// 						logger.LoggerFunc("error", logger.LoggerMessage("taskLog:error in saving profile"))
// 						// fmt.Println("error in saving profile: ", err)
// 					}
// 					logger.LoggerFunc("info", logger.LoggerMessage("taskLog:url status is updated to reported"))
// 				}
// 			}
// 		}
// 	}
// }

// func maliciousUrlContains(maliciousUrls []string, url string) bool {
// 	for _, maliciousUrl := range maliciousUrls {
// 		if strings.Contains(url, maliciousUrl) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func FetchMaliciousUrls() []string {
// 	// open csv file
// 	file, err := os.Open(extras.TEMP_MALICIOUS_URLS_FILE)
// 	if err != nil {
// 		// fmt.Println("error in opening malicious urls file: ", err)
// 		return []string{}
// 	}
// 	defer file.Close()

// 	// read csv values using csv.Reader
// 	csvReader := csv.NewReader(file)
// 	fields, _ := csvReader.ReadAll()

// 	var maliciousUrls []string
// 	for _, field := range fields {
// 		maliciousUrls = append(maliciousUrls, field[0])
// 	}

// 	return maliciousUrls
// }

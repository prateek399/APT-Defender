package service

import (
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/service/cuckoo"
	queues "anti-apt-backend/service/queue"
	"anti-apt-backend/util"
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

// import (
// 	"anti-apt-backend/dao"
// 	"anti-apt-backend/extras"
// 	"anti-apt-backend/model"
// 	"anti-apt-backend/service/cuckoo"
// 	"anti-apt-backend/util"
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"sort"
// 	"strings"
// 	"time"
// )

func GetAllProfiles(modelName string, action string) model.APIResponse {
	var resp model.APIResponse
	var err error
	var data interface{}

	switch modelName {
	case extras.ADMIN:
		data, err = dao.FetchAdminProfile(map[string]any{})
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

	case extras.USERAUTHENTICATION:
		data, err = dao.FetchUserAuthProfile(map[string]any{})
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

	case extras.FEATURE:
		data, err = dao.FetchFeatureProfile(map[string]any{})
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

	case extras.ROLE:
		data, err = dao.FetchRoleProfile(map[string]any{})
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

	case extras.ROLEANDACTION:
		data, err = dao.FetchRoleAndActionProfile(map[string]any{})
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

	case extras.SCANPROFILE:
		data, err = dao.FetchScanProfile(map[string]any{})
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

	case extras.DEVICE:
		data, err = dao.FetchDeviceProfile(map[string]any{})
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

	case extras.BACKUP:
		data, err = dao.FetchBackupConfigTime(map[string]any{})
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

	case extras.RESTORE:
		data, err = dao.FetchRestoreConfigTime(map[string]any{})
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

	case extras.FILEONDEMAND, extras.URLONDEMAND:
		fods, err := FetchAllFODs()
		if err != nil {
			// slog.Println("ERROR WHILE FETCHING: ", err)
		}

		// slog.Println("FODS: ", fods)

		uods, err := FetchAllUODs()
		if err != nil {
			// slog.Println("ERROR WHILE FETCHING: ", err)
		}

		switch action {
		case "job-queue":
			tasks := []map[string]any{}
			for _, fod := range fods {
				inputType := "MANUAL FILE CHECK"
				if fod.FromDevice {
					inputType = "DEVICE FILE CHECK"
				}

				var duration float64 = -1
				finishTime := "IN PROGRESS"
				if fod.FinishedTime.Valid {
					finishTime = fod.FinishedTime.Time.Format(extras.TIME_FORMAT)
					duration = fod.FinishedTime.Time.Sub(fod.SubmittedTime).Seconds()
				}

				tasks = append(tasks, map[string]any{
					"job_id":         fod.Id,
					"submitted_time": fod.SubmittedTime.Format(extras.TIME_FORMAT),
					"file_name":      fod.FileName,
					"input_type":     inputType,
					"file_type":      fod.ContentType,
					"queued":         fod.Status,
					// "vm_instance":    fod.OsSupported,
					"duration":    duration,
					"finished_on": finishTime,
				})
			}

			for _, uod := range uods {
				inputType := "MANUAL URL CHECK"
				if uod.FromDevice {
					inputType = "DEVICE URL CHECK"
				}

				var duration float64 = -1
				finishTime := "IN PROGRESS"
				if uod.FinishedTime.Valid {
					finishTime = uod.FinishedTime.Time.Format(extras.TIME_FORMAT)
					duration = uod.FinishedTime.Time.Sub(uod.SubmittedTime).Seconds()
				}

				tasks = append(tasks, map[string]any{
					"job_id":         uod.Id,
					"submitted_time": uod.SubmittedTime.Format(extras.TIME_FORMAT),
					"url":            uod.UrlName,
					"input_type":     inputType,
					"file_type":      "URL",
					"queued":         uod.Status,
					// "vm_instance":    uod.OsSupported,
					"duration":    duration,
					"finished_on": finishTime,
				})
			}

			data = tasks

		case "vm-job":
			tasks := []map[string]any{}

			// sandBoxTasks, err := client.ListAllTasks(context.Background())
			// if err != nil {
			// 	slog.Println("ERROR WHILE FETCHING: ", err)
			// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_SANDBOX_TASKS, err)
			// }

			// machineAndTaskMap := make(map[any]int)
			// sandboxIDs := []int{}
			// for _, sandboxTask := range sandBoxTasks {
			// 	slog.Println("SANDBOX VMBOX: ", sandboxTask.Guest)
			// 	if guest, ok := sandboxTask.Guest.(map[string]any); ok {
			// 		machine, _ := guest["name"].(string)
			// 		machineStatus, _ := guest["status"].(string)
			// 		slog.Println("MACHINE NAME: ", machine)
			// 		slog.Println("MACHINE STATUS: ", machineStatus)
			// 		if machineStatus == "running" {
			// 			if _, ok := machineAndTaskMap[machine]; !ok {
			// 				machineAndTaskMap[machine] = sandboxTask.ID
			// 				sandboxIDs = append(sandboxIDs, sandboxTask.ID)
			// 			}
			// 		}
			// 	}

			// }

			// type runningTask struct {
			// 	SandboxId int
			// 	FileName  string
			// }
			// var runningTasks []runningTask
			// err = config.Db.Model(&model.TaskLiveAnalysisTable{}).Where("task_live_analysis_tables.sandbox_id IN ?", sandboxIDs).Joins("INNER JOIN file_on_demands ON file_on_demands.id = task_live_analysis_tables.id").Select("task_live_analysis_tables.sandbox_id, file_on_demands.file_name").Find(&runningTasks).Error
			// if err != nil {
			// 	slog.Println("ERROR WHILE FETCHING: ", err)
			// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_SANDBOX_TASKS, err)
			// }

			// slog.Println("Running tasks: ", runningTasks)

			// sandboxTaskAndFodMap := make(map[int]string)
			// for _, runningTask := range runningTasks {
			// 	sandboxTaskAndFodMap[runningTask.SandboxId] = runningTask.FileName
			// }

			// for machine, sandboxId := range machineAndTaskMap {
			// 	tasks = append(tasks, map[string]any{
			// 		"machine_name": machine,
			// 		"file_name":    sandboxTaskAndFodMap[sandboxId],
			// 		// "machine_platform": platform,
			// 		"progress":       "running",
			// 		"machine_status": "running",
			// 	})
			// }

			// client2 := queues.NewMockCuckooClient()
			// machines, err := client2.ListMachinesMock(context.Background())
			// if err != nil {
			// 	slog.Println("ERROR WHILE FETCHING: ", err)
			// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_SANDBOX_TASKS, err)
			// }

			client := cuckoo.New(&cuckoo.Config{})
			machines, err := client.ListMachines(context.Background())
			if err != nil {
				// slog.Println("ERROR WHILE FETCHING: ", err)
				return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_SANDBOX_TASKS, err)
			}

			// for _, machine := range machines {
			// 	if _, ok := machineAndTaskMap[machine.Name]; !ok {
			// 		tasks = append(tasks, map[string]any{
			// 			"machine_name": machine.Name,
			// 			"file_name":    "",
			// 			// "machine_platform": platform,
			// 			"progress":       "NA",
			// 			"machine_status": "poweroff",
			// 		})
			// 	}
			// }

			// for _, machine := range machines {
			// 	tasks = append(tasks, map[string]any{
			// 		"machine_name": machine.Name,
			// 		"file_name":    "",
			// 		// "machine_platform": platform,
			// 		"progress":       "NA",
			// 		"machine_status": "poweroff",
			// 	})
			// }

			// for id := range tasks {
			// 	tasks[id]["machine_platform"] = vmOsMap[id+1]
			// }

			vmOsMap := ReadVmOs()

			sandboxTasks := queues.FetchSandboxTaskCountWhichAreNotReported()

			for i := 1; i <= 8; i++ {

				var status interface{} = "poweroff"
				// var locked bool
				if i <= len(machines) {
					// locked = machines[i-1].Locked
					status = machines[i-1].Status
				}
				platform := "windows7"
				if _, ok := vmOsMap[i]; ok {
					platform = vmOsMap[i]
				}

				// if task count >= no of vms then status = running
				if sandboxTasks >= len(machines) {
					status = "running"
				}

				tasks = append(tasks, map[string]any{
					"machine_name":     fmt.Sprintf("Machine-%d", i),
					"machine_status":   status,
					"machine_platform": platform,
				})
			}

			data = tasks

		case "file-job-search":
			tasks := []map[string]any{}
			for _, fod := range fods {
				if fod.Status == extras.REPORTED || fod.Status == extras.ABORTED || fod.Status == extras.PREVIOUSLY_SCANNED_FILE {
					duration := float64(-1)
					if fod.FinishedTime.Valid {
						duration = fod.FinishedTime.Time.Sub(fod.SubmittedTime).Seconds()
					}

					tasks = append(tasks, map[string]any{
						"job_id":         fod.Id,
						"submitted_time": fod.SubmittedTime.Format(extras.TIME_FORMAT),
						"verdict":        fod.Rating,
						"file_type":      fod.ContentType,
						"file_name":      util.QUnescape(fod.FileName),
						"final_verdict":  fod.FinalVerdict,
						"duration":       duration,
					})
				}
			}

			data = tasks
		case "url-job-search":
			tasks := []map[string]any{}
			for _, uod := range uods {
				if uod.Status == extras.REPORTED || uod.Status == extras.ABORTED || uod.Status == extras.PREVIOUSLY_SCANNED_URL {
					duration := float64(-1)
					if uod.FinishedTime.Valid {
						duration = uod.FinishedTime.Time.Sub(uod.SubmittedTime).Seconds()
					}

					tasks = append(tasks, map[string]any{
						"job_id":         uod.Id,
						"submitted_time": uod.SubmittedTime.Format(extras.TIME_FORMAT),
						"verdict":        uod.Rating,
						"url":            uod.UrlName,
						"final_verdict":  uod.FinalVerdict,
						"duration":       duration,
					})
				}
			}

			data = tasks

		case "overridden-file-search":
			tasks := []map[string]any{}
			for _, fod := range fods {
				if fod.OverriddenVerdict {
					if fod.Status == extras.REPORTED || fod.Status == extras.PREVIOUSLY_SCANNED_FILE {
						duration := float64(-1)
						if fod.FinishedTime.Valid {
							duration = fod.FinishedTime.Time.Sub(fod.SubmittedTime).Seconds()
						}
						tasks = append(tasks, map[string]any{
							"job_id":         fod.Id,
							"submitted_time": fod.SubmittedTime,
							"verdict":        fod.Rating,
							"file_type":      fod.ContentType,
							"file_name":      fod.FileName,
							"final_verdict":  fod.FinalVerdict,
							"duration":       duration,
						})
					}
				}
			}
			data = tasks

		case "overridden-url-search":
			tasks := []map[string]any{}
			for _, uod := range uods {
				if uod.OverriddenVerdict {
					if uod.Status == extras.REPORTED || uod.Status == extras.PREVIOUSLY_SCANNED_URL {
						duration := float64(-1)
						if uod.FinishedTime.Valid {
							duration = uod.FinishedTime.Time.Sub(uod.SubmittedTime).Seconds()
						}
						tasks = append(tasks, map[string]any{
							"job_id":         uod.Id,
							"submitted_time": uod.SubmittedTime,
							"verdict":        uod.Rating,
							"file_type":      "URL",
							"url":            uod.UrlName,
							"final_verdict":  uod.FinalVerdict,
							"duration":       duration,
						})
					}
				}
			}
			data = tasks

		default:
			tasks := []map[string]any{}
			if modelName == extras.FILEONDEMAND {
				for _, fod := range fods {
					tasks = append(tasks, map[string]any{
						"job_id":         fod.Id,
						"file_name":      fod.FileName,
						"comments":       fod.Comments,
						"status":         fod.Status,
						"rating":         fod.Rating,
						"submitted_time": fod.SubmittedTime.Format(extras.TIME_FORMAT),
						"submitted_by":   fod.SubmittedBy,
						"file_count":     fod.FileCount,
					})
				}
			} else {
				for _, uod := range uods {
					tasks = append(tasks, map[string]any{
						"job_id":         uod.Id,
						"url":            uod.UrlName,
						"comments":       uod.Comments,
						"status":         uod.Status,
						"rating":         uod.Rating,
						"submitted_time": uod.SubmittedTime.Format(extras.TIME_FORMAT),
						"submitted_by":   uod.SubmittedBy,
						"url_count":      uod.UrlCount,
					})
				}
			}

			data = tasks
		}

	case extras.LOGREPORT:
		var logReports []any = []any{}

		if action == "task-events" || action == "system-events" || action == "total-events" || action == "vm-events" {
			var logs []string
			logDataInBytes, err := os.ReadFile(extras.DATABASE_PATH + "logFile.txt")
			if err != nil {
				resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
				return resp
			}

			logs = strings.Split(string(logDataInBytes), "\n")
			for _, line := range logs {
				logReport := map[string]any{}
				part := strings.Fields(line)
				if len(part) < 4 {
					continue
				}

				// log time formatting
				if strings.HasPrefix(part[0], "time=") {
					logReport["time_stamp"] = strings.ReplaceAll(strings.TrimPrefix(part[0], "time=")+" "+part[1], `"`, "")
				} else {
					logReport["time_stamp"] = "2006-01-02 15:04:05"
				}

				// log message formatting
				if strings.HasPrefix(part[3], "msg=") {
					eventLog := strings.TrimPrefix(part[3], "msg=")
					for i := range part {
						if i > 3 {
							eventLog += " " + part[i]
						}
					}

					eventLog = strings.ReplaceAll(eventLog, `"`, ``)
					if strings.HasPrefix(eventLog, "taskLog:") && (action == "task-events" || action == "total-events") {
						eventLog = strings.TrimPrefix(eventLog, "taskLog:")
						logReport["event_log"] = eventLog
						logReport["flag"] = "task-events"
						logReports = append(logReports, logReport)

					}
					if strings.HasPrefix(eventLog, "sysLog:") && (action == "system-events" || action == "total-events") {
						eventLog = strings.TrimPrefix(eventLog, "sysLog:")
						logReport["event_log"] = eventLog
						logReport["flag"] = "system-events"
						logReports = append(logReports, logReport)
					}
					if strings.HasPrefix(eventLog, "vmLog:") && (action == "vm-events" || action == "total-events") {
						eventLog = strings.TrimPrefix(eventLog, "vmLog:")
						logReport["event_log"] = eventLog
						logReport["flag"] = "vm-events"
						logReports = append(logReports, logReport)
					}
				}

			}

			data = util.Reverse(logReports)
		}

	case extras.AUDIT_LOGS:
		tasks := []map[string]any{}
		var auditLogs []model.AuditTable
		queryString := "SELECT * FROM audit_tables ORDER BY audit_id DESC"
		dbOprs := dao.DatabaseOperationsRepo{
			QueryExecSet: []string{queryString},
			Result:       &auditLogs,
		}
		err := dao.GormOperations(&dbOprs, config.Db, dao.EXEC)
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

		for _, log := range auditLogs {
			tasks = append(tasks, map[string]any{
				"admin_name": log.AdminName,
				"audit_type": log.AuditType,
				"timestamp":  log.TimeStamp.Format(extras.TIME_FORMAT),
				"message":    log.Message,
			})
		}

		data = tasks
	}

	// 		var urlOnDemands = make(map[string]model.UrlOnDemand)
	// 		urlOnDemands, err = dao.FetchUrlOnDemandProfile(map[string]any{})
	// 		if err != nil {
	// 			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	// 			return resp
	// 		}

	// 		fods := util.SortMap(fileOndemands, true).([]model.FileOnDemand)
	// 		uods := util.SortMap(urlOnDemands, false).([]model.UrlOnDemand)

	// 		if action == "job-queue" {
	// 			var filteredJobQueue = []map[string]any{}
	// 			var inputType string
	// 			for _, fod := range fods {
	// 				if fod.FromDevice {
	// 					inputType = "Device File Check"
	// 				} else {
	// 					inputType = "File Manual Check"
	// 				}

	// 				var task = -1
	// 				for taskID := range fod.TaskID {
	// 					task = taskID
	// 				}

	// 				var finishedTime any = "In Progress"
	// 				if task != -1 {
	// 					if fod.TaskID[task].CompletedOn != nil {
	// 						t, _ := time.Parse(time.RFC1123, fmt.Sprintf("%v", fod.TaskID[task].CompletedOn))
	// 						finishedTime = t.Format("2006-01-02 15:04:05")
	// 					}
	// 				}
	// 				if fod.Status == extras.CLEAN_FOUND_FROM_HASH || fod.Status == extras.MALWARE_FOUND_FROM_HASH {
	// 					finishedTime = fod.SubmittedTime
	// 				}

	// 				var duration any = 0
	// 				if fod.TaskReport != nil {
	// 					duration = fod.TaskReport.Info.Duration
	// 				}
	// 				filteredJobQueue = append(filteredJobQueue, map[string]any{
	// 					"submitted_time": fod.SubmittedTime,
	// 					"file_name":      util.QUnescape(fod.FileName.Filename),
	// 					"input_type":     inputType,
	// 					"file_type":      fod.FileName.Header["Content-Type"],
	// 					"queued":         fod.Status,
	// 					"vm_instance":    fod.OsSupported,
	// 					"duration":       duration,
	// 					"finished_on":    finishedTime,
	// 				})
	// 			}

	// 			for _, uod := range uods {
	// 				if uod.FromDevice {
	// 					inputType = "Device Url Check"
	// 				} else {
	// 					inputType = "Url Manual Check"
	// 				}

	// 				var task = -1
	// 				for taskID := range uod.TaskID {
	// 					task = taskID
	// 				}

	// 				var finishedTime any = "In Progress"
	// 				if task != -1 {
	// 					if uod.TaskID[task].CompletedOn != nil {
	// 						t, _ := time.Parse(time.RFC1123, fmt.Sprintf("%v", uod.TaskID[task].CompletedOn))
	// 						finishedTime = t.Format("2006-01-02 15:04:05")
	// 					}
	// 				}
	// 				if uod.Status == extras.CLEAN_FOUND_FROM_CACHE || uod.Status == extras.MALWARE_FOUND_FROM_CACHE {
	// 					finishedTime = uod.SubmittedTime
	// 				}

	// 				var duration any = 0
	// 				if uod.TaskReport != nil {
	// 					duration = uod.TaskReport.Info.Duration
	// 				}
	// 				filteredJobQueue = append(filteredJobQueue, map[string]any{
	// 					"submitted_time": uod.SubmittedTime,
	// 					"url":            util.QUnescape(uod.UrlName),
	// 					"input_type":     inputType,
	// 					"queued":         uod.Status,
	// 					"vm_instance":    uod.OsSupported,
	// 					"duration":       duration,
	// 					"finished_on":    finishedTime,
	// 				})
	// 			}

	// 			sort.Slice(filteredJobQueue, func(i, j int) bool {
	// 				t1, _ := time.Parse(extras.TIME_FORMAT, filteredJobQueue[i]["submitted_time"].(string))
	// 				t2, _ := time.Parse(extras.TIME_FORMAT, filteredJobQueue[j]["submitted_time"].(string))
	// 				return t1.After(t2)
	// 			})

	// 			data = filteredJobQueue

	// 		} else if action == "vm-job" {
	// 			var filteredVMJobs = map[string]any{}
	// 			client := cuckoo.New(&cuckoo.Config{})
	// 			allMachines, err := client.ListMachines(context.Background())
	// 			if err != nil {
	// 				resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	// 				return resp
	// 			}

	// 			sort.Slice(allMachines, func(i, j int) bool {
	// 				ip1 := allMachines[i].Name
	// 				ip2 := allMachines[j].Name
	// 				return util.CompareIPs(ip1, ip2) < 0
	// 			})

	// 			platform, err := os.ReadFile(extras.PLATFORM_FILE_NAME)
	// 			if err != nil {
	// 				platform = []byte("windows7")
	// 			}

	// 			for i := range allMachines {
	// 				allMachines[i].Platform = string(platform)
	// 			}

	// 			machineIdx := 0
	// 			for _, job := range fods {
	// 				if machineIdx >= 8 {
	// 					break
	// 				}
	// 				if job.Status == string(model.StatusRunning) {
	// 					allMachines[machineIdx].Platform = string(platform)
	// 					if allMachines[machineIdx].Status == "aborted" {
	// 						allMachines[machineIdx].Status = "poweroff"
	// 					}
	// 					if !allMachines[machineIdx].Locked {
	// 						filteredVMJobs[allMachines[machineIdx].Name] = map[string]any{
	// 							"machine_name":     allMachines[machineIdx].Name,
	// 							"file_name":        "",
	// 							"machine_platform": allMachines[machineIdx].Platform,
	// 							"progress":         "N/A",
	// 							"machine_status":   allMachines[machineIdx].Status,
	// 						}
	// 					} else {
	// 						filteredVMJobs[allMachines[machineIdx].Name] = map[string]any{
	// 							"machine_name":     allMachines[machineIdx].Name,
	// 							"file_name":        util.QUnescape(job.FileName.Filename),
	// 							"machine_platform": allMachines[machineIdx].Platform,
	// 							"progress":         job.Status,
	// 							"machine_status":   "running",
	// 						}
	// 					}
	// 					machineIdx++
	// 				}
	// 			}

	// 			for _, job := range uods {
	// 				if machineIdx >= 8 {
	// 					break
	// 				}
	// 				if job.Status == string(model.StatusRunning) {
	// 					allMachines[machineIdx].Platform = string(platform)
	// 					if allMachines[machineIdx].Status == "aborted" {
	// 						allMachines[machineIdx].Status = "poweroff"
	// 					}
	// 					if !allMachines[machineIdx].Locked {
	// 						filteredVMJobs[allMachines[machineIdx].Name] = map[string]any{
	// 							"machine_name":     allMachines[machineIdx].Name,
	// 							"file_name":        "",
	// 							"machine_platform": allMachines[machineIdx].Platform,
	// 							"progress":         "N/A",
	// 							"machine_status":   allMachines[machineIdx].Status,
	// 						}
	// 					} else {
	// 						filteredVMJobs[allMachines[machineIdx].Name] = map[string]any{
	// 							"machine_name":     allMachines[machineIdx].Name,
	// 							"file_name":        util.QUnescape(job.UrlName),
	// 							"machine_platform": allMachines[machineIdx].Platform,
	// 							"progress":         job.Status,
	// 							"machine_status":   "running",
	// 						}
	// 					}
	// 					machineIdx++
	// 				}
	// 			}

	// 			if machineIdx < 8 {
	// 				for i := machineIdx; i < 8; i++ {
	// 					filteredVMJobs[allMachines[i].Name] = map[string]any{
	// 						"machine_name":     allMachines[i].Name,
	// 						"file_name":        "",
	// 						"machine_platform": allMachines[i].Platform,
	// 						"progress":         "N/A",
	// 						"machine_status":   allMachines[i].Status,
	// 					}
	// 				}
	// 			}

	// 			var arrOfFilteredVMJobs []any = []any{}

	// 			for key := range filteredVMJobs {
	// 				arrOfFilteredVMJobs = append(arrOfFilteredVMJobs, filteredVMJobs[key])
	// 			}

	// 			sort.Slice(arrOfFilteredVMJobs, func(i, j int) bool {
	// 				ip1 := arrOfFilteredVMJobs[i].(map[string]interface{})["machine_name"].(string)
	// 				ip2 := arrOfFilteredVMJobs[j].(map[string]interface{})["machine_name"].(string)
	// 				return util.CompareIPs(ip1, ip2) < 0
	// 			})

	// 			data = arrOfFilteredVMJobs

	// 		} else if action == "file-job-search" {
	// 			var filteredFileJobSearch = []map[string]any{}
	// 			for _, fod := range fods {
	// 				if fod.Status == extras.REPORTED {
	// 					filteredFileJobSearch = append(filteredFileJobSearch, map[string]any{
	// 						"job_id":         fod.JobId,
	// 						"submitted_time": fod.SubmittedTime,
	// 						"verdict":        fod.Rating,
	// 						"file_type":      fod.FileName.Header["Content-Type"],
	// 						"file_name":      util.QUnescape(fod.FileName.Filename),
	// 						"final_verdict":  fod.FinalVerdict,
	// 						"duration":       fod.TaskReport.Info.Duration,
	// 					})
	// 				} else if fod.Status == extras.MALWARE_FOUND_FROM_HASH {
	// 					filteredFileJobSearch = append(filteredFileJobSearch, map[string]any{
	// 						"job_id":         fod.JobId,
	// 						"submitted_time": fod.SubmittedTime,
	// 						"verdict":        "Malicious",
	// 						"file_type":      fod.FileName.Header["Content-Type"],
	// 						"file_name":      util.QUnescape(fod.FileName.Filename),
	// 						"final_verdict":  fod.FinalVerdict,
	// 						"duration":       1,
	// 					})
	// 				} else if fod.Status == extras.CLEAN_FOUND_FROM_HASH {
	// 					filteredFileJobSearch = append(filteredFileJobSearch, map[string]any{
	// 						"job_id":         fod.JobId,
	// 						"submitted_time": fod.SubmittedTime,
	// 						"verdict":        "Clean",
	// 						"file_type":      fod.FileName.Header["Content-Type"],
	// 						"file_name":      util.QUnescape(fod.FileName.Filename),
	// 						"final_verdict":  fod.FinalVerdict,
	// 						"duration":       1,
	// 					})
	// 				}
	// 			}

	// 			sort.Slice(filteredFileJobSearch, func(i, j int) bool {
	// 				t1, _ := time.Parse(extras.TIME_FORMAT, filteredFileJobSearch[i]["submitted_time"].(string))
	// 				t2, _ := time.Parse(extras.TIME_FORMAT, filteredFileJobSearch[j]["submitted_time"].(string))
	// 				return t1.After(t2)
	// 			})

	// 			data = filteredFileJobSearch
	// 		} else if action == "overridden-file-search" {
	// 			var filteredFileJobSearch = []map[string]any{}
	// 			for _, fod := range fods {
	// 				if fod.IsFinalVerdictChanged {
	// 					if fod.Status == extras.REPORTED {
	// 						filteredFileJobSearch = append(filteredFileJobSearch, map[string]any{
	// 							"job_id":         fod.JobId,
	// 							"submitted_time": fod.SubmittedTime,
	// 							"verdict":        fod.Rating,
	// 							"file_type":      fod.FileName.Header["Content-Type"],
	// 							"file_name":      util.QUnescape(fod.FileName.Filename),
	// 							"final_verdict":  fod.FinalVerdict,
	// 							"duration":       fod.TaskReport.Info.Duration,
	// 						})
	// 					} else if fod.Status == extras.MALWARE_FOUND_FROM_HASH {
	// 						filteredFileJobSearch = append(filteredFileJobSearch, map[string]any{
	// 							"job_id":         fod.JobId,
	// 							"submitted_time": fod.SubmittedTime,
	// 							"verdict":        "Malicious",
	// 							"file_type":      fod.FileName.Header["Content-Type"],
	// 							"file_name":      util.QUnescape(fod.FileName.Filename),
	// 							"final_verdict":  fod.FinalVerdict,
	// 							"duration":       1,
	// 						})
	// 					} else if fod.Status == extras.CLEAN_FOUND_FROM_HASH {
	// 						filteredFileJobSearch = append(filteredFileJobSearch, map[string]any{
	// 							"job_id":         fod.JobId,
	// 							"submitted_time": fod.SubmittedTime,
	// 							"verdict":        "Clean",
	// 							"file_type":      fod.FileName.Header["Content-Type"],
	// 							"file_name":      util.QUnescape(fod.FileName.Filename),
	// 							"final_verdict":  fod.FinalVerdict,
	// 							"duration":       1,
	// 						})
	// 					}
	// 				}
	// 			}

	// 			sort.Slice(filteredFileJobSearch, func(i, j int) bool {
	// 				t1, _ := time.Parse(extras.TIME_FORMAT, filteredFileJobSearch[i]["submitted_time"].(string))
	// 				t2, _ := time.Parse(extras.TIME_FORMAT, filteredFileJobSearch[j]["submitted_time"].(string))
	// 				return t1.After(t2)
	// 			})

	// 			data = filteredFileJobSearch

	// 		} else {
	// 			var filteredFileOndemand = []map[string]any{}
	// 			for _, fod := range fods {
	// 				// t1, _ := time.Parse(extras.TIME_FORMAT, fod.SubmittedTime)
	// 				filteredFileOndemand = append(filteredFileOndemand, map[string]any{
	// 					"file_name":      util.QUnescape(fod.FileName.Filename),
	// 					"comments":       fod.Comments,
	// 					"status":         fod.Status,
	// 					"rating":         fod.Rating,
	// 					"submitted_time": fod.SubmittedTime,
	// 					"submitted_by":   fod.SubmittedBy,
	// 					"file_count":     fod.FileCount,
	// 				})
	// 			}

	// 			sort.Slice(filteredFileOndemand, func(i, j int) bool {
	// 				t1, _ := time.Parse(extras.TIME_FORMAT, filteredFileOndemand[i]["submitted_time"].(string))
	// 				t2, _ := time.Parse(extras.TIME_FORMAT, filteredFileOndemand[j]["submitted_time"].(string))
	// 				return t1.After(t2)
	// 			})

	// 			data = filteredFileOndemand
	// 		}

	// 	case extras.URLONDEMAND:
	// 		var urlOndemands = make(map[string]model.UrlOnDemand)
	// 		urlOndemands, err = dao.FetchUrlOnDemandProfile(map[string]any{})
	// 		if err != nil {
	// 			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	// 			return resp
	// 		}

	// 		uods := util.SortMap(urlOndemands, false).([]model.UrlOnDemand)

	// 		if action == "url-job-search" {
	// 			var filteredUrlJobSearch = []map[string]any{}
	// 			for _, uod := range uods {
	// 				if uod.Status == extras.REPORTED {
	// 					filteredUrlJobSearch = append(filteredUrlJobSearch, map[string]any{
	// 						"job_id":         uod.JobId,
	// 						"submitted_time": uod.SubmittedTime,
	// 						"verdict":        uod.Rating,
	// 						"url":            util.QUnescape(uod.UrlName),
	// 						"final_verdict":  uod.FinalVerdict,
	// 						"duration":       uod.TaskReport.Info.Duration,
	// 					})
	// 				} else if uod.Status == extras.MALWARE_FOUND_FROM_CACHE {
	// 					filteredUrlJobSearch = append(filteredUrlJobSearch, map[string]any{
	// 						"job_id":         uod.JobId,
	// 						"submitted_time": uod.SubmittedTime,
	// 						"verdict":        "Malicious",
	// 						"url":            util.QUnescape(uod.UrlName),
	// 						"final_verdict":  uod.FinalVerdict,
	// 						"duration":       1,
	// 					})
	// 				} else if uod.Status == extras.CLEAN_FOUND_FROM_CACHE {
	// 					filteredUrlJobSearch = append(filteredUrlJobSearch, map[string]any{
	// 						"job_id":         uod.JobId,
	// 						"submitted_time": uod.SubmittedTime,
	// 						"verdict":        "Clean",
	// 						"url":            util.QUnescape(uod.UrlName),
	// 						"final_verdict":  uod.FinalVerdict,
	// 						"duration":       1,
	// 					})
	// 				}
	// 			}

	// 			sort.Slice(filteredUrlJobSearch, func(i, j int) bool {
	// 				t1, _ := time.Parse(extras.TIME_FORMAT, filteredUrlJobSearch[i]["submitted_time"].(string))
	// 				t2, _ := time.Parse(extras.TIME_FORMAT, filteredUrlJobSearch[j]["submitted_time"].(string))
	// 				return t1.After(t2)
	// 			})

	// 			data = filteredUrlJobSearch
	// 		} else if action == "overridden-url-search" {
	// 			var filteredUrlJobSearch = []map[string]any{}
	// 			for _, uod := range uods {
	// 				if uod.IsFinalVerdictChanged {
	// 					if uod.Status == extras.REPORTED {
	// 						filteredUrlJobSearch = append(filteredUrlJobSearch, map[string]any{
	// 							"job_id":         uod.JobId,
	// 							"submitted_time": uod.SubmittedTime,
	// 							"verdict":        uod.Rating,
	// 							"url":            util.QUnescape(uod.UrlName),
	// 							"final_verdict":  uod.FinalVerdict,
	// 							"duration":       uod.TaskReport.Info.Duration,
	// 						})
	// 					} else if uod.Status == extras.MALWARE_FOUND_FROM_CACHE {
	// 						filteredUrlJobSearch = append(filteredUrlJobSearch, map[string]any{
	// 							"job_id":         uod.JobId,
	// 							"submitted_time": uod.SubmittedTime,
	// 							"verdict":        "Malicious",
	// 							"url":            util.QUnescape(uod.UrlName),
	// 							"final_verdict":  uod.FinalVerdict,
	// 							"duration":       1,
	// 						})
	// 					} else if uod.Status == extras.CLEAN_FOUND_FROM_CACHE {
	// 						filteredUrlJobSearch = append(filteredUrlJobSearch, map[string]any{
	// 							"job_id":         uod.JobId,
	// 							"submitted_time": uod.SubmittedTime,
	// 							"verdict":        "Clean",
	// 							"url":            util.QUnescape(uod.UrlName),
	// 							"final_verdict":  uod.FinalVerdict,
	// 							"duration":       1,
	// 						})
	// 					}
	// 				}
	// 			}

	// 			sort.Slice(filteredUrlJobSearch, func(i, j int) bool {
	// 				t1, _ := time.Parse(extras.TIME_FORMAT, filteredUrlJobSearch[i]["submitted_time"].(string))
	// 				t2, _ := time.Parse(extras.TIME_FORMAT, filteredUrlJobSearch[j]["submitted_time"].(string))
	// 				return t1.After(t2)
	// 			})

	// 			data = filteredUrlJobSearch

	// 		} else {
	// 			var filteredUrlOndemand = []map[string]any{}
	// 			for _, uod := range uods {
	// 				// t1, _ := time.Parse("2006-01-02 15:04:05", uod.SubmittedTime)
	// 				filteredUrlOndemand = append(filteredUrlOndemand, map[string]any{
	// 					"url":            util.QUnescape(uod.UrlName),
	// 					"comments":       uod.Comments,
	// 					"status":         uod.Status,
	// 					"rating":         uod.Rating,
	// 					"submitted_time": uod.SubmittedTime,
	// 					"submitted_by":   uod.SubmittedBy,
	// 					"url_count":      uod.UrlCount,
	// 				})
	// 			}

	// 			sort.Slice(filteredUrlOndemand, func(i, j int) bool {
	// 				t1, _ := time.Parse(extras.TIME_FORMAT, filteredUrlOndemand[i]["submitted_time"].(string))
	// 				t2, _ := time.Parse(extras.TIME_FORMAT, filteredUrlOndemand[j]["submitted_time"].(string))
	// 				return t1.After(t2)
	// 			})

	// 			data = filteredUrlOndemand
	// 		}

	// LOG REPORT PART

	// 		if action == "vm-events" || action == "total-events" {
	// 			var vmLogReports []any
	// 			fileOnDemands, err = dao.FetchFileOnDemandProfile(map[string]any{})
	// 			if err != nil {
	// 				resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	// 				return resp
	// 			}

	// 			urlOnDemands, err = dao.FetchUrlOnDemandProfile(map[string]any{})
	// 			if err != nil {
	// 				resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	// 				return resp
	// 			}

	// 			for id, job := range fileOnDemands {
	// 				if job.Status == extras.REPORTED {
	// 					for _, log := range job.TaskReport.Debug.Cuckoo.([]any) {
	// 						if len(strings.Split(log.(string), ",")) < 2 {
	// 							continue
	// 						}

	// 						_, err := time.Parse("2006-01-02 15:04:05", strings.Split(log.(string), ",")[0])
	// 						if err != nil {
	// 							continue
	// 						}

	// 						vmLogReports = append(vmLogReports, map[string]any{
	// 							"event_log":  strings.Join(strings.Split(log.(string), ",")[1:], ","),
	// 							"time_stamp": strings.Split(log.(string), ",")[0],
	// 							"name":       util.QUnescape(job.FileName.Filename),
	// 							"job_id":     id,
	// 							"flag":       "vm-events",
	// 						})
	// 					}
	// 				}
	// 			}

	// 			for id, job := range urlOnDemands {
	// 				if job.Status == extras.REPORTED {
	// 					for _, log := range job.TaskReport.Debug.Cuckoo.([]any) {
	// 						if len(strings.Split(log.(string), ",")) < 2 {
	// 							continue
	// 						}

	// 						_, err := time.Parse("2006-01-02 15:04:05", strings.Split(log.(string), ",")[0])
	// 						if err != nil {
	// 							continue
	// 						}

	// 						vmLogReports = append(vmLogReports, map[string]any{
	// 							"event_log":  strings.Join(strings.Split(log.(string), ",")[1:], ","),
	// 							"time_stamp": strings.Split(log.(string), ",")[0],
	// 							"name":       util.QUnescape(job.UrlName),
	// 							"job_id":     id,
	// 							"flag":       "vm-events",
	// 						})

	// 					}
	// 				}
	// 			}

	// 			var newLogReports []any

	// 			for _, logs := range vmLogReports {
	// 				cuckooLog := strings.Split(logs.(map[string]any)["event_log"].(string), " ")
	// 				if containsComponent(cuckooLog, "[cuckoo.core.plugins]", "[cuckoo.core.scheduler]", "[cuckoo.core.resultserver]", "[cuckoo.auxiliary.sniffer]", "[cuckoo.machinery.virtualbox]", "[cuckoo.core.guest]") {
	// 					if len(cuckooLog) > 3 {
	// 						cuckooLog = cuckooLog[3:]
	// 					}

	// 					for i := range cuckooLog {
	// 						if strings.Contains(cuckooLog[i], "#") {
	// 							cuckooLog[i] = "#" + logs.(map[string]any)["job_id"].(string) + ":"
	// 						}
	// 					}

	// 					cuckooLog = append(cuckooLog, "name="+logs.(map[string]any)["name"].(string))
	// 					logs.(map[string]any)["event_log"] = strings.Join(cuckooLog, " ")
	// 					newLogReports = append(newLogReports, logs)
	// 				} else {
	// 					newLogReports = append(newLogReports, logs)
	// 				}
	// 				delete(logs.(map[string]any), "name")
	// 				delete(logs.(map[string]any), "job_id")
	// 			}

	// 			vmLogReports = newLogReports

	// 			logReports = append(logReports, vmLogReports...)
	// 		}

	// 		sort.Slice(logReports, func(i, j int) bool {
	// 			ti, _ := time.Parse("2006-01-02 15:04:05", logReports[i].(map[string]any)["time_stamp"].(string))
	// 			tj, _ := time.Parse("2006-01-02 15:04:05", logReports[j].(map[string]any)["time_stamp"].(string))
	// 			return ti.After(tj)
	// 		})
	// 		data = logReports

	// 	default:
	// 		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_MODEL_NAME_INVALID, extras.ErrModelNameInvalid)
	// 		return resp
	// 	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, data)
	return resp
}

func FetchAllFODs() ([]model.FileOnDemand, error) {
	var fods []model.FileOnDemand
	queryString := "SELECT file_on_demands.id, file_on_demands.file_name, file_on_demands.content_type, file_on_demands.submitted_time, file_on_demands.submitted_by, file_on_demands.file_count, file_on_demands.rating, file_on_demands.score, file_on_demands.comments, file_on_demands.overridden_verdict, file_on_demands.overridden_by, file_on_demands.final_verdict, file_on_demands.from_device, file_on_demands.finished_time, file_on_demands.status FROM file_on_demands WHERE status != '' AND status IS NOT NULL"
	fodRepo := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &fods,
	}

	err := dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
	if err != nil {
		// slog.Println("ERROR FROM DATABASE: ", err)
		return []model.FileOnDemand{}, err
	}

	var fods1 []model.FileOnDemand
	queryString = "SELECT file_on_demands.id, file_on_demands.file_name, file_on_demands.content_type, file_on_demands.submitted_time, file_on_demands.submitted_by, file_on_demands.file_count, file_on_demands.rating, file_on_demands.score, file_on_demands.comments, file_on_demands.overridden_verdict, file_on_demands.overridden_by, file_on_demands.final_verdict, file_on_demands.from_device, file_on_demands.finished_time, CASE WHEN file_on_demands.status = '' OR file_on_demands.status IS NULL THEN task_live_analysis_tables.status ELSE file_on_demands.status END AS status FROM file_on_demands INNER JOIN task_live_analysis_tables ON task_live_analysis_tables.id = file_on_demands.id"
	fodRepo = dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &fods1,
	}
	err = dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
	if err != nil {
		// slog.Println("ERROR FROM DATABASE: ", err)
		return []model.FileOnDemand{}, err
	}

	queryString = "SELECT file_on_demands.id, file_on_demands.file_name, file_on_demands.content_type, file_on_demands.submitted_time, file_on_demands.submitted_by, file_on_demands.file_count, file_on_demands.rating, file_on_demands.score, file_on_demands.comments, file_on_demands.overridden_verdict, file_on_demands.overridden_by, file_on_demands.final_verdict, file_on_demands.from_device, file_on_demands.finished_time, 'reported' as status FROM file_on_demands INNER JOIN task_finished_tables ON task_finished_tables.id = file_on_demands.id"
	fods2 := []model.FileOnDemand{}
	fodRepo = dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &fods2,
	}
	err = dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
	if err != nil {
		// slog.Println("ERROR FROM DATABASE: ", err)
		return []model.FileOnDemand{}, err
	}

	queryString = "SELECT file_on_demands.id, file_on_demands.file_name, file_on_demands.content_type, file_on_demands.submitted_time, file_on_demands.submitted_by, file_on_demands.file_count, file_on_demands.rating, file_on_demands.score, file_on_demands.comments, file_on_demands.overridden_verdict, file_on_demands.overridden_by, file_on_demands.final_verdict, file_on_demands.from_device, file_on_demands.finished_time, 'already processing the same file, waiting for its turn' as status FROM file_on_demands INNER JOIN task_duplicate_tables ON task_duplicate_tables.id = file_on_demands.id"
	fods3 := []model.FileOnDemand{}
	fodRepo = dao.DatabaseOperationsRepo{
		QueryExecSet: []string{queryString},
		Result:       &fods3,
	}
	err = dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
	if err != nil {
		// slog.Println("ERROR FROM DATABASE: ", err)
		return []model.FileOnDemand{}, err
	}

	fods = append(fods, fods1...)
	fods = append(fods, fods2...)
	fods = append(fods, fods3...)

	sort.Slice(fods, func(i int, j int) bool {
		return fods[i].Id > fods[j].Id
	})

	return fods, nil
}

func FetchAllUODs() ([]model.UrlOnDemand, error) {
	var uods []model.UrlOnDemand
	uodRepo := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{"SELECT * FROM url_on_demands ORDER BY id DESC"},
		Result:       &uods,
	}
	err := dao.GormOperations(&uodRepo, config.Db, dao.EXEC)
	if err != nil {
		// slog.Println("ERROR IN FETCHING: ", err)
		return []model.UrlOnDemand{}, err
	}

	return uods, nil
}

func ReadVmOs() map[int]string {
	vmOsMap := make(map[int]string)
	data, err := os.ReadFile(extras.PLATFORM_FILE_NAME)
	if os.IsNotExist(err) || len(data) == 0 || err != nil {
		data = []byte("")
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		vmNo, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		vmOsMap[vmNo] = parts[1]
	}

	for i := 1; i <= 8; i++ {
		if _, ok := vmOsMap[i]; !ok {
			vmOsMap[i] = "windows7"
		}
	}

	return vmOsMap
}

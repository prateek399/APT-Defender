package service

import (
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"bufio"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/host"
)

// import (
// 	"anti-apt-backend/dao"
// 	"anti-apt-backend/extras"
// 	"anti-apt-backend/model"
// 	"anti-apt-backend/util"
// 	"bufio"
// 	"errors"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/exec"
// 	"sort"
// 	"strings"
// 	"time"

// 	"github.com/shirou/gopsutil/host"
// )

type ScannedCounter struct {
	StartTime string `json:"start_time"`
	Count     int    `json:"count"`
}

type Statistics struct {
	Pending    int `json:"pending"`
	Processing int `json:"running"`
	Low        int `json:"low"`
	Medium     int `json:"medium"`
	High       int `json:"high"`
	Critical   int `json:"critical"`
	Safe       int `json:"safe"`
	Total      int `json:"total"`
}

type ScannedPerformance struct {
	ScannedUrlPercent  int64 `json:"scanned_url_percent"`
	ScannedFilePercent int64 `json:"scanned_file_percent"`
}

type HdwrUsage struct {
	RAM  float64 `json:"ram"`
	CPU  float64 `json:"cpu"`
	Disk float64 `json:"disk"`
	Time string  `json:"time"`
}

func Dashboard(action string) model.APIResponse {
	switch action {
	case "scanned-count-per-hour":
		return ScannedCount(time.Hour)
	case "scan-statistics":
		return ScanStatistics(24*time.Hour, true)
	case "total-malwares":
		return TotalMalwares(24*time.Hour, "", true)
	case "url-malwares":
		return TotalMalwares(24*time.Hour, extras.URLONDEMAND, true)
	case "file-malwares":
		return TotalMalwares(24*time.Hour, extras.FILEONDEMAND, true)
	case "scan-performance":
		return TotalScannedTasks(24*time.Hour, true)
	case "system-events":
		return EventCount(24*time.Hour, action, true)
	case "task-events":
		return EventCount(24*time.Hour, action, true)
	case "vm-events":
		return EventCount(24*time.Hour, action, true)
	case "total-events":
		return EventCount(24*time.Hour, action, true)
	case "hdwr-usage":
		return GetHdwrData()
	case "get-device":
		return GetDevice()
	}

	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_ACTION_TYPE, extras.ErrInvalidActionType)
}

func ScannedCount(duration time.Duration) model.APIResponse {
	var counter []ScannedCounter
	var fods []model.FileOnDemand
	var uods []model.UrlOnDemand

	fodRepo := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{"SELECT file_on_demands.id, file_on_demands.file_name, file_on_demands.content_type, file_on_demands.finished_time, file_on_demands.submitted_time, file_on_demands.submitted_by, file_on_demands.file_count, file_on_demands.rating, file_on_demands.score, file_on_demands.comments, file_on_demands.overridden_verdict, file_on_demands.overridden_by, file_on_demands.final_verdict, file_on_demands.from_device, 'reported' as status FROM file_on_demands WHERE status != '' OR status IS NOT NULL ORDER BY file_on_demands.id DESC"},
		Result:       &fods,
	}
	err := dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
	if err != nil {
		// slog.Println("ERROR FROM DATABASE: ", err)
	}

	fods1 := []model.FileOnDemand{}
	fodRepo = dao.DatabaseOperationsRepo{
		QueryExecSet: []string{"SELECT file_on_demands.id, file_on_demands.file_name, file_on_demands.content_type, file_on_demands.finished_time, file_on_demands.submitted_time, file_on_demands.submitted_by, file_on_demands.file_count, file_on_demands.rating, file_on_demands.score, file_on_demands.comments, file_on_demands.overridden_verdict, file_on_demands.overridden_by, file_on_demands.final_verdict, file_on_demands.from_device, 'reported' as status FROM file_on_demands INNER JOIN task_finished_tables ON task_finished_tables.id = file_on_demands.id ORDER BY file_on_demands.id DESC"},
		Result:       &fods1,
	}
	err = dao.GormOperations(&fodRepo, config.Db, dao.EXEC)
	if err != nil {
		// slog.Println("ERROR FROM DATABASE: ", err)
	}

	fods = append(fods, fods1...)

	uodRepo := dao.DatabaseOperationsRepo{
		QueryExecSet: []string{"SELECT * FROM url_on_demands"},
		Result:       &uods,
	}
	err = dao.GormOperations(&uodRepo, config.Db, dao.EXEC)
	if err != nil {
		// slog.Println("ERROR FROM DATABASE: ", err)
	}

	var tasks []map[string]any
	for _, fod := range fods {
		if fod.FinishedTime.Valid {
			tasks = append(tasks, map[string]any{
				"time":  fod.FinishedTime.Time,
				"count": fod.FileCount,
			})
		}
	}

	// slog.Println("Tasks: ", tasks)

	for _, uod := range uods {
		if uod.Status != extras.REPORTED && uod.Status != extras.PREVIOUSLY_SCANNED_URL {
			continue
		}

		if uod.FinishedTime.Valid {
			tasks = append(tasks, map[string]any{
				"time":  uod.FinishedTime.Time,
				"count": uod.UrlCount,
			})
		}
	}

	sort.Slice(tasks, func(i, j int) bool {
		t1 := tasks[i]["time"].(time.Time)
		t2 := tasks[j]["time"].(time.Time)
		return t1.After(t2)
	})

	startTime := time.Now()
	endTime := startTime.Add(-1 * duration)
	count := ScannedCounter{
		StartTime: endTime.Format(extras.TIME_FORMAT),
		Count:     0,
	}

	for _, task := range tasks {
		t1 := task["time"].(time.Time)
		if t1.After(endTime) {
			count.Count++
		} else {
			for {
				counter = append(counter, count)
				startTime = endTime
				endTime = startTime.Add(-1 * duration)
				count = ScannedCounter{
					StartTime: endTime.Format(extras.TIME_FORMAT),
					Count:     0,
				}
				if t1.After(endTime) {
					break
				}
			}
			count.Count++
		}
	}

	if count.Count > 0 {
		counter = append(counter, count)
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, counter)
}

func ScanStatistics(duration time.Duration, infinite bool) model.APIResponse {
	statistics := make(map[string]Statistics)
	statistics["device_file_check"] = Statistics{}
	statistics["device_url_check"] = Statistics{}
	statistics["manual_file_check"] = Statistics{}
	statistics["manual_url_check"] = Statistics{}

	fods, _ := FetchAllFODs()

	uods, _ := FetchAllUODs()

	type Jobs struct {
		Type       string
		Time       time.Time
		Rating     string
		Status     string
		FromDevice bool
	}
	var jobs []Jobs
	for _, fod := range fods {
		jobs = append(jobs, Jobs{
			Type:       extras.FILEONDEMAND,
			Time:       fod.SubmittedTime,
			Rating:     fod.Rating,
			Status:     fod.Status,
			FromDevice: fod.FromDevice,
		})
	}
	for _, uod := range uods {
		jobs = append(jobs, Jobs{
			Type:       extras.URLONDEMAND,
			Time:       uod.SubmittedTime,
			Rating:     uod.Rating,
			Status:     uod.Status,
			FromDevice: uod.FromDevice,
		})
	}

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].Time.After(jobs[j].Time)
	})

	endTime := time.Now().Add(-1 * duration)
	for _, job := range jobs {
		stat := Statistics{}
		incomingMode := ""
		if job.FromDevice {
			if job.Type == extras.FILEONDEMAND {
				incomingMode = "device_file_check"
			} else {
				incomingMode = "device_url_check"
			}
		} else {
			if job.Type == extras.FILEONDEMAND {
				incomingMode = "manual_file_check"
			} else {
				incomingMode = "manual_url_check"
			}
		}

		stat = statistics[incomingMode]
		switch job.Rating {
		case "Low":
			stat.Low++
		case "Medium":
			stat.Medium++
		case "High":
			stat.High++
		case "Critical":
			stat.Critical++
		case "malicious":
			stat.Critical++
		case "Clean":
			stat.Safe++
		default:
			if job.Status == extras.PendingInQueue || job.Status == extras.PendingNotInQueue || job.Status == extras.RunningNotInQueue {
				stat.Pending++
			}
			if job.Status == extras.RunningInQueue {
				stat.Processing++
			}
		}

		stat.Total = (stat.High + stat.Low + stat.Critical + stat.Safe + stat.Medium + stat.Pending + stat.Processing)
		statistics[incomingMode] = stat
		if !endTime.Before(job.Time) && !infinite {
			break
		}
	}

	statistics["all_sources"] = Statistics{
		Low:        statistics["device_url_check"].Low + statistics["manual_url_check"].Low + statistics["device_file_check"].Low + statistics["manual_file_check"].Low,
		Medium:     statistics["device_url_check"].Medium + statistics["manual_url_check"].Medium + statistics["device_file_check"].Medium + statistics["manual_file_check"].Medium,
		High:       statistics["device_url_check"].High + statistics["manual_url_check"].High + statistics["device_file_check"].High + statistics["manual_file_check"].High,
		Critical:   statistics["device_url_check"].Critical + statistics["manual_url_check"].Critical + statistics["device_file_check"].Critical + statistics["manual_file_check"].Critical,
		Safe:       statistics["device_url_check"].Safe + statistics["manual_url_check"].Safe + statistics["device_file_check"].Safe + statistics["manual_file_check"].Safe,
		Pending:    statistics["device_url_check"].Pending + statistics["manual_url_check"].Pending + statistics["device_file_check"].Pending + statistics["manual_file_check"].Pending,
		Processing: statistics["device_url_check"].Processing + statistics["manual_url_check"].Processing + statistics["device_file_check"].Processing + statistics["manual_file_check"].Processing,
		Total:      statistics["device_url_check"].Total + statistics["manual_url_check"].Total + statistics["device_file_check"].Total + statistics["manual_file_check"].Total,
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, statistics)
}

func TotalMalwares(duration time.Duration, taskType string, infinite bool) model.APIResponse {
	var malwareCount int64
	resp := ScanStatistics(duration, infinite)
	if _, ok := resp.Data.(map[string]Statistics); !ok {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_DATA_TYPE, extras.ErrInvalidDataType)
	}

	for tp, stat := range resp.Data.(map[string]Statistics) {
		if tp == "all_sources" {
			continue
		}
		switch taskType {
		case extras.URLONDEMAND:
			if tp == "device_url_check" || tp == "manual_url_check" {
				malwareCount += int64(stat.High + stat.Critical + stat.Medium + stat.Low)
			}
		case extras.FILEONDEMAND:
			if tp == "device_file_check" || tp == "manual_file_check" {
				malwareCount += int64(stat.High + stat.Critical + stat.Medium + stat.Low)
			}
		default:
			malwareCount += int64(stat.High + stat.Critical + stat.Medium + stat.Low)
		}
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, malwareCount)
}

func TotalScannedTasks(duration time.Duration, infinite bool) model.APIResponse {
	var totalScannedTaskCount int64 = 0
	var urlTaskCount int64 = 0
	var fileTaskCount int64 = 0

	resp := ScanStatistics(duration, infinite)
	if _, ok := resp.Data.(map[string]Statistics); !ok {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_DATA_TYPE, extras.ErrInvalidDataType)
	}

	for tp, stat := range resp.Data.(map[string]Statistics) {
		if tp == "all_sources" {
			continue
		}

		if tp == "device_url_check" || tp == "manual_url_check" {
			totalScannedTaskCount += int64(stat.Critical + stat.High + stat.Low + stat.Medium + stat.Safe)
			urlTaskCount += int64(stat.Critical + stat.High + stat.Low + stat.Medium + stat.Safe)
		}

		if tp == "device_file_check" || tp == "manual_file_check" {
			totalScannedTaskCount += int64(stat.Critical + stat.High + stat.Low + stat.Medium + stat.Safe)
			fileTaskCount += int64(stat.Critical + stat.High + stat.Low + stat.Medium + stat.Safe)
		}

	}

	if totalScannedTaskCount == 0 {
		scannedPerformance := ScannedPerformance{
			ScannedUrlPercent:  0,
			ScannedFilePercent: 0,
		}
		return model.NewSuccessResponse(extras.ERR_SUCCESS, scannedPerformance)
	}

	scannedPerformance := ScannedPerformance{
		ScannedUrlPercent:  urlTaskCount,
		ScannedFilePercent: fileTaskCount,
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, scannedPerformance)
}

func EventCount(duration time.Duration, action string, infinite bool) model.APIResponse {
	var eventCount int64 = 0
	resp := GetAllProfiles(extras.LOGREPORT, action)

	if resp.Data == nil || resp.StatusCode != http.StatusOK {
		return model.NewSuccessResponse(extras.ERR_SUCCESS, eventCount)
	}

	// Reverse the order of Arrays
	resp.Data = util.Reverse(resp.Data)

	endTime := time.Now().Add(-1 * duration)
	for _, eventInterface := range resp.Data.([]any) {
		if event, ok := eventInterface.(map[string]any); ok {
			eventCount++
			t, _ := time.Parse("2006-01-02 15:04:05", event["time_stamp"].(string))
			if !endTime.Before(t) && !infinite {
				break
			}
		}
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, eventCount)
}

func GetHdwrData() model.APIResponse {
	// memory
	ram, err := util.GetRamInfo()
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_GETTING_RAM_INFO, err)
	}

	// disk - start from "/" mount point for Linux
	// then use "\" instead of "/"
	disk, err := util.GetSpaceInfo()
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_GETTING_DISK_INFO, err)
	}

	// cpu - get CPU number of cores and speed
	cpu, err := util.GetCpuInfo()
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	}

	hdwrUsage := HdwrUsage{
		RAM:  float64(ram),
		CPU:  float64(cpu),
		Disk: float64(disk),
		Time: time.Now().Format("2006-01-02 15:04:05"),
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, hdwrUsage)
}

func GetDeviceInfo() model.APIResponse {
	devices, err := dao.FetchDeviceProfile(map[string]any{})
	if err == extras.ErrNoRecordForDevice {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_RECORD_NOT_FOUND, extras.ErrNoRecordForDevice)
	} else if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, devices)
}

// // InstallCommandIfMissing checks if a command is available and installs it if it's missing

// // installDmiDecode installs the dmidecode command

// GetDevice retrieves information about the device.
func GetDevice() model.APIResponse {
	// err := util.InstallCommandIfMissing(extras.DmiDecodeCmd)
	// if err != nil {
	// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	// }

	// println("Getting device 0")

	// Retrieve device model
	// Model, err := executeCommand(extras.DmiDecodeCmd, "-s", "system-product-name")
	// if err != nil {
	// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	// }
	// Model := strings.TrimSpace("WiJungle Anti-APT")

	// println("Getting device 1")

	// Retrieve device serial number
	// serial, err := executeCommand(extras.DmiDecodeCmd, "-s", "system-serial-number")
	// if err != nil {
	// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	// }
	// serial = strings.TrimSpace(serial)

	// println("Getting device 2")
	Model := ""
	serial := ""
	file, err := os.Open(extras.ROOT_DATA_DEVICE_CONFIG)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			if strings.Contains(line, "serial_no") {
				serial = parts[1]
			}
			if strings.Contains(line, "model_no") {
				Model = parts[1]
			}
		}
	}

	// Retrieve OS version
	osVersion, err := getOSVersion()
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	}

	// println("Getting device 3")

	// Retrieve host information
	info, err := host.Info()
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	}

	// println("Getting device 4")

	// Calculating appliance up since time
	uptimeDuration := time.Duration(info.Uptime) * time.Second
	applianceUpSince := time.Now().Add(-uptimeDuration)
	applianceUpSinceLocal := applianceUpSince.Local()

	// Retrieving registered email
	licenseKey, err := dao.FetchLicenseKeyProfile()
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	}

	spec := model.DeviceSpecification{
		ModelNo:          Model,
		SerialNo:         serial,
		RegisteredEmail:  licenseKey.RegisteredEmail,
		OsVersion:        osVersion,
		LicenseValidTill: util.FormatWithOrdinal(licenseKey.ExpiryTime),
		ApplianceUpSince: applianceUpSinceLocal.Format(extras.TimeFormat),
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, &spec)
}

// executeCommand executes a shell command and returns the output
func executeCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		log.Println("Error:", err)
		return "", err
	}
	return string(output), nil
}

// // getOSVersion retrieves the operating system version from /etc/os-release
func getOSVersion() (string, error) {
	output, err := executeCommand("cat", "/etc/os-release")
	if err != nil {
		return "", err
	}
	var osVersion string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			osVersion = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME=\""), "\"")
			break
		}
	}
	if osVersion == "" {
		return "", errors.New("PRETTY_NAME not found in /etc/os-release")
	}
	return osVersion, nil
}

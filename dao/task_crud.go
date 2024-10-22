package dao

import (
	"anti-apt-backend/config"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"fmt"
	"os"
	"strings"

	"gorm.io/gorm"
)

type Database interface {
	Save(db *gorm.DB) error
	Delete(db *gorm.DB) error
	Fetch(db *gorm.DB) error
	ExecQuery(db *gorm.DB) error
}

const (
	SAVE   = "save"
	EXEC   = "exec"
	DELETE = "delete"
	FETCH  = "fetch"
)

func GormOperations(service Database, db *gorm.DB, call string) error {
	switch call {
	case SAVE:
		return service.Save(db)
	case DELETE:
		return service.Delete(db)
	case FETCH:
		return service.Fetch(db)
	case EXEC:
		return service.ExecQuery(db)
	}

	return extras.ErrInvalidOperation
}

type DatabaseOperationsRepo struct {
	Fod          model.FileOnDemand
	Uod          model.UrlOnDemand
	QueryExecSet []string
	Result       interface{}
}

func (task *DatabaseOperationsRepo) ExecQuery(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, query := range task.QueryExecSet {
			var err error
			if strings.Contains(strings.ToLower(query), "select") {
				err = tx.Raw(query).Scan(task.Result).Error
			} else {
				err = tx.Exec(query).Error
			}
			if err != nil {
				// slog.Println("ERROR IN UPDATING TASK: ", err)
				return err
			}
		}
		return nil
	})
}

func (task *DatabaseOperationsRepo) Save(db *gorm.DB) error {
	return nil
}

func (task *DatabaseOperationsRepo) Delete(db *gorm.DB) error {
	return nil
}

func (task *DatabaseOperationsRepo) Fetch(db *gorm.DB) error {
	return nil
}

type FileHashesRepo struct {
	FileHash     model.FileHashes
	QueryExecSet []string
	Result       interface{}
}

func (task *FileHashesRepo) ExecQuery(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, query := range task.QueryExecSet {
			var err error
			if strings.Contains(strings.ToLower(query), "select") {
				err = tx.Raw(query).Scan(task.Result).Error
			} else {
				err = tx.Exec(query).Error
			}
			if err != nil {
				// slog.Println("ERROR IN UPDATING TASK: ", err)
				return err
			}
		}
		return nil
	})
}

func (task *FileHashesRepo) Delete(db *gorm.DB) error {
	return nil
}

func (task *FileHashesRepo) Save(db *gorm.DB) error {
	return nil
}

func (task *FileHashesRepo) Fetch(db *gorm.DB) error {
	return nil
}

func ResetQueueDb() {
	rebooted, err := readDeviceRebootedFlag()
	if err != nil {
		// slog.Println("error reading file: %v", err)
	}

	if rebooted {
		err := config.Db.Model(&model.TaskLiveAnalysisTable{}).Where("id > 0").Update("status", extras.PendingNotInQueue).Error
		if err != nil {
			// slog.Println("error updating tasks: %v", err)
		}
		setDeviceRebootedFlagTo0()
	} else {
		// update all tasks to PendingNotInQueue where status = PendingInQueue
		err := config.Db.Model(&model.TaskLiveAnalysisTable{}).Where("status = ?", extras.PendingInQueue).Update("status", extras.PendingNotInQueue).Error
		if err != nil {
			// slog.Println("error updating  pending tasks: %v", err)
		}

		// update all tasks to RunningNotInQueue where status = RunningInQueue
		err = config.Db.Model(&model.TaskLiveAnalysisTable{}).Where("status = ?", extras.RunningInQueue).Update("status", extras.RunningNotInQueue).Error
		if err != nil {
			// slog.Println("error updating running tasks: %v", err)
		}
	}
}

func readDeviceRebootedFlag() (bool, error) {
	filePath := extras.DEVICE_REBOOTED_FLAG_PATH

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, fmt.Errorf("file does not exist")
	} else if err != nil {
		return false, fmt.Errorf("error checking file: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("error reading file: %v", err)
	}

	trimmedContent := strings.TrimSpace(string(content))
	if trimmedContent == "1" {
		return true, nil
	} else if trimmedContent == "0" {
		return false, nil
	} else {
		return false, fmt.Errorf("invalid file content")
	}
}

func setDeviceRebootedFlagTo0() {
	filePath := extras.DEVICE_REBOOTED_FLAG_PATH
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			// slog.Println("error creating file: %v", err)
			return
		}
		defer file.Close()
	} else if err != nil {
		// slog.Println("error checking file: %v", err)
		return
	}

	err = os.WriteFile(filePath, []byte("0"), 0644)
	if err != nil {
		// slog.Println("error writing to file: %v", err)
		return
	}
}

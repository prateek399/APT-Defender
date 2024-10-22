package service

import (
	"anti-apt-backend/config"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"anti-apt-backend/model"
	"anti-apt-backend/service/interfaces"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

func Restore(form *multipart.Form, curUsr string) model.APIResponse {
	if form.File == nil || form.File["file"] == nil || len(form.File["file"]) == 0 {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_REQUIRED_FIELD_EMPTY, extras.ErrRequiredFieldEmpty)
	}

	if filepath.Ext(form.File["file"][0].Filename) != ".yaml" {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_EXTENSION_NOT_SUPPORTED, extras.ErrExtensionNotSupported)
	}

	backupTime, err := dao.FetchBackupConfigTime(map[string]any{})
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage(extras.ERR_IN_FETCHING_DATA))
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	}

	file, err := form.File["file"][0].Open()
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
	}
	defer file.Close()

	var uploadedData map[string]any
	if err = yaml.NewDecoder(file).Decode(&uploadedData); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Restore config has been initiated by %s", curUsr)))

	resp := CopyOldConfigData("RESTORE")
	if resp.StatusCode != http.StatusOK {
		// slog.Println("Error while copying old config data")
		return resp
	}

	// marshaling uploaded config data
	configDataBytes, err := yaml.Marshal(uploadedData["config"])
	if err != nil {
		// slog.Println("Error while marshaling uploaded config data")
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	// unmarshaling uploaded config data
	var configData model.Config
	if err := yaml.Unmarshal(configDataBytes, &configData); err != nil {
		// slog.Println("Error while unmarshaling uploaded config data")
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	// Before writing data to config file merge and overwrite it with the existing ones
	configData = config.MergeConfigAndUploadedData(extras.CONFIG_FILE_NAME, configData)

	// create config file
	configf, err := os.Create(extras.CONFIG_FILE_NAME)
	if err != nil {
		// slog.Println("Error while creating config file")
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}
	defer configf.Close()

	// marshaling uploaded task config data
	// taskConfigDataBytes, err := yaml.Marshal(uploadedData["task_config"])
	// if err != nil {
	// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	// }

	// unmarshaling uploaded task config data
	// var taskConfigData model.TaskConfig
	// if err := yaml.Unmarshal(taskConfigDataBytes, &taskConfigData); err != nil {
	// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	// }

	// create task config file
	// taskConfigf, err := os.Create(extras.TASK_CONFIG_FILE_NAME)
	// if err != nil {
	// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	// }
	// defer taskConfigf.Close()

	// Before writing data to config file merge and overwrite it with the existing ones
	// taskConfigData = config.MergeTaskConfigAndUploadedData(extras.TASK_CONFIG_FILE_NAME, taskConfigData)

	// create interface config file
	interfaceConfigf, err := os.Create(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		// slog.Println("Error while creating interface config file")
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}
	defer interfaceConfigf.Close()

	// .............. Write to Config File ..................... //

	// copy new configs into files
	if err = yaml.NewEncoder(configf).Encode(configData); err != nil {
		// slog.Println("Error while writing to config file")
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	if err = yaml.NewEncoder(interfaceConfigf).Encode(uploadedData["interface_config"]); err != nil {
		// slog.Println("Error while writing to interface config file")
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	// if err = yaml.NewEncoder(taskConfigf).Encode(taskConfigData); err != nil {
	// 	return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	// }

	// restore interface settings
	err = interfaces.RestoreInterfaceSettings("restore")
	if err != nil {
		// slog.Println("Error while restoring interface settings while system restore process ")
		logger.LoggerFunc("error", logger.LoggerMessage("Error in restoring interface settings while system restore process "))
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	// saving restore and backup time
	restoreTime := model.RestoreTime{
		Time: time.Now(),
	}
	if err = dao.SaveProfile([]interface{}{restoreTime, backupTime}, extras.POST); err != nil {
		// slog.Println("Error while saving restore and backup time")
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, "Data restored successfully")
}

func CopyOldConfigData(caller string) model.APIResponse {

	// slog.Println("Copying old config data")

	tempConfigf, err := os.Create(extras.OLD_CONFIG_FILE_NAME)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}
	defer tempConfigf.Close()

	tempInterfaceConfigf, err := os.Create(extras.OLD_INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}
	defer tempInterfaceConfigf.Close()

	mergeConfig := config.MergeInterfaceAndConfigData(caller)

	if err = yaml.NewEncoder(tempConfigf).Encode(mergeConfig["config"]); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	if err = yaml.NewEncoder(tempInterfaceConfigf).Encode(mergeConfig["interface_config"]); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	if caller == extras.ERASE {
		tempTaskConfigf, err := os.Create(extras.OLD_TASK_CONFIG_FILE_NAME)
		if err != nil {
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
		}
		defer tempTaskConfigf.Close()

		if err = yaml.NewEncoder(tempTaskConfigf).Encode(mergeConfig["task_config"]); err != nil {
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
		}
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, mergeConfig)
}

func RestoreOldConfigData() model.APIResponse {
	tempConfigBytes, err := os.ReadFile(extras.OLD_CONFIG_FILE_NAME)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	tempInterfaceConfigBytes, err := os.ReadFile(extras.OLD_INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	tempTaskConfigBytes, err := os.ReadFile(extras.OLD_TASK_CONFIG_FILE_NAME)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	if err := os.WriteFile(extras.CONFIG_FILE_NAME, tempConfigBytes, 0777); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	if err := os.WriteFile(extras.INTERFACE_CONFIG_FILE_NAME, tempInterfaceConfigBytes, 0777); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	if err := os.WriteFile(extras.TASK_CONFIG_FILE_NAME, tempTaskConfigBytes, 0777); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, "Successfully restored data")
}

func EraseConfig(caller string, curUsr string) model.APIResponse {
	resp := CopyOldConfigData(extras.ERASE)
	if resp.StatusCode != http.StatusOK {
		return resp
	}

	factoryDefaultDataBytes, err := os.ReadFile(extras.FACTORY_DEFAULT_CONFIG_FILE_NAME)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	ok := true
	defer func(ok *bool) {
		if !*ok {
			resp := RestoreOldConfigData()
			if resp.StatusCode != http.StatusOK {
				log.Println("error while restoring data: ", resp)
				return
			}
		}
	}(&ok)

	if err := os.Remove(extras.CONFIG_FILE_NAME); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	if err := os.Remove(extras.TASK_CONFIG_FILE_NAME); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	if err := os.WriteFile(extras.DATABASE_PATH+"logFile.txt", []byte{}, 0777); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	outConfigf, err := os.OpenFile(extras.CONFIG_FILE_NAME, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	if _, err := outConfigf.Write(factoryDefaultDataBytes); err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	go func() {
		time.Sleep(4 * time.Second)
		err = interfaces.ResetToFactoryDefaultSettingsForInterfaces()
		if err != nil {
			logger.LoggerFunc("error", logger.LoggerMessage("Error in resetting interface settings while system erase"))
		}

		ServiceActions([]string{extras.SERVICE_KEEPALIVED, extras.SERVICE_ZEBRA}, extras.Stop, 0)
		err = reboot()
		if err != nil {
			logger.LoggerFunc("error", logger.LoggerMessage("Error in rebooting system after factory default settings"))
		}
	}()

	ok = true

	if caller != extras.INITIAL_SETUP {
		logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Reset to Factory default settings has been initiated by %s", curUsr)))
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, "Factory default settings has been successfully initiated. System will be rebooted in 3 seconds. Please wait for a while.")
}

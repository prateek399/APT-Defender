package config

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	InterfaceModel "anti-apt-backend/model/interface_model"
	"anti-apt-backend/util"
	"fmt"
	"os"
	"reflect"
	"sort"

	lock "github.com/subchen/go-trylock"
	"gopkg.in/yaml.v3"
)

var WriteMu = lock.New()

func readConfigFile(fp string) ([]byte, error) {
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		err := os.WriteFile(fp, []byte("{}"), 0777)
		if err != nil {
			return nil, err
		}
	}

	return os.ReadFile(fp)
}

func MergeTaskConfigAndUploadedData(filePath string, uploadedData model.TaskConfig) model.TaskConfig {
	if ok := WriteMu.TryLock(extras.LOCK_TIME_OUT); !ok {
		return model.TaskConfig{}
	}

	defer WriteMu.Unlock()

	configDataInBytes, err := os.ReadFile(filePath)
	if err != nil {
		return model.TaskConfig{}
	}

	var taskConfig model.TaskConfig

	if err := yaml.Unmarshal(configDataInBytes, &taskConfig); err != nil {
		return model.TaskConfig{}
	}

	var t = reflect.TypeOf(uploadedData)
	for i := 0; i < t.NumField(); i++ {
		configField := reflect.ValueOf(&taskConfig).Elem().Field(i)
		uploadField := reflect.ValueOf(&uploadedData).Elem().Field(i)
		if configField.Type() == uploadField.Type() {
			configField.Set(uploadField)
		}
	}

	return taskConfig
}

func MergeConfigAndUploadedData(filePath string, uploadedData model.Config) model.Config {

	// slog.Println("Merging config and uploaded data")

	if ok := WriteMu.TryLock(extras.LOCK_TIME_OUT); !ok {
		return model.Config{}
	}

	defer WriteMu.Unlock()

	configDataInBytes, err := os.ReadFile(filePath)
	if err != nil {
		// slog.Println("config file not found")
		return model.Config{}
	}

	var config model.Config

	if err := yaml.Unmarshal(configDataInBytes, &config); err != nil {
		// slog.Println("config file not found")
		return model.Config{}
	}

	var configMap = make(map[string]map[string]bool)
	configMap["UserAuthentications"] = make(map[string]bool)
	configMap["ScanProfiles"] = make(map[string]bool)
	configMap["Admins"] = make(map[string]bool)
	configMap["Roles"] = make(map[string]bool)
	configMap["Features"] = make(map[string]bool)
	configMap["RoleAndActions"] = make(map[string]bool)

	for _, v := range config.UserAuthentications {
		configMap["UserAuthentications"][v.Key] = true
	}

	for _, v := range config.Admins {
		configMap["Admins"][v.Key] = true
	}

	for _, v := range config.Roles {
		configMap["Roles"][v.Key] = true
	}

	for _, v := range config.Features {
		configMap["Features"][v.Key] = true
	}

	for _, v := range config.RoleAndActions {
		configMap["RoleAndActions"][v.Key] = true
	}

	for _, v := range config.ScanProfiles {
		// slog.Println("USER AUTH KEY: ", v.UserAuthenticationKey)
		configMap["ScanProfiles"][v.UserAuthenticationKey] = true
	}

	var t = reflect.TypeOf(uploadedData)
	for i := 0; i < t.NumField(); i++ {
		// slog.Println("FIELD: ", t.Field(i).Name)
		if modelMap, foundModel := configMap[t.Field(i).Name]; foundModel {
			switch t.Field(i).Name {

			case extras.CONFIG_SCANPROFILE:
				config.ScanProfiles = uploadedData.ScanProfiles

			case extras.CONFIG_ADMIN:
				for j := 0; j < len(uploadedData.Admins); j++ {
					if _, foundVal := modelMap[uploadedData.Admins[j].Key]; !foundVal {
						config.Admins = append(config.Admins, uploadedData.Admins[j])
					}
				}

			case extras.CONFIG_FEATURE:
				for j := 0; j < len(uploadedData.Features); j++ {
					if _, foundVal := modelMap[uploadedData.Features[j].Key]; !foundVal {
						config.Features = append(config.Features, uploadedData.Features[j])
					}
				}

			case extras.CONFIG_ROLE:
				for j := 0; j < len(uploadedData.Roles); j++ {
					if _, foundVal := modelMap[uploadedData.Roles[j].Key]; !foundVal {
						config.Roles = append(config.Roles, uploadedData.Roles[j])
					}
				}

			case extras.CONFIG_ROLEANDACTION:
				for j := 0; j < len(uploadedData.RoleAndActions); j++ {
					if _, foundVal := modelMap[uploadedData.RoleAndActions[j].Key]; !foundVal {
						config.RoleAndActions = append(config.RoleAndActions, uploadedData.RoleAndActions[j])
					}
				}

			case extras.CONFIG_USERAUTHENTICATION:
				for j := 0; j < len(uploadedData.UserAuthentications); j++ {
					if _, foundVal := modelMap[uploadedData.UserAuthentications[j].Key]; !foundVal {
						config.UserAuthentications = append(config.UserAuthentications, uploadedData.UserAuthentications[j])
					}
				}

			case extras.DEVICE:
				config.Devices = uploadedData.Devices

			default:
				configField := reflect.ValueOf(&config).Elem().Field(i)
				uploadField := reflect.ValueOf(&uploadedData).Elem().Field(i)
				if configField.Type() == uploadField.Type() {
					configField.Set(uploadField)
				}
			}
		} else {
			configField := reflect.ValueOf(&config).Elem().Field(i)
			uploadField := reflect.ValueOf(&uploadedData).Elem().Field(i)
			if configField.Type() == uploadField.Type() {
				configField.Set(uploadField)
			}
		}
	}

	// slog.Println("Uploaded data : ", uploadedData.ScanProfiles)
	// slog.Println("Config data: ", config.ScanProfiles)

	return config
}

func MergeInterfaceAndConfigData(caller string) map[string]any {

	// slog.Println("Merging interface and config data")

	if ok := WriteMu.TryLock(extras.LOCK_TIME_OUT); !ok {
		return nil
	}

	defer WriteMu.Unlock()

	// Open the file
	configDataInBytes, err := os.ReadFile(extras.CONFIG_FILE_NAME)
	if err != nil {
		// slog.Println("config file not found")
		return nil
	}

	interfaceConfigDataInBytes, err := os.ReadFile(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		// slog.Println("interface config file not found")
		return nil
	}

	var config model.Config
	if err = yaml.Unmarshal(configDataInBytes, &config); err != nil {
		// slog.Println("config file not found")
		return nil
	}

	var interfaceConfig InterfaceModel.Config
	if err = yaml.Unmarshal(interfaceConfigDataInBytes, &interfaceConfig); err != nil {
		// slog.Println("interface config file not found")
		return nil
	}

	mergedConfig := make(map[string]any)
	mergedConfig["config"] = config
	mergedConfig["interface_config"] = interfaceConfig

	if caller == extras.ERASE {
		// slog.Println("Erase config data")
		taskConfigDataInBytes, err := os.ReadFile(extras.TASK_CONFIG_FILE_NAME)
		if err != nil {
			return nil
		}
		var taskConfig model.TaskConfig
		if err = yaml.Unmarshal(taskConfigDataInBytes, &taskConfig); err != nil {
			return nil
		}
		mergedConfig["task_config"] = taskConfig

	}

	// slog.Println("Merged interface and config data")

	return mergedConfig
}

var muTask = lock.New()

func UpdateTaskConfig(resp model.TaskConfig, caller string) error {
	if ok := muTask.TryLock(extras.LOCK_TIME_OUT); !ok {
		return fmt.Errorf("internal server error")
	}

	defer muTask.Unlock()

	yamlData, err := readConfigFile(extras.TASK_CONFIG_FILE_NAME)
	if err != nil {
		return err
	}

	var config model.TaskConfig

	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return err
	}

	var t = reflect.TypeOf(resp)
	// var v = reflect.ValueOf(resp)

	for i := 0; i < t.NumField(); i++ {
		switch t.Field(i).Name {
		case extras.CONFIG_FILEONDEMAND:
			if config.FileOnDemands == nil {
				config.FileOnDemands = map[string]model.FileOnDemand{}
			}
			for job := range resp.FileOnDemands {
				config.FileOnDemands[job] = resp.FileOnDemands[job]
			}

		case extras.CONFIG_URLONDEMAND:
			if config.UrlOnDemands == nil {
				config.UrlOnDemands = map[string]model.UrlOnDemand{}
			}
			for job := range resp.UrlOnDemands {
				config.UrlOnDemands[job] = resp.UrlOnDemands[job]
			}

		case extras.CONFIG_OVERRIDDENVERDICT:
			if len(resp.OverriddenVerdicts) > 0 {
				config.OverriddenVerdicts = append(config.OverriddenVerdicts, resp.OverriddenVerdicts...)
			}

		}
	}

	yamlData, err = yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = os.WriteFile(extras.TASK_CONFIG_FILE_NAME, yamlData, 0777)
	if err != nil {
		return err
	}

	return nil
}

func WriteHashesFile(filePath string, data interface{}) error {
	// writeMutex.Lock()
	// defer writeMutex.Unlock()

	yamlData, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_, err = os.Create(filePath)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(filePath, yamlData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func UpdateLicenseKeyData(resp model.LicenseKeyConfig) error {
	// var mu2 = lock.New()
	// if ok := mu2.TryLock(extras.LOCK_TIME_OUT); !ok {
	// 	return fmt.Errorf("internal server error")
	// }

	// defer mu2.Unlock()

	if resp.LicenseKey.ApplianceKey == "" {
		return nil
	}

	yamlData, err := readConfigFile(extras.LICENSEKEY_FILE_PATH)
	if err != nil {
		return err
	}

	var config model.LicenseKeyConfig
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return err
	}

	config = resp

	yamlData, err = yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = os.WriteFile(extras.LICENSEKEY_FILE_PATH, yamlData, 0777)
	if err != nil {
		return err
	}

	return nil
}

func UpdateConfig(resp model.Config, licenseKeyConfig model.LicenseKeyConfig, caller string) error {
	if ok := WriteMu.TryLock(extras.LOCK_TIME_OUT); !ok {
		return fmt.Errorf("internal server error")
	}

	defer WriteMu.Unlock()

	yamlData, err := readConfigFile(extras.CONFIG_FILE_NAME)
	if err != nil {
		return err
	}

	var config model.Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return err
	}

	var t = reflect.TypeOf(resp)

	for i := 0; i < t.NumField(); i++ {
		var found bool = false

		switch t.Field(i).Name {
		case extras.CONFIG_ADMIN:
			if len(resp.Admins) > 0 {
				for i := range config.Admins {
					if config.Admins[i].Key == resp.Admins[0].Key {
						found = true
						config.Admins[i] = resp.Admins[0]
						break
					}
				}
			}
			if !found && len(resp.Admins) > 0 {
				config.Admins = append(config.Admins, resp.Admins...)
			}

		case extras.CONFIG_USERAUTHENTICATION:
			if len(resp.UserAuthentications) > 0 {
				for i := range config.UserAuthentications {
					if config.UserAuthentications[i].Key == resp.UserAuthentications[0].Key {
						found = true
						config.UserAuthentications[i] = resp.UserAuthentications[0]
						break
					}
				}
			}
			if !found && len(resp.UserAuthentications) > 0 {
				config.UserAuthentications = append(config.UserAuthentications, resp.UserAuthentications...)
			}

		case extras.CONFIG_ROLE:
			if len(resp.Roles) > 0 {
				for i := range config.Roles {
					if config.Roles[i].Key == resp.Roles[0].Key {
						found = true
						config.Roles[i] = resp.Roles[0]
						break
					}
				}
			}
			if !found && len(resp.Roles) > 0 {
				config.Roles = append(config.Roles, resp.Roles...)
			}

		case extras.CONFIG_FEATURE:
			if len(resp.Features) > 0 {
				for i := range config.Features {
					if config.Features[i].Key == resp.Features[0].Key {
						found = true
						config.Features[i] = resp.Features[0]
						break
					}
				}
			}
			if !found && len(resp.Features) > 0 {
				config.Features = append(config.Features, resp.Features...)
			}

		case extras.CONFIG_ROLEANDACTION:
			if len(resp.RoleAndActions) > 0 {
				for i := range config.RoleAndActions {
					for j := range resp.RoleAndActions {
						if config.RoleAndActions[i].Key == resp.RoleAndActions[j].Key {
							found = true
							config.RoleAndActions[i] = resp.RoleAndActions[j]
							break
						}
					}
				}
			}
			if !found && len(resp.RoleAndActions) > 0 {
				config.RoleAndActions = append(config.RoleAndActions, resp.RoleAndActions...)
			}

		case extras.CONFIG_SCANPROFILE:
			if len(resp.ScanProfiles) > 0 {
				for i := range config.ScanProfiles {
					if config.ScanProfiles[i].UserAuthenticationKey == resp.ScanProfiles[0].UserAuthenticationKey {
						found = true
						config.ScanProfiles[i] = resp.ScanProfiles[0]
						break
					}
				}
			}
			if !found && len(resp.ScanProfiles) > 0 {
				config.ScanProfiles = append(config.ScanProfiles, resp.ScanProfiles...)
			}

		case extras.CONFIG_DEVICE:
			if len(resp.Devices) > 0 {
				for i := range config.Devices {
					if config.Devices[i].Key == resp.Devices[0].Key {
						found = true
						config.Devices[i] = resp.Devices[0]
						break
					}
				}
			}
			if !found && len(resp.Devices) > 0 {
				config.Devices = append(config.Devices, resp.Devices...)
			}

		case extras.CONFIG_BACKUP:
			if !resp.BackupConfigTime.Time.IsZero() {
				config.BackupConfigTime = resp.BackupConfigTime
			}

		case extras.CONFIG_RESTORE:
			if !resp.RestoreConfigTime.Time.IsZero() {
				config.RestoreConfigTime = resp.RestoreConfigTime
			}
		}

	}

	// err = UpdateTaskConfig(taskResp, caller)
	// if err != nil {
	// 	return err
	// }

	err = UpdateLicenseKeyData(licenseKeyConfig)
	if err != nil {
		return err
	}

	yamlData, err = yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = os.WriteFile(extras.CONFIG_FILE_NAME, yamlData, 0777)
	if err != nil {
		return err
	}

	return nil
}

// Right now there is no need for this
func DeleteTaskConfig(resp model.TaskConfig) error {
	if ok := WriteMu.TryLock(extras.LOCK_TIME_OUT); !ok {
		return fmt.Errorf("internal server error")
	}

	defer WriteMu.Unlock()

	yamlData, err := readConfigFile(extras.TASK_CONFIG_FILE_NAME)
	if err != nil {
		return err
	}

	var config model.TaskConfig
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return err
	}

	var t = reflect.TypeOf(resp)
	var newConfig model.TaskConfig
	for i := 0; i < t.NumField(); i++ {
		switch t.Field(i).Name {
		case extras.CONFIG_FILEONDEMAND:
			if config.FileOnDemands == nil {
				config.FileOnDemands = map[string]model.FileOnDemand{}
			}
			if newConfig.FileOnDemands == nil {
				newConfig.FileOnDemands = map[string]model.FileOnDemand{}
			}
			for job := range config.FileOnDemands {
				if _, found := resp.FileOnDemands[job]; found {
					found = true
					continue
				}
				newConfig.FileOnDemands[job] = config.FileOnDemands[job]
			}

		case extras.CONFIG_URLONDEMAND:
			if config.UrlOnDemands == nil {
				config.UrlOnDemands = map[string]model.UrlOnDemand{}
			}
			if newConfig.UrlOnDemands == nil {
				newConfig.UrlOnDemands = map[string]model.UrlOnDemand{}
			}
			for job := range config.UrlOnDemands {
				if _, found := resp.UrlOnDemands[job]; found {
					found = true
					continue
				}
				newConfig.UrlOnDemands[job] = config.UrlOnDemands[job]
			}
		}
	}

	yamlData, err = yaml.Marshal(&newConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(extras.TASK_CONFIG_FILE_NAME, yamlData, 0777)
	if err != nil {
		return err
	}

	return nil
}

func DeleteConfig(resp model.Config) error {
	if ok := WriteMu.TryLock(extras.LOCK_TIME_OUT); !ok {
		return fmt.Errorf("internal server error")
	}

	defer WriteMu.Unlock()

	yamlData, err := readConfigFile(extras.CONFIG_FILE_NAME)
	if err != nil {
		return err
	}

	var config model.Config

	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return err
	}

	var t = reflect.TypeOf(resp)
	// var v = reflect.ValueOf(resp)

	var newConfig model.Config
	for i := 0; i < t.NumField(); i++ {
		var found bool = false

		switch t.Field(i).Name {
		case extras.CONFIG_ADMIN:
			if len(resp.Admins) > 0 {
				for i := range config.Admins {
					if config.Admins[i].Key == resp.Admins[0].Key {
						found = true
						continue
					}
					newConfig.Admins = append(newConfig.Admins, config.Admins[i])
				}
			} else if !found && len(resp.Admins) > 0 {
				return extras.ErrNoRecordForAdmin
			} else {
				newConfig.Admins = config.Admins
			}

		case extras.CONFIG_USERAUTHENTICATION:
			if len(resp.UserAuthentications) > 0 {
				for i := range config.UserAuthentications {
					if config.UserAuthentications[i].Key == resp.UserAuthentications[0].Key {
						found = true
						continue
					}
					newConfig.UserAuthentications = append(newConfig.UserAuthentications, config.UserAuthentications[i])
				}
			} else if !found && len(resp.UserAuthentications) > 0 {
				return extras.ErrNoRecordForUserAuth
			} else {
				newConfig.UserAuthentications = config.UserAuthentications
			}

		case extras.CONFIG_ROLE:
			if len(resp.Roles) > 0 {
				for i := range config.Roles {
					if config.Roles[i].Key == resp.Roles[0].Key {
						found = true
						continue
					}
					newConfig.Roles = append(newConfig.Roles, config.Roles[i])
				}
			} else if !found && len(resp.Roles) > 0 {
				return extras.ErrNoRecordForRole
			} else {
				newConfig.Roles = config.Roles
			}

		case extras.CONFIG_FEATURE:
			if len(resp.Features) > 0 {
				for i := range config.Features {
					if config.Features[i].Key == resp.Features[0].Key {
						found = true
						continue
					}
					newConfig.Features = append(newConfig.Features, config.Features[i])
				}
			} else if !found && len(resp.Features) > 0 {
				return extras.ErrNoRecordForFeature
			} else {
				newConfig.Features = config.Features
			}

		case extras.CONFIG_ROLEANDACTION:
			if len(resp.RoleAndActions) > 0 {
				for i := range config.RoleAndActions {
					if config.RoleAndActions[i].Key == resp.RoleAndActions[0].Key {
						found = true
						continue
					}
					newConfig.RoleAndActions = append(newConfig.RoleAndActions, config.RoleAndActions[i])
				}
			} else if !found && len(resp.RoleAndActions) > 0 {
				return extras.ErrNoRecordForRoleAndAction
			} else {
				newConfig.RoleAndActions = config.RoleAndActions
			}

		case extras.CONFIG_SCANPROFILE:
			if len(resp.ScanProfiles) > 0 {
				for i := range config.ScanProfiles {
					if config.ScanProfiles[i].UserAuthenticationKey == resp.ScanProfiles[0].UserAuthenticationKey {
						found = true
						continue
					}
					newConfig.ScanProfiles = append(newConfig.ScanProfiles, config.ScanProfiles[i])
				}
			} else if !found && len(resp.ScanProfiles) > 0 {
				return extras.ErrNoRecordForScanProfile
			} else {
				newConfig.ScanProfiles = config.ScanProfiles
			}

		case extras.CONFIG_DEVICE:
			if len(resp.Devices) > 0 {
				for i := range config.Devices {
					if config.Devices[i].Key == resp.Devices[0].Key {
						found = true
						continue
					}
					newConfig.Devices = append(newConfig.Devices, config.Devices[i])
				}
			} else if !found && len(resp.Devices) > 0 {
				return extras.ErrNoRecordForDevice
			} else {
				newConfig.Devices = config.Devices
			}
		case extras.CONFIG_BACKUP:
			newConfig.BackupConfigTime = config.BackupConfigTime

		case extras.CONFIG_RESTORE:
			newConfig.RestoreConfigTime = config.RestoreConfigTime
		}
	}

	yamlData, err = yaml.Marshal(&newConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(extras.CONFIG_FILE_NAME, yamlData, 0777)
	if err != nil {
		return err
	}

	return nil
}

func FetchTaskConfigData(modelName string) (model.TaskConfig, error) {
	// if ok := WriteMu.TryLock(extras.LOCK_TIME_OUT); !ok {
	// 	return model.TaskConfig{}, fmt.Errorf("internal server error")
	// }

	// defer WriteMu.Unlock()

	yamlData, err := readConfigFile(extras.TASK_CONFIG_FILE_NAME)
	if err != nil {
		return model.TaskConfig{}, err
	}

	var config, resp model.TaskConfig
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return resp, err
	}

	switch modelName {
	case extras.FILEONDEMAND:
		resp.FileOnDemands = config.FileOnDemands
	case extras.URLONDEMAND:
		resp.UrlOnDemands = config.UrlOnDemands
	case extras.OVERRIDDENVERDICT:
		resp.OverriddenVerdicts = config.OverriddenVerdicts
	}

	return resp, nil
}

func FetchConfigData(modelName string) (model.Config, error) {
	// if ok := WriteMu.TryLock(extras.LOCK_TIME_OUT); !ok {
	// 	return model.Config{}, fmt.Errorf("internal server error")
	// }

	// defer WriteMu.Unlock()

	var config, resp model.Config

	yamlData, err := readConfigFile(extras.CONFIG_FILE_NAME)
	if err != nil {
		return resp, err
	}

	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return resp, err
	}

	switch modelName {
	case extras.USERAUTHENTICATION:
		resp.UserAuthentications = config.UserAuthentications
	case extras.ADMIN:
		resp.Admins = config.Admins
	case extras.ROLE:
		resp.Roles = config.Roles
	case extras.FEATURE:
		resp.Features = config.Features
	case extras.ROLEANDACTION:
		resp.RoleAndActions = config.RoleAndActions
	case extras.SCANPROFILE:
		resp.ScanProfiles = config.ScanProfiles
	case extras.DEVICE:
		resp.Devices = config.Devices
	case extras.BACKUP:
		resp.BackupConfigTime = config.BackupConfigTime
	case extras.RESTORE:
		resp.RestoreConfigTime = config.RestoreConfigTime
	case extras.VIEWLOG:
		resp.ViewLogs = config.ViewLogs
	}
	// fmt.Println(resp)

	return resp, nil
}

func FetchSpecificTaskConfigData(fields map[string]any, modelName string) (interface{}, error) {
	var resp model.TaskConfig
	config, err := FetchTaskConfigData(modelName)
	if err != nil {
		return resp, err
	}

	switch modelName {
	case extras.FILEONDEMAND:
		var resp = make(map[string]model.FileOnDemand)
		if jobID, ok := fields["JobID"]; ok {
			if _, whatType := jobID.(string); !whatType {
				return nil, fmt.Errorf(`invalid data type found for jobID`)
			}
			if _, found := config.FileOnDemands[jobID.(string)]; !found {
				return nil, extras.ErrNoRecordForFileOnDemand
			}
			resp[jobID.(string)] = config.FileOnDemands[jobID.(string)]
		} else {
			for job, fileondemand := range config.FileOnDemands {
				var t = reflect.TypeOf(fileondemand)
				var v = reflect.ValueOf(fileondemand)
				for j := 0; j < t.NumField(); j++ {
					if jVal, ok := fields[t.Field(j).Name]; ok {
						if err := util.CompareDataType(v.Field(j).Interface(), jVal); err != nil {
							return nil, err
						}
						if v.Field(j).Interface() == jVal {
							resp[job] = fileondemand
						}
					}
				}
			}
		}

		if len(resp) > 0 {
			return resp, nil
		}
		return nil, extras.ErrNoRecordForFileOnDemand

	case extras.URLONDEMAND:
		var resp = make(map[string]model.UrlOnDemand)
		if jobID, ok := fields["JobID"]; ok {
			if _, whatType := jobID.(string); !whatType {
				return nil, fmt.Errorf(`invalid data type found for jobID`)
			}
			if _, found := config.UrlOnDemands[jobID.(string)]; !found {
				return nil, extras.ErrNoRecordForUrlOnDemand
			}
			resp[jobID.(string)] = config.UrlOnDemands[jobID.(string)]
		} else {
			for job, urlondemand := range config.UrlOnDemands {
				var t = reflect.TypeOf(urlondemand)
				var v = reflect.ValueOf(urlondemand)
				for j := 0; j < t.NumField(); j++ {
					if jVal, ok := fields[t.Field(j).Name]; ok {
						if err := util.CompareDataType(v.Field(j).Interface(), jVal); err != nil {
							return nil, err
						}
						if v.Field(j).Interface() == jVal {
							resp[job] = urlondemand
						}
					}
				}
			}
		}

		if len(resp) > 0 {
			return resp, nil
		}
		return nil, extras.ErrNoRecordForUrlOnDemand
	}

	return nil, fmt.Errorf(`invalid value found for "%s" model`, modelName)
}

func FetchSpecificConfigData(fields map[string]any, modelName string) (interface{}, error) {
	var resp model.Config

	config, err := FetchConfigData(modelName)
	if err != nil {
		return resp, err
	}

	// fmt.Println(config)

	switch modelName {
	case extras.USERAUTHENTICATION:
		var resp []model.UserAuthentication
		for i := 0; i < len(config.UserAuthentications); i++ {
			var t = reflect.TypeOf(config.UserAuthentications[i])
			var v = reflect.ValueOf(config.UserAuthentications[i])
			for j := 0; j < t.NumField(); j++ {
				if jVal, ok := fields[t.Field(j).Name]; ok {
					if err := util.CompareDataType(v.Field(j).Interface(), jVal); err != nil {
						return nil, err
					}
					if v.Field(j).Interface() == jVal {
						resp = append(resp, config.UserAuthentications[i])
					}
				}
			}
		}
		if len(resp) > 0 {
			sort.Slice(resp, func(i, j int) bool {
				return resp[i].CreatedAt.After(resp[j].CreatedAt)
			})
			return resp, nil
		}
		return nil, extras.ErrNoRecordForUserAuth

	case extras.ADMIN:
		var resp []model.Admin
		for i := 0; i < len(config.Admins); i++ {
			var t = reflect.TypeOf(config.Admins[i])
			var v = reflect.ValueOf(config.Admins[i])
			for j := 0; j < t.NumField(); j++ {
				if jVal, ok := fields[t.Field(j).Name]; ok {
					if err := util.CompareDataType(v.Field(j).Interface(), jVal); err != nil {
						return nil, err
					}
					if v.Field(j).Interface() == jVal {
						resp = append(resp, config.Admins[i])
					}
				}
			}
		}
		if len(resp) > 0 {
			sort.Slice(resp, func(i, j int) bool {
				return resp[i].CreatedAt.After(resp[j].CreatedAt)
			})
			return resp, nil
		}
		return nil, extras.ErrNoRecordForAdmin

	case extras.ROLE:
		var resp []model.Role
		for i := 0; i < len(config.Roles); i++ {
			var t = reflect.TypeOf(config.Roles[i])
			var v = reflect.ValueOf(config.Roles[i])
			for j := 0; j < t.NumField(); j++ {
				if jVal, ok := fields[t.Field(j).Name]; ok {
					if err := util.CompareDataType(v.Field(j).Interface(), jVal); err != nil {
						return nil, err
					}
					if v.Field(j).Interface() == jVal {
						resp = append(resp, config.Roles[i])
					}
				}
			}
		}
		if len(resp) > 0 {
			sort.Slice(resp, func(i, j int) bool {
				return resp[i].CreatedAt.After(resp[j].CreatedAt)
			})
			return resp, nil
		}
		return nil, extras.ErrNoRecordForRole

	case extras.FEATURE:
		var resp []model.Feature
		for i := 0; i < len(config.Features); i++ {
			var t = reflect.TypeOf(config.Features[i])
			var v = reflect.ValueOf(config.Features[i])
			for j := 0; j < t.NumField(); j++ {
				if jVal, ok := fields[t.Field(j).Name]; ok {
					if err := util.CompareDataType(v.Field(j).Interface(), jVal); err != nil {
						return nil, err
					}
					if v.Field(j).Interface() == jVal {
						resp = append(resp, config.Features[i])
					}
				}
			}
		}
		if len(resp) > 0 {
			return resp, nil
		}
		return nil, extras.ErrNoRecordForFeature

	case extras.ROLEANDACTION:
		var resp []model.RoleAndAction
		for i := 0; i < len(config.RoleAndActions); i++ {
			var t = reflect.TypeOf(config.RoleAndActions[i])
			var v = reflect.ValueOf(config.RoleAndActions[i])
			for j := 0; j < t.NumField(); j++ {
				if jVal, ok := fields[t.Field(j).Name]; ok {
					if err := util.CompareDataType(v.Field(j).Interface(), jVal); err != nil {
						return nil, err
					}
					if v.Field(j).Interface() == jVal {
						resp = append(resp, config.RoleAndActions[i])
					}
				}
			}
		}
		if len(resp) > 0 {
			sort.Slice(resp, func(i, j int) bool {
				return resp[i].CreatedAt.After(resp[j].CreatedAt)
			})
			return resp, nil
		}
		return nil, extras.ErrNoRecordForRoleAndAction

	case extras.SCANPROFILE:
		var resp []model.ScanProfile
		for i := 0; i < len(config.ScanProfiles); i++ {
			var t = reflect.TypeOf(config.ScanProfiles[i])
			var v = reflect.ValueOf(config.ScanProfiles[i])
			for j := 0; j < t.NumField(); j++ {
				if jVal, ok := fields[t.Field(j).Name]; ok {
					if err := util.CompareDataType(v.Field(j).Interface(), jVal); err != nil {
						return nil, err
					}
					if v.Field(j).Interface() == jVal {
						resp = append(resp, config.ScanProfiles[i])
					}
				}
			}
		}
		if len(resp) > 0 {
			return resp, nil
		}
		return nil, extras.ErrNoRecordForScanProfile

	case extras.DEVICE:
		var resp []model.Device
		for i := 0; i < len(config.Devices); i++ {
			var t = reflect.TypeOf(config.Devices[i])
			var v = reflect.ValueOf(config.Devices[i])
			for j := 0; j < t.NumField(); j++ {
				if jVal, ok := fields[t.Field(j).Name]; ok {
					if err := util.CompareDataType(v.Field(j).Interface(), jVal); err != nil {
						return nil, err
					}
					if v.Field(j).Interface() == jVal {
						resp = append(resp, config.Devices[i])
					}
				}
			}
		}
		if len(resp) > 0 {
			sort.Slice(resp, func(i, j int) bool {
				return resp[i].CreatedAt.After(resp[j].CreatedAt)
			})
			return resp, nil
		}
		return nil, extras.ErrNoRecordForDevice
	}

	return nil, fmt.Errorf(`invalid value found for "%s" model`, modelName)
}

func FetchLicenseKeyData() (model.LicenseKeyConfig, error) {
	// if ok := WriteMu.TryLock(extras.LOCK_TIME_OUT); !ok {
	// 	return model.LicenseKeyConfig{}, fmt.Errorf("internal server error")
	// }

	// defer WriteMu.Unlock()

	var config model.LicenseKeyConfig

	yamlData, err := readConfigFile(extras.LICENSEKEY_FILE_PATH)
	if err != nil {
		return model.LicenseKeyConfig{}, err
	}

	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return model.LicenseKeyConfig{}, err
	}

	if config.LicenseKey.ApplianceKey == "" {
		return model.LicenseKeyConfig{}, extras.ErrNoRecordForLicenseKey
	}
	// fmt.Println(resp)

	return config, nil
}

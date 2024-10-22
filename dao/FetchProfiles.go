package dao

import (
	"anti-apt-backend/config"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
)

func FetchAdminProfile(fields map[string]any) ([]model.Admin, error) {
	if len(fields) == 0 {
		data, err := config.FetchConfigData(extras.ADMIN)
		if err != nil {
			return []model.Admin{}, err
		}

		return data.Admins, nil
	}

	data, err := config.FetchSpecificConfigData(fields, extras.ADMIN)
	if err != nil {
		return []model.Admin{}, err
	}

	return data.([]model.Admin), nil
}

func FetchUserAuthProfile(fields map[string]any) ([]model.UserAuthentication, error) {
	if len(fields) == 0 {
		data, err := config.FetchConfigData(extras.USERAUTHENTICATION)
		if err != nil {
			return []model.UserAuthentication{}, err
		}

		return data.UserAuthentications, nil
	}

	data, err := config.FetchSpecificConfigData(fields, extras.USERAUTHENTICATION)
	if err != nil {
		return []model.UserAuthentication{}, err
	}

	return data.([]model.UserAuthentication), nil
}

func FetchLicenseKeyProfile() (model.KeysTable, error) {
	data, err := config.FetchLicenseKeyData()
	if err != nil {
		return model.KeysTable{}, err
	}

	return data.LicenseKey, nil
}

func FetchRoleProfile(fields map[string]any) ([]model.Role, error) {
	if len(fields) == 0 {
		data, err := config.FetchConfigData(extras.ROLE)
		if err != nil {
			return []model.Role{}, err
		}

		return data.Roles, nil
	}

	data, err := config.FetchSpecificConfigData(fields, extras.ROLE)
	if err != nil {
		return []model.Role{}, err
	}

	return data.([]model.Role), nil
}

func FetchRoleAndActionProfile(fields map[string]any) ([]model.RoleAndAction, error) {
	if len(fields) == 0 {
		data, err := config.FetchConfigData(extras.ROLEANDACTION)
		if err != nil {
			return []model.RoleAndAction{}, err
		}

		return data.RoleAndActions, nil
	}

	data, err := config.FetchSpecificConfigData(fields, extras.ROLEANDACTION)
	if err != nil {
		return []model.RoleAndAction{}, err
	}

	return data.([]model.RoleAndAction), nil
}

func FetchFeatureProfile(fields map[string]any) ([]model.Feature, error) {
	if len(fields) == 0 {
		data, err := config.FetchConfigData(extras.FEATURE)
		if err != nil {
			return []model.Feature{}, err
		}

		return data.Features, nil
	}

	data, err := config.FetchSpecificConfigData(fields, extras.FEATURE)
	if err != nil {
		return []model.Feature{}, err
	}

	return data.([]model.Feature), nil
}

func FetchScanProfile(fields map[string]any) ([]model.ScanProfile, error) {
	if len(fields) == 0 {
		data, err := config.FetchConfigData(extras.SCANPROFILE)
		if err != nil {
			return []model.ScanProfile{}, err
		}

		data.ScanProfiles[0].UserAuthenticationKey = ""
		return data.ScanProfiles, nil
	}

	data, err := config.FetchSpecificConfigData(fields, extras.SCANPROFILE)
	if err != nil {
		return []model.ScanProfile{}, err
	}

	data.([]model.ScanProfile)[0].UserAuthenticationKey = ""
	return data.([]model.ScanProfile), nil
}

func FetchDeviceProfile(fields map[string]any) ([]model.Device, error) {
	if len(fields) == 0 {
		data, err := config.FetchConfigData(extras.DEVICE)
		if err != nil {
			return []model.Device{}, err
		}

		return data.Devices, nil
	}

	data, err := config.FetchSpecificConfigData(fields, extras.DEVICE)
	if err != nil {
		return []model.Device{}, err
	}

	resp := util.Reverse(data).([]model.Device)

	return resp, nil
}

func FetchOverriddenVerdicts(fields map[string]any) ([]model.OverriddenVerdict, error) {

	data, err := config.FetchTaskConfigData(extras.OVERRIDDENVERDICT)
	if err != nil {
		return []model.OverriddenVerdict{}, err
	}

	return data.OverriddenVerdicts, nil
}

func FetchBackupConfigTime(fields map[string]any) (model.BackupTime, error) {
	data, err := config.FetchConfigData(extras.BACKUP)
	if err != nil {
		return model.BackupTime{}, err
	}
	return data.BackupConfigTime, nil
}

func FetchRestoreConfigTime(fields map[string]any) (model.RestoreTime, error) {
	data, err := config.FetchConfigData(extras.RESTORE)
	if err != nil {
		return model.RestoreTime{}, err
	}
	return data.RestoreConfigTime, nil
}

func FetchViewLogs() (model.ViewLog, error) {
	data, err := config.FetchConfigData(extras.VIEWLOG)
	if err != nil {
		return model.ViewLog{}, err
	}

	return data.ViewLogs, nil
}

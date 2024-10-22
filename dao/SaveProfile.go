package dao

import (
	"anti-apt-backend/config"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"reflect"
)

func SaveProfile(profiles []interface{}, caller string) error {
	var newconfig model.Config
	var newTaskconfig model.TaskConfig
	var licenseKeyConfig model.LicenseKeyConfig
	// fmt.Println(profiles)
	for _, profile := range profiles {
		var t = reflect.TypeOf(profile)
		switch t.Name() {
		case extras.ADMIN:
			newconfig.Admins = append(newconfig.Admins, profile.(model.Admin))
		case extras.LICENSEKEY:
			licenseKeyConfig.LicenseKey = profile.(model.KeysTable)
		case extras.USERAUTHENTICATION:
			newconfig.UserAuthentications = append(newconfig.UserAuthentications, profile.(model.UserAuthentication))
		case extras.ROLE:
			newconfig.Roles = append(newconfig.Roles, profile.(model.Role))
		case extras.ROLEANDACTION:
			newconfig.RoleAndActions = append(newconfig.RoleAndActions, profile.(model.RoleAndAction))
		case extras.FEATURE:
			newconfig.Features = append(newconfig.Features, profile.(model.Feature))
		case extras.SCANPROFILE:
			newconfig.ScanProfiles = append(newconfig.ScanProfiles, profile.(model.ScanProfile))
		case extras.OVERRIDDENVERDICT:
			newTaskconfig.OverriddenVerdicts = append(newTaskconfig.OverriddenVerdicts, profile.(model.OverriddenVerdict))
		case extras.DEVICE:
			newconfig.Devices = append(newconfig.Devices, profile.(model.Device))
		case extras.BACKUP:
			newconfig.BackupConfigTime = profile.(model.BackupTime)
		case extras.RESTORE:
			newconfig.RestoreConfigTime = profile.(model.RestoreTime)
		}
	}

	// fmt.Println("After Cuckoo.go: ", newconfig)

	if err := config.UpdateConfig(newconfig, licenseKeyConfig, caller); err != nil {
		return err
	}
	return nil
}

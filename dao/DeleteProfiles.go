package dao

import (
	"anti-apt-backend/config"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"reflect"
)

func DeleteProfile(profiles []interface{}) error {
	var newconfig model.Config
	// fmt.Println(profiles)
	for _, profile := range profiles {
		var t = reflect.TypeOf(profile)
		switch t.Name() {
		case extras.ADMIN:
			newconfig.Admins = append(newconfig.Admins, profile.(model.Admin))
		case extras.USERAUTHENTICATION:
			newconfig.UserAuthentications = append(newconfig.UserAuthentications, profile.(model.UserAuthentication))
		case extras.ROLE:
			newconfig.Roles = append(newconfig.Roles, profile.(model.Role))
		case extras.ROLEANDACTION:
			newconfig.RoleAndActions = append(newconfig.RoleAndActions, profile.(model.RoleAndAction))
		case extras.FEATURE:
			newconfig.Features = append(newconfig.Features, profile.(model.Feature))
		case extras.DEVICE:
			newconfig.Devices = append(newconfig.Devices, profile.(model.Device))
		}
	}

	// fmt.Println(newconfig)

	if err := config.DeleteConfig(newconfig); err != nil {
		return err
	}
	return nil
}

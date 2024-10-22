package service

import (
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"net/http"
	"slices"
	"time"
)

func SetPermissions(roleAndActions []model.RoleAndAction) model.APIResponse {
	var resp model.APIResponse
	var err error
	var profiles []interface{}

	for _, roleAction := range roleAndActions {
		profiles = append(profiles, roleAction)
	}

	if dao.SaveProfile(profiles, extras.PATCH) != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Successfully updated role and action")
	return resp
}

func CreateRolePermission(role model.Role, fromSignUp bool) model.APIResponse {
	var resp model.APIResponse
	var profiles []interface{}
	var err error

	if role.Name == "" {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_REQUIRED_FIELD_EMPTY, extras.ErrRequiredFieldEmpty)
		return resp
	}

	_, err = dao.FetchRoleProfile(map[string]any{"Name": role.Name})
	if err == nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_ROLE_ALREADY_EXISTS, extras.ErrRecordAlreadyExists)
		return resp
	} else if err != extras.ErrNoRecordForRole {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	role.Key = util.GenerateUUID()
	role.CreatedAt = time.Now()

	profiles = append(profiles, role)

	var features []model.Feature
	features, err = dao.FetchFeatureProfile(map[string]any{})
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	var parents []string
	for _, feature := range features {
		parents = append(parents, feature.ParentKey)
	}

	for _, feature := range features {

		if feature.SelfType == 1 {
			profiles = append(profiles, model.RoleAndAction{
				Key:        util.GenerateUUID(),
				RoleKey:    role.Key,
				FeatureKey: feature.Key,
				Permission: 4,
			})
		}

		if feature.SelfType == 3 {
			profiles = append(profiles, model.RoleAndAction{
				Key:        util.GenerateUUID(),
				RoleKey:    role.Key,
				FeatureKey: feature.Key,
				Permission: 4,
			})
		}

		if feature.SelfType == 2 {
			if slices.Contains(parents, feature.Key) {
				profiles = append(profiles, model.RoleAndAction{
					Key:        util.GenerateUUID(),
					RoleKey:    role.Key,
					FeatureKey: feature.Key,
					Permission: 4,
				})
			} else {
				profiles = append(profiles, model.RoleAndAction{
					Key:        util.GenerateUUID(),
					RoleKey:    role.Key,
					FeatureKey: feature.Key,
					Permission: 4,
				})
			}
		}
	}

	if fromSignUp {
		return model.NewSuccessResponse(extras.ERR_SUCCESS, []interface{}{role.Key, profiles})
	}

	if err = dao.SaveProfile(profiles, extras.POST); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		return resp
	}
	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, role.Key)
	return resp
}

func SetRolePermissionForAdmin(roleKey string) model.APIResponse {
	var roleActions []model.RoleAndAction
	var resp model.APIResponse
	var err error

	roleActions, err = dao.FetchRoleAndActionProfile(map[string]any{"RoleKey": roleKey})
	if err == extras.ErrNoRecordForRoleAndAction {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_RECORD_NOT_FOUND, extras.ErrNoRecordForRoleAndAction)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	var features []model.Feature
	features, err = dao.FetchFeatureProfile(map[string]any{})
	if err == extras.ErrNoRecordForFeature {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_RECORD_NOT_FOUND, extras.ErrNoRecordForFeature)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	var rolePermissions = make(map[string]int)
	for _, roleAction := range roleActions {
		rolePermissions[roleAction.FeatureKey] = roleAction.Permission
	}

	var sideMenu = make(map[string]model.ShowFeatures)

	for _, feature := range features {
		if feature.SelfType == 2 {
			parent := sideMenu[feature.ParentKey]
			child := model.ShowFeatures{
				Id:         feature.Key,
				Title:      feature.Title,
				MessageID:  feature.MessageID,
				Permission: rolePermissions[feature.Key],
				Icon:       feature.Icon,
				Path:       feature.Path,
				Typeof:     feature.Type,
				Position:   feature.Position,
			}
			parent.Children = append(parent.Children, child)
			sideMenu[feature.ParentKey] = parent

		}
		if feature.SelfType == 1 {
			parent := sideMenu[feature.Key]
			sideMenu[feature.Key] = model.ShowFeatures{
				Id:         feature.Key,
				Title:      feature.Title,
				MessageID:  feature.MessageID,
				Permission: rolePermissions[feature.Key],
				Icon:       feature.Icon,
				Path:       feature.Path,
				Typeof:     feature.Type,
				Children:   parent.Children,
				Position:   feature.Position,
			}
		}
	}

	var menu []model.ShowFeatures
	for _, feature := range sideMenu {
		menu = append(menu, feature)
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, menu)
	return resp
}

package service

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"net/http"
	"strings"
)

func ChangePassword(adminPassChange model.ChangePass, userAuthKey string) model.APIResponse {
	var resp model.APIResponse

	if adminPassChange.Password == "" || adminPassChange.NewPassword == "" || adminPassChange.ConfirmNewPassword == "" {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, extras.ErrRequiredFieldEmpty)
		return resp
	}

	adminPassChange.Password = strings.TrimSpace(adminPassChange.Password)
	adminPassChange.NewPassword = strings.TrimSpace(adminPassChange.NewPassword)
	adminPassChange.ConfirmNewPassword = strings.TrimSpace(adminPassChange.ConfirmNewPassword)

	userAuth, err := dao.FetchUserAuthProfile(map[string]any{"Key": userAuthKey})
	if err == extras.ErrNoRecordForUserAuth {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_RECORD_NOT_FOUND, extras.ErrNoRecordForUserAuth)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	if !auth.ComparePasswordHash(adminPassChange.Password, userAuth[0].Password) {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INCORRECT_PASSWORD, extras.ErrInvalidPassword)
		return resp
	}

	if adminPassChange.NewPassword != adminPassChange.ConfirmNewPassword {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_PASSWORDS_DO_NOT_MATCH, extras.ErrPasswordDoNotMatch)
		return resp
	}

	hash, err := auth.HashPassword(adminPassChange.NewPassword)
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
		return resp
	}

	userAuth[0].Password = hash
	if err := dao.SaveProfile([]interface{}{userAuth[0]}, extras.PATCH); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		return resp
	}

	return model.NewSuccessResponse(extras.ERR_SUCCESS, "Password updated successfully")
}

package service

import (
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"net/http"
)

func ScanProfile(scanProfileRequest model.ScanProfile, userKey string) model.APIResponse {
	var resp model.APIResponse
	var err error

	scanProfileRequest.UserAuthenticationKey = userKey
	if err = dao.SaveProfile([]interface{}{scanProfileRequest}, extras.PATCH); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Successfully updated scan profile")
	return resp
}

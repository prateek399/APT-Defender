package service

import (
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"anti-apt-backend/validation"
	"net/http"
	"time"
)

func CreateDevice(deviceRequest model.Device) model.APIResponse {
	var resp model.APIResponse

	if errCode, err := validation.ValidateDevice(&deviceRequest, extras.POST); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, errCode, err)
		return resp
	}

	if _, err := dao.FetchDeviceProfile(map[string]any{"SerialNumber": deviceRequest.SerialNumber}); err == nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SERIAL_NUMBER_ALREADY_EXISTS, extras.ErrRecordAlreadyExists)
		return resp
	} else if err != extras.ErrNoRecordForDevice {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	// if _, err := dao.FetchDeviceProfile(map[string]any{"DeviceName": deviceRequest.DeviceName}); err == nil {
	// 	resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_NAME_ALREADY_EXISTS, extras.ErrRecordAlreadyExists)
	// 	return resp
	// } else if err != extras.ErrNoRecordForDevice {
	// 	resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	// 	return resp
	// }

	// if _, err := dao.FetchDeviceProfile(map[string]any{"Email": deviceRequest.Email}); err == nil {
	// 	resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_EMAIL_ALREADY_EXISTS, extras.ErrRecordAlreadyExists)
	// 	return resp
	// } else if err != extras.ErrNoRecordForDevice {
	// 	resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	// 	return resp
	// }

	// if _, err := dao.FetchDeviceProfile(map[string]any{"MobileNumber": deviceRequest.MobileNumber}); err == nil {
	// 	resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_PHONE_ALREADY_EXISTS, extras.ErrRecordAlreadyExists)
	// 	return resp
	// } else if err != extras.ErrNoRecordForDevice {
	// 	resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	// 	return resp
	// }

	deviceRequest.Key = util.GenerateUUID()
	deviceRequest.CreatedAt = time.Now()
	if err := dao.SaveProfile([]interface{}{deviceRequest}, extras.POST); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Successfully created device")
	return resp
}

func DeleteDevice(key string) model.APIResponse {
	var resp model.APIResponse
	if key == "" {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_REQUIRED_FIELD_EMPTY, extras.ErrRequiredFieldEmpty)
		return resp
	}

	var device model.Device
	device.Key = key

	if err := dao.DeleteProfile([]interface{}{device}); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Successfully deleted device")
	return resp
}

func UpdateDevice(deviceRequest model.Device) model.APIResponse {
	var resp model.APIResponse
	var err error

	if deviceRequest.Key == "" || deviceRequest.IpAddress == "" {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_REQUIRED_FIELD_EMPTY, extras.ErrRequiredFieldEmpty)
	}

	if !util.IsValidIp(deviceRequest.IpAddress) {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_IP_ADDRESS, extras.ErrInvalidIPAddress)
	}

	var device []model.Device
	device, err = dao.FetchDeviceProfile(map[string]any{"Key": deviceRequest.Key})
	if err == extras.ErrNoRecordForDevice {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_NAME_ALREADY_EXISTS, extras.ErrRecordAlreadyExists)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	device[0].IpAddress = deviceRequest.IpAddress
	if err := dao.SaveProfile([]interface{}{device[0]}, extras.PATCH); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Updated device successfully")
	return resp
}

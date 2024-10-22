package validation

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"strings"
)

func ValidateDevice(deviceRequest *model.Device, method string) (string, error) {
	if method == extras.POST {
		if deviceRequest.DeviceName == "" || deviceRequest.SerialNumber == "" || deviceRequest.ProductCategory == "" ||
			deviceRequest.IpAddress == "" || deviceRequest.Email == "" || deviceRequest.MobileNumber == "" ||
			deviceRequest.Country == "" || deviceRequest.State == "" || deviceRequest.City == "" {

			return extras.ERR_REQUIRED_FIELD_EMPTY, extras.ErrRequiredFieldEmpty
		}

		if len(deviceRequest.SerialNumber) < extras.MINIMUM_ALLOWED_SERIAL_NUMBER_LENGTH || len(deviceRequest.SerialNumber) > extras.MAXIMUM_ALLOWED_SERIAL_NUMBER_LENGTH {
			return extras.ERR_SERIAL_NUMBER_LENGTH_NOT_IN_RANGE, extras.ErrSerialNumberLengthNotInRange
		}

		// if deviceRequest.SerialNumber[0] != "W" {

		// }

		deviceRequest.DeviceName = util.TrimString(deviceRequest.DeviceName)
		// if !util.IsValidName(deviceRequest.DeviceName) {
		// 	return extras.ERR_INVALID_NAME_FORMAT, extras.ErrInvalidNameFormat
		// }

		deviceRequest.Email = util.TrimString(deviceRequest.Email)
		if !util.IsValidEmail(deviceRequest.Email) {
			return extras.ERR_EMAIL_FOUND_INVALID, extras.ErrInvalidEmailFound
		}

		if !util.IsValidCountryCode(deviceRequest.Country) {
			return extras.ERR_INVALID_COUNTRY_CODE, extras.ErrInvalidCountryCode
		}

		deviceRequest.MobileNumber = util.TrimString(deviceRequest.MobileNumber)
		if !util.IsValidPhone(deviceRequest.MobileNumber, deviceRequest.Country) {
			return extras.ERR_INVALID_PHONE_NUMBER, extras.ErrInvalidPhoneNumber
		}

		deviceRequest.IpAddress = strings.TrimSpace(deviceRequest.IpAddress)
		if !util.IsValidIp(deviceRequest.IpAddress) {
			return extras.ERR_INVALID_IP_ADDRESS, extras.ErrInvalidIPAddress
		}

	} else {
		if len(deviceRequest.SerialNumber) > 0 || len(deviceRequest.DeviceName) > 0 {
			return extras.ERR_NON_EDITABLE_FIELD, extras.ErrNonEditableFieldFound
		}

		if deviceRequest.Email != "" {
			if !util.IsValidEmail(deviceRequest.Email) {
				return extras.ERR_EMAIL_FOUND_INVALID, extras.ErrInvalidEmailFound
			}
		}

		if deviceRequest.Country != "" {
			if !util.IsValidCountryCode(deviceRequest.Country) {
				return extras.ERR_INVALID_COUNTRY_CODE, extras.ErrInvalidCountryCode
			}

			if deviceRequest.MobileNumber == "" {
				return extras.ERR_REQUIRED_FIELD_EMPTY, extras.ErrRequiredFieldEmpty
			}

			if !util.IsValidPhone(deviceRequest.MobileNumber, deviceRequest.Country) {
				return extras.ERR_INVALID_PHONE_NUMBER, extras.ErrInvalidPhoneNumber
			}
		}

		if deviceRequest.IpAddress != "" {
			if !util.IsValidIp(deviceRequest.IpAddress) {
				return extras.ERR_INVALID_IP_ADDRESS, extras.ErrInvalidIPAddress
			}
		}

	}

	return extras.ERR_SUCCESS, nil
}

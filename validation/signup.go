package validation

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"net/http"
)

func ValidateSignupRequest(signupRequest model.SignupRequest) model.APIResponse {
	var resp model.APIResponse

	if signupRequest.Name == "" || signupRequest.Username == "" || signupRequest.Password == "" || signupRequest.Email == "" ||
		signupRequest.ConfirmPassword == "" || signupRequest.CountryCode == "" || signupRequest.Phone == "" {

		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_REQUIRED_FIELD_EMPTY, extras.ErrRequiredFieldEmpty)
		return resp
	}

	if signupRequest.Password != signupRequest.ConfirmPassword {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_PASSWORDS_DO_NOT_MATCH, extras.ErrPasswordDoNotMatch)
		return resp
	}

	if !util.IsValidEmail(signupRequest.Email) {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_EMAIL_FOUND_INVALID, extras.ErrInvalidEmailFound)
		return resp
	}

	if signupRequest.IsSuperAdmin {
		if err := util.ValidateLicenseKey(signupRequest.LicenseKey); err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_LICENSE_KEY, err)
			return resp
		}
	}

	if !util.IsValidCountryCode(signupRequest.CountryCode) {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_COUNTRY_CODE, extras.ErrInvalidCountryCode)
		return resp
	}

	if !util.IsValidPhone(signupRequest.Phone, signupRequest.CountryCode) {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_PHONE_NUMBER, extras.ErrInvalidPhoneNumber)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, nil)
	return resp
}

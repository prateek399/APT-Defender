package service

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"anti-apt-backend/validation"
	"net/http"

	"github.com/gorilla/sessions"
)

func Signup(signupRequest model.SignupRequest, session *sessions.Session) model.APIResponse {
	var profiles []interface{}
	var resp model.APIResponse
	var err error

	// Validate signup request
	if resp := validation.ValidateSignupRequest(signupRequest); resp.StatusCode != http.StatusOK {
		return resp
	}

	// Check if license key already exists
	_, err = dao.FetchLicenseKeyProfile()
	if err == nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_LICENSE_KEY_ALREADY_EXISTS, extras.ErrLicenseKeyInUse)
		return resp
	}

	signupRequest.IsSuperAdmin = true
	// Check if admin profile with the same name already exists
	_, err = dao.FetchAdminProfile(map[string]any{"Name": signupRequest.Name})
	if err == nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_NAME_ALREADY_EXISTS, extras.ErrRecordAlreadyExists)
		return resp
	} else if err != extras.ErrNoRecordForAdmin {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	// Check if admin profile with the same email already exists
	_, err = dao.FetchAdminProfile(map[string]any{"Email": signupRequest.Email})
	if err == nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_EMAIL_ALREADY_EXISTS, extras.ErrRecordAlreadyExists)
		return resp
	} else if err != extras.ErrNoRecordForAdmin {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	// Check if admin profile with the same phone already exists
	_, err = dao.FetchAdminProfile(map[string]any{"Phone": signupRequest.Phone})
	if err == nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_PHONE_ALREADY_EXISTS, extras.ErrRecordAlreadyExists)
		return resp
	} else if err != extras.ErrNoRecordForAdmin {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	// Check if user authentication profile with the same username already exists
	_, err = dao.FetchUserAuthProfile(map[string]any{"Username": signupRequest.Username})
	if err == nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_USER_ALREADY_EXISTS, extras.ErrRecordAlreadyExists)
		return resp
	} else if err != extras.ErrNoRecordForUserAuth {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	// Hash password
	hash, err := auth.HashPassword(signupRequest.Password)
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_WHILE_HASHING, err)
		return resp
	}

	// Create user authentication profile
	var userAuthentication = model.UserAuthentication{
		Key:          util.GenerateUUID(),
		Username:     signupRequest.Username,
		Password:     hash,
		UserType:     extras.TYPE_ADMIN,
		IsActive:     true,
		IsSuperAdmin: signupRequest.IsSuperAdmin,
	}

	// If super admin, validate and set license key
	resp = CreateLicenseKey(signupRequest.LicenseKey, signupRequest.Email)
	if resp.StatusCode != http.StatusOK {
		return resp
	}
	licenseKey := resp.Data.(model.KeysTable)

	role := model.Role{
		Name: "SuperAdminRole",
	}
	resp = CreateRolePermission(role, true)
	if resp.StatusCode != http.StatusOK {
		return resp
	}

	// Create admin profile
	var admin = model.Admin{
		Key:                   util.GenerateUUID(),
		Name:                  signupRequest.Name,
		Email:                 signupRequest.Email,
		CountryCode:           signupRequest.CountryCode,
		Phone:                 signupRequest.Phone,
		UserAuthenticationKey: userAuthentication.Key,
		RoleKey:               resp.Data.([]interface{})[0].(string),
		AlreadySignedUp:       true,
	}

	var scanProfile model.ScanProfile
	scanProfile.UserAuthenticationKey = userAuthentication.Key

	// Append user authentication and admin profiles to profiles slice
	profiles = append(profiles, resp.Data.([]interface{})[1].([]interface{})...)
	profiles = append(profiles, scanProfile)
	profiles = append(profiles, userAuthentication)
	profiles = append(profiles, licenseKey)
	profiles = append(profiles, admin)

	// Save profiles
	if err := dao.SaveProfile(profiles, extras.POST); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		return resp
	}

	// Prepare success response

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Successfully signed up")
	return resp
}

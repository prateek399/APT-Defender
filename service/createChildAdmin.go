package service

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"anti-apt-backend/validation"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

func CreateChildAdmin(signupRequest model.SignupRequest) model.APIResponse {
	var resp model.APIResponse
	var profiles []interface{}
	var err error

	// Validate signup request
	if resp := validation.ValidateSignupRequest(signupRequest); resp.StatusCode != http.StatusOK {
		return resp
	}

	// Validate role key
	if signupRequest.RoleKey == "" {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_REQUIRED_FIELD_EMPTY, extras.ErrRequiredFieldEmpty)
		return resp
	}

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

	// Check if user authentication profile with the same username already exists
	_, err = dao.FetchUserAuthProfile(map[string]any{"IsSuperAdmin": true})
	if err == extras.ErrNoRecordForUserAuth {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_RECORD_NOT_FOUND, extras.ErrNoRecordForUserAuth)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	// Fetch role profile
	_, err = dao.FetchRoleProfile(map[string]any{"Key": signupRequest.RoleKey})
	if err == extras.ErrNoRecordForRole {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_RECORD_NOT_FOUND, err)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	// Encrypt password
	hash, err := auth.HashPassword(signupRequest.Password)
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_WHILE_HASHING, err)
		return resp
	}

	curTime := time.Now()
	// Create user authentication profile
	var userAuthentication = model.UserAuthentication{
		Key:          util.GenerateUUID(),
		CreatedAt:    curTime,
		Username:     signupRequest.Username,
		Password:     hash,
		UserType:     extras.TYPE_ADMIN,
		IsActive:     false,
		IsSuperAdmin: false,
	}

	// Create admin profile
	var admin = model.Admin{
		Key:                   util.GenerateUUID(),
		CreatedAt:             curTime,
		Name:                  signupRequest.Name,
		Email:                 signupRequest.Email,
		CountryCode:           signupRequest.CountryCode,
		Phone:                 signupRequest.Phone,
		UserAuthenticationKey: userAuthentication.Key,
		RoleKey:               signupRequest.RoleKey,
	}

	// Append user authentication and admin profiles to profiles slice
	profiles = append(profiles, userAuthentication)
	profiles = append(profiles, admin)

	// Save profiles
	if err := dao.SaveProfile(profiles, extras.POST); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		return resp
	}

	// Prepare success response
	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Successfully Signed Up")
	return resp
}

func UpdateAdminPersonalDetails(signupRequest model.SignupRequest, session *sessions.Session) model.APIResponse {
	var resp model.APIResponse
	var err error

	// Check if admin profile with the same email already exists
	var admin []model.Admin
	admin, err = dao.FetchAdminProfile(map[string]any{"UserAuthenticationKey": session.Values["user_id"].(string)})
	if err == extras.ErrNoRecordForAdmin {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_EMAIL_ALREADY_EXISTS, extras.ErrRecordAlreadyExists)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	if signupRequest.Name != "" {
		admin[0].Name = signupRequest.Name
	}

	if signupRequest.Email != "" {
		admin[0].Email = signupRequest.Email
	}

	var profiles []interface{}
	profiles = append(profiles, admin[0])

	// Save profiles
	if err := dao.SaveProfile(profiles, extras.PATCH); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		return resp
	}

	// change admin name in session
	session.Values["admin_name"] = signupRequest.Name

	// Prepare success response
	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Successfully saved admin personal details")
	return resp
}

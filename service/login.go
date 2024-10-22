package service

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

func Login(request model.LoginRequest, session *sessions.Session, bearer []string) model.APIResponse {
	var resp model.APIResponse
	var err error
	// Check if bearer token exists
	if bearer != nil {
		// Check session for required values
		if session.Values["refresh_token"] == nil || session.Values["exp"] == nil || session.Values["admin_role"] == nil {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
			resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_SESSION_INVALID, extras.ErrUnauthorizedUser)
			return resp
		}

		// Check if access token expired, generate new if necessary
		if session.Values["exp"].(int64) < int64(time.Now().Unix()) {
			newAccessToken, expNew, err := auth.GenerateAccessToken(session.Values["refresh_token"].(string))
			if err != nil {
				// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:eror in generating access token"))
				resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_GENERATING_ACCESS_TOKEN, err)
				return resp
			}
			session.Values["access_token"] = newAccessToken
			session.Values["exp"] = expNew
		}

		// Validate bearer token against access token
		access := session.Values["access_token"].(string)
		if err := util.AuthenticateToken(bearer, access); err != nil {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
			resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_UNAUTHORIZED_USER, extras.ErrUnauthorizedUser)
			return resp
		}

		// Fetch user authentication profile
		var userAuth []model.UserAuthentication
		userAuth, err = dao.FetchUserAuthProfile(map[string]any{"Key": session.Values["user_id"]})
		if err == extras.ErrNoRecordForUserAuth {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:no record for user"))
			resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_RECORD_NOT_FOUND, err)
			return resp
		} else if err != nil {
			resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

		// Fetch admin profile
		var admin []model.Admin
		admin, err = dao.FetchAdminProfile(map[string]any{"UserAuthenticationKey": userAuth[0].Key})
		if err == extras.ErrNoRecordForAdmin {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:no record for user"))
			resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_RECORD_NOT_FOUND, err)
			return resp
		} else if err != nil {
			resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

		// Set role permissions for admin
		resp = SetRolePermissionForAdmin(session.Values["admin_role"].(string))
		if resp.StatusCode != http.StatusOK {
			return resp
		}

		// Prepare data for successful response
		data := map[string]any{
			"name":         admin[0].Name,
			"email":        admin[0].Email,
			"phone":        admin[0].Phone,
			"organization": admin[0].Organization,
			"access":       access,
			"antiaptmenu":  resp.Data,
		}
		resp = model.NewSuccessResponse(extras.ERR_SUCCESS, data)
		return resp
	}

	// defer CreateAuditLogs(&resp, request.Username+" logged in successfully", "Authentication/Authorization", request.Username)

	licenseKey, err := dao.FetchLicenseKeyProfile()
	if err == extras.ErrNoRecordForLicenseKey {
		// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:"+err.Error()))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_NO_LICENSE_KEY_ATTACHED, extras.ErrNoLicenseKeyAttached)
		return resp
	} else if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:"+err.Error()))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	if request.Username == licenseKey.DeviceSerialId && request.Password == licenseKey.ApplianceKey {
		// Fetch user authentication profile
		var userAuth []model.UserAuthentication
		userAuth, err = dao.FetchUserAuthProfile(map[string]any{"Username": "admin"})
		if err == extras.ErrNoRecordForUserAuth {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:no record for user"))
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_RECORD_NOT_FOUND, err)
			return resp
		} else if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

		newPass := "admin1A@"
		username := "admin"
		userAuth[0].IsSuperAdmin = true
		userAuth[0].Username = username

		hash, err := auth.HashPassword(newPass)
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_WHILE_HASHING, err)
			return resp
		}
		userAuth[0].Password = hash

		if err := dao.SaveProfile([]interface{}{userAuth[0]}, extras.PATCH); err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
			return resp
		}

		resp = model.NewSuccessResponse(extras.ERR_PASSWORD_CHANGED_SUCCESSFULLY, "Password changed succesfully")
		return resp
	}

	// If bearer token doesn't exist, authenticate user
	resp = Authenticate(request, extras.TYPE_ADMIN)
	if resp.StatusCode != http.StatusOK {
		// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:"+resp.Error))
		return resp
	}

	// Generate refresh token
	refresh, err := auth.GenerateRefreshToken(request.Username)
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:error in generating refresh token"))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_GENERATING_REFRESH_TOKEN, err)
		return resp
	}

	// Generate access token
	access, exp, err := auth.GenerateAccessToken(refresh)
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:error in generating access token"))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_GENERATING_ACCESS_TOKEN, err)
		return resp
	}

	// Fetch user authentication profile
	var userAuth []model.UserAuthentication
	userAuth, err = dao.FetchUserAuthProfile(map[string]any{"Username": request.Username})
	if err == extras.ErrNoRecordForUserAuth {
		// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:no record for user"))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_RECORD_NOT_FOUND, err)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	// Fetch admin profile
	var admin []model.Admin
	admin, err = dao.FetchAdminProfile(map[string]any{"UserAuthenticationKey": userAuth[0].Key})
	if err == extras.ErrNoRecordForAdmin {
		// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:no record for user"))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_RECORD_NOT_FOUND, err)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	if licenseKey.ExpiryTime.Before(time.Now()) {
		logger.LoggerFunc("error", logger.LoggerMessage("sysLog:License Exipired!!!"))

		// Fetch Roles corresponding to this user
		roleAndAction, err := dao.FetchRoleAndActionProfile(map[string]any{"RoleKey": admin[0].RoleKey})
		if err != nil {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:"+err.Error()))
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

		for i := range roleAndAction {
			roleAndAction[i].Permission = extras.READONLY
		}

		// fmt.Println("Length: -------> ", len(roleAndAction))

		resp = SetPermissions(roleAndAction)
		if resp.StatusCode != http.StatusOK {
			return resp
		}
	}

	// Set role permissions for admin
	resp = SetRolePermissionForAdmin(admin[0].RoleKey)
	if resp.StatusCode != http.StatusOK {
		// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:"+resp.Error))
		return resp
	}

	// Store session values
	session.Values["user_id"] = userAuth[0].Key
	session.Values["exp"] = exp
	session.Values["refresh_token"] = refresh
	session.Values["access_token"] = access
	session.Values["admin_role"] = admin[0].RoleKey
	session.Values["admin_name"] = admin[0].Name
	session.Values["organization_id"] = "xyz"

	// Prepare data for successful response
	data := map[string]any{
		"name":        admin[0].Name,
		"email":       admin[0].Email,
		"phone":       admin[0].Phone,
		"antiaptmenu": resp.Data,
		"access":      access,
	}
	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("sysLog:%s logged in successfully", admin[0].Name)))
	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, data)
	return resp
}

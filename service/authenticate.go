package service

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

func Authenticate(loginRequest model.LoginRequest, userType int) model.APIResponse {
	var resp model.APIResponse

	var userAuth []model.UserAuthentication
	var err error

	userAuth, err = dao.FetchUserAuthProfile(map[string]any{"Username": loginRequest.Username})
	if err == extras.ErrNoRecordForUserAuth {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_RECORD_NOT_FOUND, err)
		return resp
	} else if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	if userType == extras.TYPE_ADMIN {
		var originOS, originIP string

		originOS = runtime.GOOS
		originIP = util.GetLocalIP()

		retry := &RetryAttemptLimiter{Username: userAuth[0].Username, UserType: userAuth[0].UserType, Attempts: userAuth[0].InvalidAttempt, HoldTime: userAuth[0].HoldingDatetime.Unix()}

		//checking if max time limit after 5 consecutive invalid attempts is reached or not
		if !retry.checkUnderTimeout(userAuth) {
			//User is Valid
			if auth.ComparePasswordHash(loginRequest.Password, userAuth[0].Password) {
				retry.clearTimeout(userAuth)
				data := map[string]any{"User": userAuth, "Os": originOS, "Ip": originIP}
				resp = model.NewSuccessResponse(extras.ERR_SUCCESS, data)
				return resp

			} else { //Invalid Password attempt
				retry.InvalidAttempted(userAuth)
				resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_LOGIN_FAIL, extras.ErrInvalidPassword)
				return resp
			}

		} else { //Max login limit reached.
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_MAXIMUM_LOGIN_LIMIT_REACHED, extras.ErrMaxLoginLimitReached)
			return resp
		}
	}

	resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_USER_TYPE, err)
	return resp
}

func GetCurUsr(c *gin.Context) (string, model.APIResponse) {
	var resp model.APIResponse
	session, err := auth.Store.Get(c.Request, "sessionid")
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage(err.Error()))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SESSION_INVALID, err)
		return extras.EMPTY_STRING, resp
	}

	if _, found := session.Values["admin_name"].(string); !found {
		// logger.LoggerFunc("warn", logger.LoggerMessage(extras.ERR_SESSION_INVALID))
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_SESSION_INVALID, extras.ErrUnauthorizedUser)
		return extras.EMPTY_STRING, resp
	}
	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, nil)
	return session.Values["admin_name"].(string), resp
}

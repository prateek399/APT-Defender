package auth

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var resp model.APIResponse
		var err error

		session, _ := Store.Get(ctx.Request, "sessionid")
		// ctx.Header("Origin", ctx.ClientIP())
		if session.Values["refresh_token"] == nil {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
			resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_SESSION_INVALID, extras.ErrUnauthorizedUser)
			ctx.JSON(resp.StatusCode, resp)
			ctx.Abort()
			return
		}

		if session.Values["exp"].(int64) < int64(time.Now().Unix()) {
			accessToken, expNew, err := GenerateAccessToken(session.Values["refresh_token"].(string))
			if err != nil {
				// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:error generating access token"))
				resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_GENERATING_ACCESS_TOKEN, err)
				ctx.JSON(resp.StatusCode, resp)
				ctx.Abort()
				return
			}

			session.Values["access_token"] = accessToken
			session.Values["exp"] = expNew
			if err := session.Save(ctx.Request, ctx.Writer); err != nil {
				// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
				resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_SESSION_NOT_SAVED, err)
				ctx.JSON(resp.StatusCode, resp)
				ctx.Abort()
				return
			}
		}

		access, err := jwt.Parse(session.Values["access_token"].(string), func(t *jwt.Token) (interface{}, error) {
			return []byte(Access_key), nil
		})
		if err != nil {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
			resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_UNAUTHORIZED_USER, err)
			ctx.JSON(resp.StatusCode, resp)
			ctx.Abort()
			return
		}

		refresh, err := jwt.Parse(session.Values["refresh_token"].(string), func(t *jwt.Token) (interface{}, error) {
			return []byte(Refresh_key), nil
		})
		if err != nil {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:error in generating refresh token"))
			resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_GENERATING_REFRESH_TOKEN, err)
			ctx.JSON(resp.StatusCode, resp)
			ctx.Abort()
			return
		}

		if !access.Valid || !refresh.Valid {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
			resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_UNAUTHORIZED_USER, extras.ErrUnauthorizedUser)
			ctx.JSON(resp.StatusCode, resp)
			ctx.Abort()
			return
		}

		bearer := strings.Split(ctx.GetHeader("Authorization"), " ")
		if err = util.AuthenticateToken(bearer, session.Values["access_token"].(string)); err != nil {
			// logger.LoggerFunc("error", logger.LoggerMessage("sysLog:not authorized"))
			resp = model.NewErrorResponse(http.StatusUnauthorized, extras.ERR_UNAUTHORIZED_USER, extras.ErrUnauthorizedUser)
			ctx.JSON(resp.StatusCode, resp)
			ctx.Abort()
			return
		}

		// logger.LoggerFunc("info", logger.LoggerMessage("sysLog:user is authorized"))
		ctx.Next()
	}
}

// func LicenseKeyExpiryMiddleware() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {

// 	}
// }

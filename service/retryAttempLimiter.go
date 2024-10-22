package service

import (
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"time"
)

type RetryAttemptLimiter struct {
	Username string
	UserType int
	Attempts int
	HoldTime int64
}

func (retry *RetryAttemptLimiter) checkUnderTimeout(userAuth []model.UserAuthentication) bool {
	if retry.Attempts >= extras.INVALID_PASSWORD_LIMIT {
		fiveMin := time.Now().Add(5*time.Minute).Unix() - time.Now().Unix()
		timeGap := time.Now().Unix() - retry.HoldTime
		if timeGap <= fiveMin {
			return true
		}
		retry.clearTimeout(userAuth)
		return false
	}
	return false
}

func (retry *RetryAttemptLimiter) clearTimeout(userAuth []model.UserAuthentication) {
	userAuth[0].InvalidAttempt = 0
	userAuth[0].HoldingDatetime = time.Now()

	if err := dao.SaveProfile([]interface{}{userAuth[0]}, extras.PATCH); err != nil {
		return
	}
}

func (retry *RetryAttemptLimiter) InvalidAttempted(userAuth []model.UserAuthentication) {
	userAuth[0].InvalidAttempt++
	userAuth[0].HoldingDatetime = time.Now()

	if err := dao.SaveProfile([]interface{}{userAuth[0]}, extras.PATCH); err != nil {
		return
	}
}

package service

import (
	"anti-apt-backend/config"
	"anti-apt-backend/model"
	"net/http"
	"time"
)

func CreateAuditLogs(resp *model.APIResponse, message string, auditType string, adminName string) error {
	defer func() {
		if err := recover(); err != nil {
			// slog.Error("CREATE AUDIT LOGS ERROR: ", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		message = resp.Message
	}

	auditLog := model.AuditTable{
		AdminName: adminName,
		Message:   message,
		AuditType: auditType,
		TimeStamp: time.Now(),
	}

	if err := config.Db.Model(&model.AuditTable{}).Create(&auditLog).Error; err != nil {
		// slog.Error("Failed to create audit log: ", err)
		return err
	}

	return nil
}

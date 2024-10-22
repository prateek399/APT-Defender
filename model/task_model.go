package model

import (
	"database/sql"
	"time"
)

type FileOnDemand struct {
	Id                int                   `gorm:"primaryKey" json:"task_id"`
	FileName          string                `json:"filename"`
	ContentType       string                `json:"content_type"`
	SubmittedTime     time.Time             `json:"submitted_time"`
	FinishedTime      sql.NullTime          `json:"finished_time"`
	SubmittedBy       string                `json:"submitted_by"`
	FileCount         int                   `json:"file_count"`
	Rating            string                `json:"rating"`
	Score             float32               `json:"score"`
	FinalVerdict      string                `json:"final_verdict"`
	Status            string                `json:"status"`
	Comments          string                `json:"comments"`
	FromDevice        bool                  `json:"input_type"`
	OverriddenVerdict bool                  `json:"overridden_verdict"`
	OverriddenBy      string                `json:"overridden_by"`
	OsSupported       string                `json:"os_supported"`
	Md5               string                `json:"md5"`
	SHA               string                `json:"sha"`
	SHA256            string                `json:"sha256"`
	ClientIp          string                `json:"client_ip"`
	TaskLiveAnalysis  TaskLiveAnalysisTable `gorm:"foreignKey:Id;constraint:OnDelete:CASCADE" json:"task_live_analysis"`
	TaskFinished      TaskFinishedTable     `gorm:"foreignKey:Id;constraint:OnDelete:CASCADE" json:"task_finished"`
	TaskDuplicate     TaskDuplicateTable    `gorm:"foreignKey:Id;constraint:OnDelete:CASCADE" json:"task_duplicate"`
}

type UrlOnDemand struct {
	Id                int          `gorm:"primaryKey" json:"task_id"`
	UrlName           string       `json:"urlname"`
	SubmittedTime     time.Time    `json:"submitted_time"`
	FinishedTime      sql.NullTime `json:"finished_time"`
	SubmittedBy       string       `json:"submitted_by"`
	UrlCount          int          `json:"url_count"`
	Rating            string       `json:"rating"`
	FinalVerdict      string       `json:"final_verdict"`
	Status            string       `json:"status"`
	Comments          string       `json:"comments"`
	FromDevice        bool         `json:"input_type"`
	OverriddenVerdict bool         `json:"overridden_verdict"`
	OverriddenBy      string       `json:"overridden_by"`
	OsSupported       string       `json:"os_supported"`
}

type TaskLiveAnalysisTable struct {
	TaskLiveAnalysisId int    `gorm:"primaryKey" json:"task_live_analysis_id"`
	Id                 int    `json:"task_id"` // Foreign Key
	Status             string `json:"status"`
	SandboxId          int    `json:"sandbox_id"`
	QueueRetryCount    int    `json:"queue_retry_count"`
	RunningRetryCount  int    `json:"running_retry_count"`
	SandboxRetryCount  int    `json:"sandbox_retry_count"`
	LogQueueFailed     bool   `json:"log_queue_failed"`
	Md5                string `json:"md5"`
	SHA                string `json:"sha"`
	SHA256             string `json:"sha256"`
}

type TaskFinishedTable struct {
	TaskFinishedId int  `gorm:"primaryKey" json:"task_finished_id"`
	Id             int  `json:"task_id"` // Foreign Key
	SandboxId      int  `json:"sandbox_id"`
	Aborted        bool `json:"aborted"`
}

type TaskDuplicateTable struct {
	TaskDuplicateId int    `gorm:"primaryKey" json:"task_duplicate_id"`
	Id              int    `json:"task_id"`
	Md5             string `json:"md5"`
	SHA             string `json:"sha"`
	SHA256          string `json:"sha256"`
}

type AuditTable struct {
	AuditId   int       `gorm:"primaryKey" json:"audit_id"`
	AdminName string    `json:"admin_name"`
	AuditType string    `json:"audit_type"`
	Message   string    `json:"message"`
	TimeStamp time.Time `json:"timestamp"`
}

// type DeviceInfo struct {

// }

type FileHashes struct {
	Id     int    `json:"id" gorm:"primaryKey"`
	Sha1   string `json:"sha1"`
	Sha256 string `json:"sha256"`
	Md5    string `json:"md5"`
}

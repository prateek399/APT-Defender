package model

import (
	"net/http"
	"time"
)

type APIResponse struct {
	Data       interface{} `json:"data,omitempty"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message,omitempty"`
	Error      string      `json:"error,omitempty"`
}

func NewSuccessResponse(successMessage string, data interface{}) APIResponse {
	return APIResponse{
		StatusCode: http.StatusOK,
		Data:       data,
		Message:    successMessage,
	}
}

func NewErrorResponse(statusCode int, errorMessage string, err error) APIResponse {
	return APIResponse{
		StatusCode: statusCode,
		Message:    errorMessage,
		Error:      err.Error(),
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangePass struct {
	Password           string `json:"password"`
	NewPassword        string `json:"new_password"`
	ConfirmNewPassword string `json:"confirm_new_password"`
}

type SignupRequest struct {
	Name            string `json:"name"`
	OrgName         string `json:"organization"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	CountryCode     string `json:"country_code"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	IsSuperAdmin    bool   `json:"is_super_admin"`
	LicenseKey      string `json:"license_key"` // Foreign Key from Keys Table
	RoleKey         string `json:"role"`        // Foreign Key from Roles Table
}

type ShowFeatures struct {
	Id         string         `json:"id"`
	Title      string         `json:"title"`
	MessageID  string         `json:"messageId"`
	Permission int            `json:"permission"`
	Icon       string         `json:"icon,omitempty"`
	Typeof     string         `json:"type,omitempty"`
	Path       string         `json:"path,omitempty"`
	Position   int            `json:"position"`
	Children   []ShowFeatures `json:"children,omitempty"`
}

type Config struct {
	// LicenseKeys         []KeysTable          `yaml:"license_keys"`
	UserAuthentications []UserAuthentication `yaml:"user_authentications"`
	Admins              []Admin              `yaml:"admins"`
	Roles               []Role               `yaml:"roles"`
	Features            []Feature            `yaml:"features"`
	RoleAndActions      []RoleAndAction      `yaml:"role_and_actions"`
	ScanProfiles        []ScanProfile        `yaml:"scan_profiles"`
	Devices             []Device             `yaml:"devices"`
	BackupConfigTime    BackupTime           `yaml:"backup_config_time"`
	RestoreConfigTime   RestoreTime          `yaml:"restore_config_time"`
	ViewLogs            ViewLog              `yaml:"view_logs"`
}

type TaskConfig struct {
	FileOnDemands      map[string]FileOnDemand `yaml:"file_on_demands"`
	UrlOnDemands       map[string]UrlOnDemand  `yaml:"url_on_demands"`
	OverriddenVerdicts []OverriddenVerdict     `yaml:"overridden_verdicts"`
}

type ViewLog struct {
	Enabled bool `json:"enabled"`
}

type BackupTime struct {
	Time time.Time `json:"time"`
}

type RestoreTime struct {
	Time time.Time `json:"time"`
}

type LicenseKeyConfig struct {
	LicenseKey KeysTable `yaml:"license_key"`
}

type KeysTable struct {
	ShippingDate    time.Time `json:"shipping_date"`
	DeviceSerialId  string    `json:"device_serial_id"`
	ApplianceKey    string    `json:"key_value"`
	ExpiryTime      time.Time `json:"expiry_time"`
	ModelNo         string    `json:"model_no"`
	RegisteredEmail string    `json:"registered_email"`
}

type UserAuthentication struct {
	Key             string    `json:"key"`
	CreatedAt       time.Time `json:"created_at"`
	Username        string    `json:"username"`
	Password        string    `json:"password"`
	UserType        int       `json:"user_type"`
	InvalidAttempt  int       `json:"invalid_attempt"`
	HoldingDatetime time.Time `json:"holding_datetime"`
	IsActive        bool      `json:"is_active"`
	IsSuperAdmin    bool      `json:"is_super_admin"`
}

type Admin struct {
	Key                   string    `json:"key"`
	CreatedAt             time.Time `json:"created_at"`
	Name                  string    `json:"name"`
	Email                 string    `json:"email"`
	CountryCode           string    `json:"country_code"`
	Phone                 string    `json:"phone"`
	Organization          string    `json:"organization"`
	UserAuthenticationKey string    `json:"user_authentication"` // Foreign Key from User Authentication Table
	AlreadySignedUp       bool      `json:"already_signed_up"`
	RoleKey               string    `json:"role"`
}

type Role struct {
	Key         string    `json:"key"`
	CreatedAt   time.Time `json:"created_at"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Custom      int       `json:"custom"`
}

type RoleAndAction struct {
	Key        string    `json:"key"`
	CreatedAt  time.Time `json:"created_at"`
	RoleKey    string    `json:"role"`
	FeatureKey string    `json:"feature"`
	Permission int       `json:"permission"`
}

type Feature struct {
	Key                string `json:"id"`
	Title              string `json:"title"`
	SelfType           int    `json:"self_type"`
	GrandParentKey     string `json:"grand_parent"`
	ParentKey          string `json:"parent"`
	IsHide             int    `json:"is_hide"`
	Position           int    `json:"position"`
	PossiblePermission int    `json:"possible_permission"`
	MessageID          string `json:"messageId"`
	Icon               string `json:"icon"`
	Type               string `json:"type"`
	Path               string `json:"path"`
}

// type FileType struct {
// 	Key  string `json:"key"`
// 	Name string `json:"name"`
// 	Exts string `json:"exts"`
// }

type ScanProfile struct {
	PdfFile               bool   `json:"pdf"`
	OfficeFile            bool   `json:"office"`
	ExeFile               bool   `json:"exe"`
	FlashFile             bool   `json:"flash"`
	WebFile               bool   `json:"web"`
	CompressedArchieve    bool   `json:"archive"`
	AudioFile             bool   `json:"audio"`
	VideoFile             bool   `json:"video"`
	TextFile              bool   `json:"text"`
	ScriptFile            bool   `json:"script"`
	UserAuthenticationKey string `json:"userauthenticatin,omitempty"` // User Authentication Key for J
}

type OverrideVerdictRequest struct {
	Type    string `json:"type"`
	JobID   int    `json:"job_id"`
	Verdict string `json:"verdict"`
	Comment string `json:"comment"`
}

type OverriddenVerdict struct {
	JobID           string `json:"job_id"`
	Filename        string `json:"filename,omitempty"`
	URL             string `json:"url,omitempty"`
	SubmittedBy     string `json:"submitted_by"`
	SubmittedTime   string `json:"submitted_time"`
	OriginalVerdict string `json:"original_verdict"`
	FinalVerdict    string `json:"final_verdict"`
	Comment         string `json:"comment"`
	UpdatedBy       string `json:"updated_by"`
	UpdatedAt       string `json:"updated_at"`
}

type LogReport struct {
	JobID      string `json:"job_id"`
	Name       string `json:"name"`
	CuckooLogs string `json:"cuckoo_logs"`
}

type Device struct {
	Key             string    `json:"key"`
	CreatedAt       time.Time `json:"created_at"`
	DeviceName      string    `json:"device_name"`
	ProductCategory string    `json:"product_category"`
	SerialNumber    string    `json:"serial_number"`
	IpAddress       string    `json:"ip_address"`
	Email           string    `json:"email"`
	MobileNumber    string    `json:"mobile_number"`
	Country         string    `json:"country"`
	State           string    `json:"state"`
	City            string    `json:"city"`
}

type TroubleshootRequest struct {
	Name      string `json:"name"`
	Iface     string `json:"iface"`
	Type      int    `json:"type"`
	Port      string `json:"port"`
	IpAddress string `json:"ip_address"`
	Command   string `json:"command"`
}

type Verdict string

const (
	Clean      Verdict = "Clean"
	LowRisk    Verdict = "Low"
	MediumRisk Verdict = "Medium"
	HighRisk   Verdict = "High"
	Critical   Verdict = "Critical"
	Unknown    Verdict = "Unknown"
)

type CreateTaskFileResponse struct {
	TaskId int `json:"task_id"`
}

type TaskStatus string

var (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusReported  TaskStatus = "reported"
)

type Task struct {
	Category       string        `json:"category"`
	Machine        interface{}   `json:"machine"`
	Errors         []interface{} `json:"errors"`
	Target         string        `json:"target"`
	Package        interface{}   `json:"package"`
	SampleID       interface{}   `json:"sample_id"`
	Guest          interface{}   `json:"guest"`
	Custom         interface{}   `json:"custom"`
	Owner          string        `json:"owner"`
	Priority       int64         `json:"priority"`
	Platform       interface{}   `json:"platform"`
	Options        interface{}   `json:"options"`
	Status         TaskStatus    `json:"status"`
	EnforceTimeout bool          `json:"enforce_timeout"`
	Timeout        int64         `json:"timeout"`
	Memory         bool          `json:"memory"`
	Tags           []string      `json:"tags"`
	ID             int           `json:"id"`
	AddedOn        string        `json:"added_on"`
	CompletedOn    interface{}   `json:"completed_on"`
}

type Report struct {
	Info     ReportInfo     `json:"info"`
	Target   ReportTarget   `json:"target"`
	Behavior ReportBehavior `json:"behavior"`
	Debug    ReportDebug    `json:"debug"`
}

type ReportInfo struct {
	Added    float64     `json:"added"`
	Started  float64     `json:"started"`
	Duration int         `json:"duration"`
	Ended    float64     `json:"ended"`
	Owner    string      `json:"owner"`
	Score    float32     `json:"score"`
	Id       int         `json:"id"`
	Category string      `json:"category"`
	Machine  interface{} `json:"machine"`
}

type ReportTarget struct {
	Category string     `json:"category"`
	File     ReportFile `json:"file"`
}

type ReportFile struct {
	Sha1   string   `json:"sha1"`
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Sha256 string   `json:"sha256"`
	Urls   []string `json:"urls"`
	Size   int64    `json:"size"`
	Md5    string   `json:"md5"`
}

type ReportBehavior struct {
	Generic   []ReportGeneric   `json:"generic"`
	Processes []ReportProcesses `json:"processes"`
}

type ReportGeneric struct {
	ProcessName string  `json:"process_name"`
	ProcessId   int     `json:"pid"`
	FirstSeen   float64 `json:"first_seen"`
}

type ReportProcesses struct {
	ProcessPath   string  `json:"process_path"`
	Track         bool    `json:"track"`
	ProcessId     int     `json:"pid"`
	ProcessName   string  `json:"process_name"`
	ModulesLength int     `json:"modules_length"`
	FirstSeen     float64 `json:"first_seen"`
	Type          string  `json:"type"`
}

type ReportDebug struct {
	Logs   interface{} `json:"logs"`
	Cuckoo interface{} `json:"cuckoo"`
	Errors interface{} `json:"errors"`
	Action interface{} `json:"action"`
}

type Machine struct {
	Status           interface{} `json:"status"`
	Locked           bool        `json:"locked"`
	Name             string      `json:"name"`
	ResultserverIP   string      `json:"resultserver_ip"`
	IP               string      `json:"ip"`
	Tags             []string    `json:"tags"`
	Label            string      `json:"label"`
	LockedChangedOn  interface{} `json:"locked_changed_on"`
	Platform         string      `json:"platform"`
	Snapshot         interface{} `json:"snapshot"`
	Interface        interface{} `json:"interface"`
	StatusChangedOn  interface{} `json:"status_changed_on"`
	ID               int64       `json:"id"`
	ResultserverPort int64       `json:"resultserver_port"`
}

type JobInfo struct {
	Summary  JobSummary `json:"summary"`
	Details  JobDetail  `json:"details"`
	Filename string     `json:"filename"`
}

type JobSummary struct {
	JobID         string `json:"jobID"`
	Status        string `json:"status"`
	ReceivedTime  string `json:"receivedTime"`
	RatedBy       string `json:"ratedBy"`
	SubmitType    string `json:"submitType"`
	VmScanTimeout int    `json:"vmScanTimeout"`
	Rating        string `json:"rating"`
	FinalVerdict  string `json:"finalVerdict"`
}

type JobDetail struct {
	Filename      string `json:"filename"`
	ScanStartTime string `json:"scanStartTime"`
	ScanEndTime   string `json:"scanEndTime"`
	TotalScanTime int    `json:"totalScanTime"`
	FileType      string `json:"fileType"`
	// FileSize        int    `json:"fileSize"`
	MD5             string `json:"md5"`
	SHA1            string `json:"sha1"`
	SHA256          string `json:"sha256"`
	SubmittedBy     string `json:"submittedBy"`
	SubmitDevice    string `json:"submittedFrom"`
	SubmittedDevice string `json:"submittedTo"`
	VM              string `json:"vm"`
	VMReason        string `json:"vmReason"`
}

type UrlJobInfo struct {
	Summary JobSummary   `json:"summary"`
	Details UrlJobDetail `json:"details"`
	Url     string       `json:"url"`
}

type UrlJobDetail struct {
	Url             string `json:"url"`
	Type            string `json:"type"`
	ScanStartTime   string `json:"scanStartTime"`
	ScanEndTime     string `json:"scanEndTime"`
	TotalScanTime   int    `json:"totalScanTime"`
	SubmittedBy     string `json:"submittedBy"`
	SubmitDevice    string `json:"submittedFrom"`
	SubmittedDevice string `json:"submittedTo"`
	VM              string `json:"vm"`
	VMReason        string `json:"vmReason"`
}

type HaDeviceInfo struct {
	SerialNo string `json:"serial_no"`
	ModelNo  string `json:"model_no"`
	Password string `json:"password"`
}

type LastSyncedAt struct {
	LastSyncedAt string `json:"last_synced_at"`
}

type DeviceSpecification struct {
	ModelNo          string `json:"modelNo"`
	SerialNo         string `json:"serialNo"`
	RegisteredEmail  string `json:"registeredEmail"`
	OsVersion        string `json:"osVersion"`
	ApplianceUpSince string `json:"applianceUpSince"`
	LicenseValidTill string `json:"licenseValidTill"`
}

type DeviceConfigFile struct {
	DeviceName       string `json:"device_name"`
	DeviceID         string `json:"device_id"`
	ApplianceKey1yr  string `json:"appliance_key_1yr"`
	ApplianceKey3yr  string `json:"appliance_key_3yr"`
	ApplianceKey5yr  string `json:"appliance_key_5yr"`
	DeviceMAC        string `json:"device_mac"`
	ModelName        string `json:"model_name"`
	OrganizationName string `json:"organization_name"`
}

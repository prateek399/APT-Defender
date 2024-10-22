package extras

import (
	"fmt"
	"time"
)

const (
	EMPTY_STRING                          = ""
	ERR_SUCCESS                           = "Success"
	ERR_FROM_CLIENT_SIDE                  = "Error from client side"
	ERR_FROM_SERVER_SIDE                  = "Error from server side"
	ERR_SESSION_INVALID                   = "Session invalid"
	ERR_SESSION_NOT_SAVED                 = "Session not saved"
	ERR_IN_FETCHING_DATA                  = "Error while fetching data"
	ERR_IN_SAVING_DATA                    = "Error while saving data"
	ERR_REQUIRED_FIELD_EMPTY              = "Required field empty"
	ERR_PASSWORDS_DO_NOT_MATCH            = "Passwords do not match"
	ERR_WHILE_HASHING                     = "Error while hashing"
	ERR_EMAIL_FOUND_INVALID               = "Email found invalid"
	ERR_GENERATING_REFRESH_TOKEN          = "Error while generating refresh token"
	ERR_GENERATING_ACCESS_TOKEN           = "Error while generating access token"
	ERR_UNAUTHORIZED_USER                 = "Unauthorized user"
	ERR_INCORRECT_PASSWORD                = "Incorrect password"
	ERR_USER_INACTIVE                     = "User is not active"
	ERR_INVALID_USER_TYPE                 = "Invalid user type"
	ERR_LOGIN_FAIL                        = "Login failed, please try again"
	ERR_MAXIMUM_LOGIN_LIMIT_REACHED       = "Maximum login limit reached"
	ERR_INVALID_LICENSE_KEY               = "Invalid license key"
	ERR_RECORD_NOT_FOUND                  = "Record not found"
	ERR_LICENSE_KEY_ALREADY_EXISTS        = "License key already exists"
	ERR_USER_ALREADY_EXISTS               = "User already exists"
	ERR_NAME_ALREADY_EXISTS               = "Name already exists"
	ERR_EMAIL_ALREADY_EXISTS              = "Email already exists"
	ERR_PHONE_ALREADY_EXISTS              = "Phone already exists"
	ERR_ROLE_ALREADY_EXISTS               = "Role already exists"
	ERR_INVALID_COUNTRY_CODE              = "Invalid country code"
	ERR_INVALID_PHONE_NUMBER              = "Invalid phone number"
	ERR_INVALID_IP_ADDRESS                = "Invalid ip address"
	ERR_LICENSE_KEY_IN_USE                = "License key in use"
	ERR_FILE_TOO_LARGE                    = "File too large"
	ERR_WHILE_ANALYSING                   = "error while analysis"
	ERR_STATUS_IS_NOT_PENDING             = "status is not pending"
	ERR_STILL_ANALYSING                   = "still analysing the file"
	ERR_ALREADY_OVERRIDE                  = "already overridden the final verdict"
	ERR_INVALID_VERDICT                   = "invalid action (only allow or block is allowed)"
	ERR_MODEL_NAME_INVALID                = "model name invalid"
	ERR_INVALID_JOB_ID                    = "invalid job id"
	ERR_INVALID_TYPE_IN_OVERRIDE          = "invalid type in override (only file or url is allowed)"
	ERR_INVALID_CONTENT_TYPE              = "invalid content type"
	ERR_SERIAL_NUMBER_ALREADY_EXISTS      = "Serial number already exists"
	ERR_NON_EDITABLE_FIELD                = "non editable field"
	ERR_INVALID_TYPE_IN_TROUBLESHOOT      = "invalid type in troubleshoot"
	ERR_FILE_IS_MALWARE                   = "file is malware"
	ERR_FILE_NOT_FOUND                    = "file not found"
	ERR_SERIAL_NUMBER_LENGTH_NOT_IN_RANGE = "serial number length not in range"
	ERR_INVALID_NAME_FORMAT               = "invalid name format"
	ERR_EXTENSION_NOT_SUPPORTED           = "extension not supported"
	ERR_WHILE_PARSING_CONTENT             = "error while parsing content"
	ERR_INVALID_ACTION_TYPE               = "invalid action type"
	ERR_INVALID_DATA_TYPE                 = "invalid data type"
	ERR_FILE_NOT_SUPPORTED                = "file not supported"
	ERR_GETTING_RAM_INFO                  = "error getting RAM info"
	ERR_IN_GETTING_DISK_INFO              = "error getting disk info"
	ERR_WHILE_PARSING_URL                 = "error while parsing URL"
	ERR_INVALID_URL_FOUND                 = "invalid URL found"
	ERR_NO_LICENSE_KEY_ATTACHED           = "License Key not found, Signup to activate License Key"
	ERROR_IN_DECRYPTING_KEY               = "error in decypting key"
	ERR_INVALID_DATE_FORMAT               = "invalid date format"
	ERR_PASSWORD_CHANGED_SUCCESSFULLY     = "password changed successfully"
	ERR_IN_FETCHING_SANDBOX_TASKS         = "Error while fetching sandbox tasks"
)

const (
	TYPE_ADMIN = 1
	TYPE_USER  = 2
)

const (
	INVALID_PASSWORD_LIMIT = 5
)

const (
	ROOT_DATA_DEVICE_CONFIG          = "/data/device_config"
	CONFIG_BASE_PATH                 = "/var/www/html/web/backend"
	DATABASE_PATH                    = "/var/www/html/web/database/"
	TEMP_BUILD_PATH                  = "/var/www/html/web/"
	FIRMWARE_FILE_PATH               = "/var/log/firmware/"
	CONFIG_FILE_NAME                 = "/var/www/html/web/database/config.yaml"
	TASK_CONFIG_FILE_NAME            = "/var/www/html/web/database/task_config.yaml"
	INTERFACE_CONFIG_FILE_NAME       = "/var/www/html/web/database/interface_config.yaml"
	LICENSEKEY_FILE_PATH             = "/var/www/html/web/database/licensekey.yaml"
	OLD_CONFIG_FILE_NAME             = "/var/www/html/web/database/old_config.yaml"
	OLD_TASK_CONFIG_FILE_NAME        = "/var/www/html/web/database/old_task_config.yaml"
	OLD_INTERFACE_CONFIG_FILE_NAME   = "/var/www/html/web/database/old_interface_config.yaml"
	MERGED_CONFIG_FILE_NAME          = "/var/www/html/web/database/ha/merged_config.yaml"
	FACTORY_DEFAULT_CONFIG_FILE_NAME = "/var/www/html/data/database/factory_default.yaml"
	HA_STATE_FILE                    = "/var/www/html/data/hastate"
	CUCKOO_CONF_FILE_PATH            = "/home/wijungle/.cuckoo/conf/cuckoo.conf"
	PLATFORM_FILE_NAME               = "/etc/os_support"
	LOCK_TIME_OUT                    = 3 * time.Second
	TEMP_MALICIOUS_URLS_FILE         = "/var/www/html/web/database/temp_malicious_urls.csv"
	DEVICE_REBOOTED_FLAG_PATH        = "/etc/device_rebooted"
	SANDBOX_FILE_PATHS               = "/var/www/html/web/database/files/"
	LOGGING_ENABLED_FLAG_PATH        = "/etc/logging_enabled"
	TASK_LOGS_PATH                   = "/home/wijungle/task_logs/"
	REPORT_DOWNLOADS_PATH            = "/log/apt/reports/"
	DATA_PATH                        = "/var/www/html/data/"
	IGNORE_EXTENSIONS_FILE_PATH      = "/var/www/html/data/ignore_extensions.txt"
	MAX_SANDBOX_TASKS_FILE_PATH      = "/var/www/html/data/max_sandbox_tasks"
	TIMEOUT_FILE_PATH                = "/var/www/html/data/timeout"
)

var (
	ErrNoRecordForUserAuth      = fmt.Errorf(`no record match for user`)
	ErrNoRecordForAdmin         = fmt.Errorf(`no record match for user`)
	ErrNoRecordForLicenseKey    = fmt.Errorf(`no record match for license key`)
	ErrNoRecordForOrganization  = fmt.Errorf(`no record match for organization`)
	ErrNoRecordForRole          = fmt.Errorf(`no record match for role`)
	ErrNoRecordForRoleAndAction = fmt.Errorf(`no record match for role and action`)
	ErrNoRecordForFeature       = fmt.Errorf(`no record match for feature`)
	ErrNoRecordForScanProfile   = fmt.Errorf(`no record match for scan profiles`)
	ErrNoRecordForFileOnDemand  = fmt.Errorf(`no record match for file on demand`)
	ErrNoRecordForUrlOnDemand   = fmt.Errorf(`no record match for url on demand`)
	ErrNoRecordForOverridden    = fmt.Errorf(`no record match for overridden verdict`)
	ErrNoRecordForLogReport     = fmt.Errorf(`no record match for log report`)
	ErrNoRecordForDevice        = fmt.Errorf(`no record match for device`)
)

var (
	ErrUnauthorizedUser             = fmt.Errorf("unauthorized user: Access to this resource is restricted")
	ErrRequiredFieldEmpty           = fmt.Errorf("required field is empty")
	ErrPasswordDoNotMatch           = fmt.Errorf("password and confirm password do not match")
	ErrInvalidEmailFound            = fmt.Errorf("invalid email found")
	ErrRecordAlreadyExists          = fmt.Errorf("record already exists")
	ErrInvalidCountryCode           = fmt.Errorf("invalid country code")
	ErrInvalidPhoneNumber           = fmt.Errorf("invalid phone number")
	ErrInvalidIPAddress             = fmt.Errorf("invalid ip address")
	ErrMaxLoginLimitReached         = fmt.Errorf("maximum login limit reached")
	ErrInvalidPassword              = fmt.Errorf("invalid password")
	ErrInvalidLicenseKey            = fmt.Errorf("invalid license key")
	ErrLicenseKeyInUse              = fmt.Errorf("license key in use")
	ErrFileTooLarge                 = fmt.Errorf("file too large")
	ErrAlreadyOverridden            = fmt.Errorf("already overridden the final verdict")
	ErrInvalidVerdict               = fmt.Errorf("invalid action")
	ErrStatusIsNotPending           = fmt.Errorf("status is not pending")
	ErrModelNameInvalid             = fmt.Errorf("model name invalid")
	ErrInvalidJobId                 = fmt.Errorf("invalid job id")
	ErrReportNotGenerated           = fmt.Errorf("report not generated yet")
	ErrInvalidTypeInOverride        = fmt.Errorf("invalid type in override")
	ErrInvalidContentType           = fmt.Errorf("invalid content type")
	ErrNonEditableFieldFound        = fmt.Errorf("non editable field found")
	ErrInvalidTypeInTroubleshoot    = fmt.Errorf("invalid type in troubleshoot")
	ErrInvalidCommandInTroubleshoot = fmt.Errorf("invalid command in troubleshoot")
	ErrFileIsMalware                = fmt.Errorf("file is malware")
	ErrInvalidActionType            = fmt.Errorf("invalid action type")
	ErrInvalidFieldFormat           = fmt.Errorf("invalid field format")
	ErrSerialNumberLengthNotInRange = fmt.Errorf("serial number length not in range")
	ErrInvalidNameFormat            = fmt.Errorf("invalid name format")
	ErrExtensionNotSupported        = fmt.Errorf("extension not supported")
	ErrWhileParsingContent          = fmt.Errorf("error while parsing content")
	ErrInvalidDataType              = fmt.Errorf("invalid data type")
	ErrFileNotSupported             = fmt.Errorf("file not supported")
	ErrInvalidUrlFound              = fmt.Errorf("url not found or invalid url format")
	ErrNoLicenseKeyAttached         = fmt.Errorf("license Key not found, Signup to activate License Key")
)

var (
	ErrInvalidOperation   = fmt.Errorf("invalid operation")
	ErrRouteAlreadyExists = fmt.Errorf("route already exists")
	ErrRouteNotFound      = fmt.Errorf("route not found")
)

var (
	ErrFileNotFound    = fmt.Errorf("file not found")
	ErrTaskNotFound    = fmt.Errorf("task not found")
	ErrMachineNotfound = fmt.Errorf("machine not found")
	ErrReportNotFound  = fmt.Errorf("report not found")
)

const (
	ORGANIZATION       = "Organization"
	LICENSEKEY         = "KeysTable"
	USERAUTHENTICATION = "UserAuthentication"
	ADMIN              = "Admin"
	FEATURE            = "Feature"
	ROLE               = "Role"
	ROLEANDACTION      = "RoleAndAction"
	SCANPROFILE        = "ScanProfile"
	FILEONDEMAND       = "FileOnDemand"
	OVERRIDDENVERDICT  = "OverriddenVerdict"
	URLONDEMAND        = "UrlOnDemand"
	LOGREPORT          = "LogReport"
	DEVICE             = "Device"
	BACKUP             = "BackupTime"
	RESTORE            = "RestoreTime"
	VIEWLOG            = "ViewLog"
	PRIMARY_STRING     = "primary"
	BACKUP_STRING      = "backup"
	AUDIT_LOGS         = "audit_tables"
)

const (
	CONFIG_ADMIN              = "Admins"
	CONFIG_LICENSEKEY         = "LicenseKeys"
	CONFIG_USERAUTHENTICATION = "UserAuthentications"
	CONFIG_ORGANIZATION       = "Organizations"
	CONFIG_FEATURE            = "Features"
	CONFIG_ROLE               = "Roles"
	CONFIG_ROLEANDACTION      = "RoleAndActions"
	CONFIG_SCANPROFILE        = "ScanProfiles"
	CONFIG_FILEONDEMAND       = "FileOnDemands"
	CONFIG_OVERRIDDENVERDICT  = "OverriddenVerdicts"
	CONFIG_URLONDEMAND        = "UrlOnDemands"
	CONFIG_LOGREPORT          = "LogReports"
	CONFIG_DEVICE             = "Devices"
	CONFIG_BACKUP             = "BackupConfigTime"
	CONFIG_RESTORE            = "RestoreConfigTime"
	CONFIG_HA                 = "HA"
	CONFIG_VIEWLOG            = "ViewLog"
)

const (
	POST   = "CREATE"
	GET    = "FETCH"
	PATCH  = "UPDATE"
	DELETE = "DELETE"
)

var FileExt = map[string][]string{
	"PdfFile":            {".pdf"},
	"OfficeFile":         {".docx", ".doc", ".docm", ".dot", ".dotm", ".dotx", ".htm", ".html", ".mht", ".mhtml", ".odt", ".xls", ".xlt", ".xlsx", ".xlsm", ".xltx", ".xltm", ".xlsb", ".xla", ".xlam", ".ppt", ".pot", ".pps", ".ppa", ".pptx", ".pptm", ".potx", ".potm", ".ppam", ".ppsx", ".ppsm"},
	"ExeFile":            {".exe", ".bat", ".com", ".o", ".dll", ".gnt", ".int"},
	"FlashFile":          {".swf"},
	"WebFile":            {".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp"},
	"CompressedArchieve": {".zip", ".rar", ".tar", ".gz", ".7z"},
	"AudioFile":          {".mp3", ".wav"},
	"VideoFile":          {".mp4", ".mkv"},
	"TextFile":           {".txt", ".md", ".yaml", ".yml", ".json", ".xml", ".csv", ".log"},
	"ScriptFile":         {".sh", ".py", ".js", ".go", ".c", ".cpp", ".java", ".php", ".rb"},
}

const (
	PENDING                  = "pending"
	RUNNING                  = "running"
	COMPLETED                = "completed"
	REPORTED                 = "reported"
	ABORTED                  = "aborted"
	DONE                     = "done"
	MALWARE_FOUND_FROM_HASH  = "malware found from hash"
	CLEAN_FOUND_FROM_HASH    = "clean found from hash"
	MALWARE_FOUND_FROM_CACHE = "malware found from cache"
	CLEAN_FOUND_FROM_CACHE   = "clean found from cache"
	PREVIOUSLY_SCANNED_FILE  = "file was previously scanned or sourced from preset data "
	PREVIOUSLY_SCANNED_URL   = "url was previously scanned or sourced from preset data "
)

const (
	ALLOW = "allow"
	BLOCK = "block"
)

const (
	READONLY = 1
)

const (
	MINIMUM_ALLOWED_SERIAL_NUMBER_LENGTH = 13
	MAXIMUM_ALLOWED_SERIAL_NUMBER_LENGTH = 20
)

const (
	TIMEOUT = 300 * time.Second
)

const (
	TimeFormat   = "2006-01-02 15:04:05"
	DmiDecodeCmd = "dmidecode"
	FreeCmd      = "free"
)

var (
	LICENSEKEY_ONE_YEAR   = "YY27QPpaKipLn/isgT8MgfMDJi5MkQRe1uDV9aQp83rdCZdxvM4tXBN3wFthihP2Ml8L7vXG2xjzZZWP4rQUlU5wjboNCcbnQa3nce/5vPJ0/U7t3JukLvpmlOqLPBcUC8HUal20YsWOEArd9f4i5Q=="
	LICENSEKEY_THREE_YEAR = "djIQRArw1AF5y1O0QLmzLg476rEdM/QRJUQu7uNCrvChEnRHwA91Wh/O/Arx4Oe/6urVgIkWF0L4UsJeL4NYAKZwyjcxSlHueLEzB0mQ3/xTcLIwKHZkXjltyj4rwqdTX9K+Yuwxu3S4FRCWjtXbyg=="
	LICENSEKEY_FIVE_YEAR  = "Kap4kQyns7ilcXagNGY4DHgIuTjWSF9CG"
)

var VERSION_CHECK = "no"
var IGNORE_VULNERABLILITY = "no"
var DELETE_ORIGINAL = "no"
var DELETE_BIN_COPY = "no"
var MACHINERY = "virtualbox"
var MEMORY_DUMP = "no"
var RESCHEDULE = "no"
var TERMINATE_PROCESSES = "no"
var PROCESS_RESULTS = "yes"
var MAX_ANALYSIS_COUNT int64 = 10
var MAX_MACHINES_COUNT int64 = 8
var MAX_VM_STARTUP_COUNT int64 = 8
var FREESPACE int64 = 1024
var TMP_PATH = "/log/tmp/"
var ROOTER = "/tmp/cuckoo-rooter"
var SOURCE_IP = "192.168.56.1"
var SOURCE_PORT int64 = 2042
var MAX_ALLOWED_FILE_SIZE int64 = 104857600
var VM_STATE int64 = 60
var CRITICAL_TIMEOUT int64 = 60
var ANALYSIS_SIZE_LIMIT int64 = 134217728
var MAX_ALLOWED_FIRMWARE_FILE_SIZE int64 = 4831838208
var MAX_ALLOWED_BUILD_SIZE int64 = 1024 * 1024 * 100
var CHUNK_SIZE int64 = 1024 * 1024 * 8

const (
	FW_EMPTY   = 0
	FW_CLEAN   = 1
	FW_BLOCK   = 2
	FW_RETRY   = 3
	FW_UNKNOWN = 4
)

const (
	NOT_PRESENT = -1
	ANALYSING   = 0
	CLEAN       = 1
	MALICIOUS   = 2
	UNKNOWN     = 3
)

const (
	SERVICE_KEEPALIVED = "keepalived"
	SERVICE_BACKEND    = "backend"
	SERVICE_ZEBRA      = "zebra"
)

type ServicesActionType int

const (
	Start ServicesActionType = iota
	Stop
	Restart
)

func (a ServicesActionType) String() string {
	switch a {
	case Start:
		return "start"
	case Stop:
		return "stop"
	case Restart:
		return "restart"
	default:
		return "unknown"
	}
}

const (
	INITIAL_SETUP = "initial_setup"
	ERASE         = "erase"
)

const TIME_FORMAT = "2006-01-02 15:04:05.999999999"

const (
	TaskLiveAnalysisTable = "task_live_analysis_tables"
	TaskFinishedTable     = "task_finished_tables"
	TaskDuplicateTable    = "task_duplicate_tables"
	FileOnDemandTable     = "file_on_demands"
	UrlOnDemandTable      = "url_on_demands"
)

const (
	PendingNotInQueue = "pending not in queue"
	PendingInQueue    = "pending in queue"
	Pending           = "pending"
	RunningNotInQueue = "running not in queue"
	RunningInQueue    = "running in queue"
	Running           = "running" // in log_queue
	Reported          = "reported"
	Aborted           = "aborted"
	Completed         = "completed"
)

const HTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Text File Content</title>
</head>
<body>
    <pre>{{.Content}}</pre>
</body>
</html>
`

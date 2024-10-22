package main

import (
	"anti-apt-backend/auth"
	"anti-apt-backend/config"
	"anti-apt-backend/controller"
	"anti-apt-backend/controller/interface_handler"
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"anti-apt-backend/middlewares"
	"anti-apt-backend/service/interfaces"
	queues "anti-apt-backend/service/queue"

	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
)

var (
	webPort = ":8082"
)

func init() {
	logger.LoggingEnabled = logger.IsLoggingEnabled()
	// hash.InitHashes()
	// hash.InitUrlCache()
	initFiles()
	interfaces.InitPhysicalInterfacesConfig()
	config.DBconfig()
	dao.ResetQueueDb()
}

// main is the entry point of the application
func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Create a new default Gin router
	router := gin.Default()
	interface_handler.TempRouter = router
	// Attach CORS handling middleware to the router
	readDeviceConfig()

	ips, err := interfaces.FetchIps()
	if err != nil {
		fmt.Println("Error in fetching ips + ", err)
	}

	router.Use(middlewares.HandleCors(ips))

	fmt.Println("build updated sucessfully")
	fmt.Println("Starting main server...")

	go queues.NotInQueueHandler()
	go queues.PendingQueueHandler()
	go queues.RunningQueueHandler()
	go queues.LogQueueHandler()

	// service.CronTask()
	// service.NewWorkerPool()

	// Define routes for various HTTP methods
	// router.HandleContext()
	router.PUT("/update-logging-enabled-flag", func(ctx *gin.Context) {
		err := logger.UpdateLoggingEnabledFlag()
		if err != nil {
			ctx.JSON(400, gin.H{"error": fmt.Sprintf("Error updating logging enabled flag: %v", err)})
			return
		}
		ctx.JSON(200, gin.H{"message": "Logging enabled flag updated successfully"})
	})

	router.DELETE("/flush-sandbox-data", controller.FlushSandboxData)

	router.GET("/test", controller.Test)
	router.POST("/login", controller.Login)
	router.POST("/signup", controller.Signup)
	router.POST("/create-key", controller.CreateLicenseKey)
	router.GET("/check-license", controller.CheckLicenseKeyAvailability)

	// Create a new router group for "/wijungle" endpoints and attach JWT authentication middleware
	wijungleGroup := router.Group("/wijungle", auth.JWTAuthMiddleware(), middlewares.HandleCors(ips))
	wijungleGroup.POST("/update-personal-info", controller.UpdateAdminPersonalDetails)
	wijungleGroup.POST("/role-permission", controller.CreateRolePermission)
	wijungleGroup.POST("/scan-profile", controller.ScanProfile)
	wijungleGroup.POST("/file-on-demand", controller.CreateFileOnDemand)
	wijungleGroup.POST("/create-child-admin", controller.CreateChildAdmin)
	wijungleGroup.POST("/url-on-demand", controller.CreateUrlOnDemand)
	wijungleGroup.GET("/get-profiles", controller.GetAllProfiles)
	wijungleGroup.POST("/device", controller.CreateDevice)
	wijungleGroup.DELETE("/device", controller.DeleteDevice)
	wijungleGroup.PATCH("/device", controller.UpdateDevice)
	wijungleGroup.POST("/change-password", controller.ChangePassword)
	wijungleGroup.GET("/check-hash", controller.CheckHashOnDemand)
	wijungleGroup.POST("/extend-license", controller.ExtendLicense)

	newAuthGroup := router.Group("", auth.JWTAuthMiddleware())
	newAuthGroup.PUT("/override-verdict", controller.OverrideVerdict)
	newAuthGroup.GET("/overridden-audit-logs", controller.GetOverriddenVerdictLogs)
	router.POST("/create-job-fw", controller.CreateJobForFw)
	// router.GET("/get-job-fw", controller.GetJobForFw)
	newAuthGroup.POST("/update-build", controller.UploadBuild)
	newAuthGroup.POST("/firmware-update", controller.FirmwareUpdate)
	newAuthGroup.GET("/backup", controller.Backup)
	newAuthGroup.POST("/restore", controller.Restore)
	newAuthGroup.GET("/erase", controller.Erase)
	// router.POST("/log-report", controller.FirmwareUpdate)
	newAuthGroup.GET("wijungle/dashboard", controller.Dashboard)

	router.POST("init-device-config", controller.ApplyDeviceConfigFile)

	newAuthGroup.GET("/report", controller.GetReport)
	newAuthGroup.GET("/report/download", controller.DownloadReport)
	newAuthGroup.GET("/portmapping", interface_handler.GetPortMapping)

	newAuthGroup.POST("/troubleshoot", controller.Troubleshoot)
	router.POST("/ha", controller.CreateHa)
	newAuthGroup.GET("/ha", controller.GetHa)
	router.POST("/ha/compareDeviceInfo", controller.CompareDeviceInfoFromAnotherAppliance)
	router.POST("/ha/copy-config", controller.HaCopyConfig)
	router.POST("/ha/sync-backup", controller.SyncBackup)
	router.POST("/ha/generate-keepalived-config-for-backup", controller.GenerateKeepalivedConfigForBackup)
	router.POST("/ha/disable", controller.DisableHaInAnotherAppliance)
	router.POST("/ha/update-last-synced-at", controller.UpdateLastSyncedAt)
	// physical interface API's
	newAuthGroup.GET("/physical_link", interface_handler.ListPhysicalInterfacesHandler)
	newAuthGroup.GET("/physical_link/:physical_interface_name", interface_handler.ListPhysicalInterfacesHandler)
	newAuthGroup.PUT("/physical_link/:physical_interface_name", interface_handler.UpdatePhysicalInterfaceHandler)

	// vlan API's
	newAuthGroup.POST("/vlan", interface_handler.CreateVlanInterfaceHandler)
	newAuthGroup.GET("/vlan", interface_handler.ListVlanInterfacesHandler)
	newAuthGroup.GET("/vlan/:vlan_interface_name", interface_handler.ListVlanInterfacesHandler)
	newAuthGroup.PUT("/vlan/:vlan_interface_name", interface_handler.UpdateVlanInterfaceHandler)
	newAuthGroup.DELETE("/vlan/:vlan_interface_name", interface_handler.DeleteVlanInterfaceHandler)

	// bond API's
	newAuthGroup.POST("/bond", interface_handler.CreateBondedLink)
	newAuthGroup.GET("/bond", interface_handler.ListBondInterfaces)
	newAuthGroup.GET("/bond/:bond_interface_name", interface_handler.ListBondInterfaces)
	newAuthGroup.PUT("/bond/:bond_interface_name", interface_handler.UpdateBondInterface)
	newAuthGroup.DELETE("/bond/:bond_interface_name", interface_handler.DeleteBondInterface)

	// bridge API's
	newAuthGroup.POST("/bridge", interface_handler.CreateBridgeInterfaceHandler)
	newAuthGroup.GET("/bridge/:bridge_interface_name", interface_handler.ListBridgeInterfacesHandler)
	newAuthGroup.GET("/bridge", interface_handler.ListBridgeInterfacesHandler)
	newAuthGroup.PUT("/bridge/:bridge_interface_name", interface_handler.UpdateBridgeInterfaceHandler)
	newAuthGroup.DELETE("/bridge/:bridge_interface_name", interface_handler.DeleteBridgeInterfaceHandler)

	// static routing API's
	newAuthGroup.POST("/routing/static_routing", interface_handler.CreateStaticRouteHandler)
	newAuthGroup.GET("/routing/static_routing", interface_handler.ListStaticRoutesHandler)
	newAuthGroup.DELETE("/routing/static_routing/:operation", interface_handler.DeleteStaticRouteHandler)
	newAuthGroup.PUT("/routing/static_routing/:operation", interface_handler.UpdateStaticRouteHandler)

	router.POST("/api/v1/checkhash", controller.TestAPTFw)
	// for it team
	if _, err := os.Stat(extras.TEMP_BUILD_PATH + "device_config"); err == nil {
		router.LoadHTMLGlob("/var/www/html/web/device_config/*")
		router.GET("/vHausTY1&skiuLnBGySnS10nsoOnsd", controller.DeviceConfigUiForm)
	} else {
		fmt.Println("Device config UI not found")
	}

	// Load TLS certificate and private key
	// certFile := "/etc/ssl-certs/cert.pem"
	// keyFile := "/etc/ssl-certs/cert.key"
	// router.RunTLS(webPort, certFile, keyFile)

	// Start the HTTP server and listen on the specified port
	router.Run(webPort)

	// This will update the CORS trusted origins after every api call
	// ips, err = interfaces.FetchIps()
	// if err != nil {
	// 	fmt.Println("Error in fetching ips + ", err)
	// }
	// router.Use(middlewares.HandleCors(ips))
}

// HandleCors returns a Gin middleware for handling CORS.

func readDeviceConfig() {
	file, err := os.Open(extras.ROOT_DATA_DEVICE_CONFIG)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			if strings.Contains(line, "license_1year") {
				extras.LICENSEKEY_ONE_YEAR = parts[1]
			}
			if strings.Contains(line, "license_3year") {
				extras.LICENSEKEY_THREE_YEAR = parts[1]
			}
			if strings.Contains(line, "license_5year") {
				extras.LICENSEKEY_FIVE_YEAR = parts[1]
			}
		}
	}
}

func initFiles() {
	// Create directories if they don't exist

	var directories = []string{
		extras.ROOT_DATA_DEVICE_CONFIG,
		extras.DATA_PATH,
		extras.TASK_LOGS_PATH,
		extras.REPORT_DOWNLOADS_PATH,
		extras.SANDBOX_FILE_PATHS,
	}

	for _, dir := range directories {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0777)
			if err != nil {
				log.Println(err)
			}
		}
	}

	var paths = []string{
		extras.ROOT_DATA_DEVICE_CONFIG,
		extras.HA_STATE_FILE,
		extras.PLATFORM_FILE_NAME,
		extras.DEVICE_REBOOTED_FLAG_PATH,
		extras.LOGGING_ENABLED_FLAG_PATH,
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0777)
			if err != nil {
				log.Println(err)
			}
			file.Close()
		}
	}

}

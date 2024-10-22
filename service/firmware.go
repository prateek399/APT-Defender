package service

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Firmware struct {
}

const (
	installStatusFilePath = "/var/www/html/data/firmware_install_status"
)

var (
	RAUC         = "rauc"
	BUSCTL       = "busctl"
	GRUB_INSTALL = "grub-install"
	UPDATE_GRUB  = "update-grub"
)

type BootLoader struct {
	BEGIN_STR string
}

func NewBootLoader() *BootLoader {
	return &BootLoader{
		BEGIN_STR: "default=0\ntimeout=3\n\nset ORDER=\"A B\"\nset A_OK=0\nset B_OK=0\nset A_TRY=0\nset B_TRY=0\nload_env\n\n# select bootable slot\nfor SLOT in $ORDER; do\n    if [ \"$SLOT\" == \"A\" ]; then\n        INDEX=1\n        OK=$A_OK\n        TRY=$A_TRY\n        A_TRY=1\n    fi\n    if [ \"$SLOT\" == \"B\" ]; then\n        INDEX=2\n        OK=$B_OK\n        TRY=$B_TRY\n        B_TRY=1\n    fi\n    if [ \"$OK\" -eq 1 -a \"$TRY\" -eq 0 ]; then\n        default=$INDEX\n        break\n    fi\ndone\n\n# reset booted flags\nif [ \"$default\" -eq 0 ]; then\n    if [ \"$A_OK\" -eq 1 -a \"$A_TRY\" -eq 1 ]; then\n        A_TRY=0\n    fi\n    if [ \"$B_OK\" -eq 1 -a \"$B_TRY\" -eq 1 ]; then\n        B_TRY=0\n    fi\nfi\n\nsave_env ORDER\n\n",
	}
}

func (bl *BootLoader) getLatestKernel() (string, string) {
	files, _ := filepath.Glob("/boot/*")
	vmLinuzFiles := filterAndSortFiles(files, "vmlinuz-")
	initrdFiles := filterAndSortFiles(files, "initrd.img")

	return vmLinuzFiles[len(vmLinuzFiles)-1], initrdFiles[len(initrdFiles)-1]
}

func filterAndSortFiles(files []string, prefix string) []string {
	var filteredFiles []string
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), prefix) {
			filteredFiles = append(filteredFiles, file)
		}
	}
	sort.Strings(filteredFiles)
	return filteredFiles
}

func (bl *BootLoader) createGrubConfig() {
	vmLinuzFile, initrdFile := bl.getLatestKernel()
	op := ""
	if _, err := os.Stat("/boot/grub/extra_options"); err == nil {
		content, _ := os.ReadFile("/boot/grub/extra_options")
		op = string(content)
	}
	str := bl.BEGIN_STR
	str += `CMDLINE="console=ttyS0,115200 console=tty12 net.ifnames=0 biosdevname=0 panic=60 quiet  reboot=bios"`

	str += "\n\n\nif [ \"$ORDER\" == \"A B\" ]; then\n    menuentry \"WiJungle OS Slot A\" {\n        linux (hd0,gpt1)/" + vmLinuzFile + " root=/dev/sda3 console=ttyS0,115200 console=tty12 net.ifnames=0 biosdevname=0 panic=60 quiet " + op + " rauc.slot=A\n        initrd (hd0,gpt1)/" + initrdFile + "\n    }\nfi\n\nif [ \"$ORDER\" == \"B A\" ]; then\n    menuentry \"WiJungle OS Slot B\" {\n        linux (hd0,gpt1)/" + vmLinuzFile + " root=/dev/sda8 console=ttyS0,115200 console=tty12 net.ifnames=0 biosdevname=0 panic=60 quiet " + op + " rauc.slot=B\n        initrd (hd0,gpt1)/" + initrdFile + "\n    }\nfi"

	log.Println("Str in createGrubConfig: ", str)

	os.WriteFile("/var/www/html/data/grub.cfg", []byte(str), 0644)
	execCmd("sudo /bin/cp -f /var/www/html/data/grub.cfg /boot/grub/grub.cfg")
}

func execCmd(command string) {
	log.Println("Command: ", command)
	cmd := exec.Command(command)
	err := cmd.Run()
	if err != nil {
		log.Println("Error: ", err)
	}
}

func StartBoot() {
	firmware := Firmware{}
	result := firmware.install(0, "rauc")
	log.Println("Firmware installation : ", result)
}

func getDeviceInfo() map[string]string {
	serial := ""
	file, err := os.Open(extras.ROOT_DATA_DEVICE_CONFIG)
	if err != nil {
		log.Println("Error: ", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			if strings.Contains(line, "serial_no") {
				serial = parts[1]
			}
		}
	}
	return map[string]string{"device_serial_id": serial, "softwareversion": "1.0.0"}
}

// pick from /data/device_config & /etc/software

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func filePutContents(path, data string) {
	err := os.WriteFile(path, []byte(data), 0644)
	if err != nil {
		log.Println("Error writing to file: ", err)
	}
}

func fileGetContents(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Println("Error reading file: ", err)
	}
	return string(content)
}

func calculateDirectorySize(dirPath string) (int64, error) {
	var totalSize int64
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() || info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	return totalSize, err
}

var dirPath = "/log"

func (f *Firmware) initSpace() string {
	totalSize, err := calculateDirectorySize(dirPath)
	if err != nil {
		log.Println("error while calculating TotalSize in initSpace: ", err)
	}
	log.Println("TotalSize in initSpace: ", totalSize)

	ExecCmd(RAUC + "status mark-good")
	oneMB := 1024 * 1024
	oneGB := 1024 * oneMB
	device := getDeviceInfo()
	if !fileExists("/etc/device_serial_id") {
		filePutContents("/etc/device_serial_id", device["device_serial_id"])
	}
	if !fileExists("/etc/firmware_path") && fileExists("/dev/sda9") {
		filePutContents("/etc/firmware_path", extras.FIRMWARE_FILE_PATH)
	}
	// if !fileExists("/etc/software_version") {
	// 	filePutContents("/etc/software_version", device["softwareversion"])
	// }
	filePutContents(installStatusFilePath, "Idle")
	if totalSize > int64(4*oneGB) {
		if !fileExists("/var/log/firmware") {
			err = os.MkdirAll("/var/log/firmware", 0777)
			if err != nil {
				log.Println("Error while creating directory in initspace: ", err)
			}
		}
		filePutContents("/etc/firmware_path", extras.FIRMWARE_FILE_PATH)
		return extras.FIRMWARE_FILE_PATH
	}
	if !fileExists("/var/www/firmware") {
		err = os.MkdirAll("/var/www/firmware", 0777)
		if err != nil {
			log.Println("Error while creating directory in initspace: ", err)
		}
	}
	filePutContents("/etc/firmware_path", "/var/www/firmware/")
	return "/var/www/firmware/"
}

func (f *Firmware) compatibilityCheck() {

}

func (f *Firmware) errorSanitization(str string) string {
	fmt.Println("Error: ", str)
	if strings.Contains("Did you pass a valid RAUC bundle", str) {
		return "Invalid Firmware Image Uploaded"
	}
	if strings.Contains("Installing", str) && strings.Contains("failed", str) {
		return "Installing Failed"
	}
	return str
}

func (f *Firmware) currentStatus() int {

	log.Println("Checking current status of firmware upgrade")

	cmd := BUSCTL + " get-property de.pengutronix.rauc / de.pengutronix.rauc.Installer Operation"
	res := ExecCmd(cmd)
	if strings.Contains("idle", res) {
		return -1
	}
	if strings.Contains("installing", res) {
		return -2
	}
	return 0
}

func (f *Firmware) install(callType int, method string) map[string]interface{} {
	log.Println("Initializing Firmware upgrade")
	firmwarePath := ""
	if fileExists("/etc/firmware_path") {
		firmwarePath = fileGetContents("/etc/firmware_path")
	} else {
		firmwarePath = f.initSpace()
	}
	// NUM_SLOTS := 2
	// firmwarePath += "fw.raucb"

	files, err := os.ReadDir(firmwarePath)
	if err != nil {
		log.Println("Error reading directory: ", err)
	}

	if len(files) != 1 {
		return map[string]interface{}{"status": "fail", "errors": "Firmware Image Not Found or Multiple Firmware Images Found"}
	}

	firmwarePath += files[0].Name()

	if !fileExists(firmwarePath) {
		log.Println(fmt.Sprintf("firmwarePath: %s", firmwarePath))
		return map[string]interface{}{"status": "fail", "errors": "Firmware Image Not Found"}
	}
	if method == "rauc" {
		res := ExecCmd(RAUC + " install " + firmwarePath)
		if strings.Contains("failed", res) {
			return map[string]interface{}{"status": "fail", "errors": f.errorSanitization(res)}
		}
		if strings.Contains("Installing", res) && strings.Contains("succeeded", res) {
			res = "Installation done. Rebooting in 5 sec"
		}
		ExecCmd("sudo /bin/cp -f /etc/rauc/grub.conf /boot/grub/grub.conf")
		ExecCmd(RAUC + " status mark-active other")
		ExecCmd(GRUB_INSTALL + " /dev/sda")
		ExecCmd(UPDATE_GRUB)
		filePutContents(installStatusFilePath, "Installed "+time.Now().Format("2006-01-02 15:04:05"))
		bootLoader := NewBootLoader()
		bootLoader.createGrubConfig()

		logger.LoggerFunc("error", logger.LoggerMessage("sysLog:Device is rebooting after firmware upgrade"))
		err := reboot()
		if err != nil {
			logger.LoggerFunc("error", logger.LoggerMessage("sysLog:Error while rebooting device after firmware upgrade"+err.Error()))
		}
		return map[string]interface{}{"status": "success", "message": res}
	}
	status := f.currentStatus()
	res := ""
	if callType == 0 {
		filePutContents(installStatusFilePath, "Idle")
	}
	if status == -1 {
		cs := fileGetContents(installStatusFilePath)
		if cs == "Idle" {
			filePutContents(installStatusFilePath, "Installing "+time.Now().Format("2006-01-02 15:04:05"))
			res = ExecCmd(BUSCTL + " call de.pengutronix.rauc / de.pengutronix.rauc.Installer InstallBundle sa{sv} " + firmwarePath + " 0")
			res = "Installing"
		} else {
			res = ExecCmd(BUSCTL + " get-property de.pengutronix.rauc / de.pengutronix.rauc.Installer LastError")
			if res == "s \"\"" {
				ExecCmd("sudo /bin/cp -f /etc/rauc/grub.conf /boot/grub/grub.conf")
				ExecCmd("rauc status mark-active other")
				res = "Installation done. Rebooting in 5 sec"
				filePutContents(installStatusFilePath, "Installed "+time.Now().Format("2006-01-02 15:04:05"))
				logger.LoggerFunc("error", logger.LoggerMessage("sysLog:Device is rebooting after firmware upgrade"))
				err := reboot()
				if err != nil {
					logger.LoggerFunc("error", logger.LoggerMessage("sysLog:Error while rebooting device after firmware upgrade"+err.Error()))
				}
			} else {
				return map[string]interface{}{"status": "fail", "errors": f.errorSanitization(res)}
			}
		}
	} else if status == -2 {
		res = ExecCmd(BUSCTL + " call de.pengutronix.rauc / de.pengutronix.rauc.Installer InstallBundle sa{sv} " + firmwarePath + " 0")
		if strings.Contains("Already processing a different method", res) {
			res = "Still Installing "
		} else if strings.Contains("Copying", res) {
			res = "Installing " + strings.Split(res, " ")[1] + "%"
		} else if strings.Contains("Installing done", res) {
			ExecCmd("sudo /bin/cp -f /etc/rauc/grub.conf /boot/grub/grub.conf")
			ExecCmd("rauc status mark-active other")
			res = "Installation done. Rebooting in 5 sec"
			filePutContents(installStatusFilePath, "Installed "+time.Now().Format("2006-01-02 15:04:05"))
			logger.LoggerFunc("error", logger.LoggerMessage("sysLog:Device is rebooting after firmware upgrade"))
			err := reboot()
			if err != nil {
				logger.LoggerFunc("error", logger.LoggerMessage("sysLog:Error while rebooting device after firmware upgrade"+err.Error()))
			}
		} else if strings.Contains("failed", res) {
			filePutContents(installStatusFilePath, "Error: "+res+" "+time.Now().Format("2006-01-02 15:04:05"))
			res = "Installation Failed. Please try again"
			return map[string]interface{}{"status": "fail", "errors": f.errorSanitization(res)}
		} else {
			res = ExecCmd(BUSCTL + " get-property de.pengutronix.rauc / de.pengutronix.rauc.Installer LastError")
			return map[string]interface{}{"status": "fail", "errors": f.errorSanitization(res)}
		}
	}
	return map[string]interface{}{"status": "success", "message": res}
}

func ExecCmd(command string) string {
	log.Println("Command: ", command)
	cmd := exec.Command("bash", "-c", command)
	err := cmd.Run()
	if err != nil {
		log.Println("Error: ", err)
	}
	cmdOutput, err := cmd.Output()
	if err != nil {
		log.Println("Error: ", err)
	}
	return string(cmdOutput)
}

func reboot() error {
	cmd := exec.Command("reboot")
	err := cmd.Run()
	return err
}

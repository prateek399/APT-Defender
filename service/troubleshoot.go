package service

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"anti-apt-backend/model"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var allowedCommands = map[string]bool{
	"ping":       true,
	"nslookup":   true,
	"traceroute": true,
	"nmap":       true,
	"systemctl":  true,
	"config":     true,
	"ip":         true,
}

func isCommandAllowed(command string) bool {
	parts := strings.Split(command, " ")
	if len(parts) == 0 {
		return false
	}
	return allowedCommands[parts[0]]
}

func Troubleshoot(req model.TroubleshootRequest) model.APIResponse {
	var resp model.APIResponse

	var cmdResp string

	switch req.Type {
	case 1:
		err, res := ping(strings.TrimSpace(req.Name))
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, err.Error(), extras.ErrInvalidFieldFormat)
			return resp
		}
		cmdResp = res
	case 2:
		err, res := nslookup(strings.TrimSpace(req.Name))
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, err.Error(), extras.ErrInvalidFieldFormat)
			return resp
		}
		cmdResp = res
	case 3:
		err, res := traceroute(strings.TrimSpace(req.Name))
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, err.Error(), extras.ErrInvalidFieldFormat)
			return resp
		}
		cmdResp = res
	case 4:
		err, res := portScan(strings.TrimSpace(req.IpAddress), strings.TrimSpace(req.Port))
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, err.Error(), extras.ErrInvalidFieldFormat)
			return resp
		}
		cmdResp = res

		if strings.Contains("unfiltered", cmdResp) {
			cmdResp = "The port is unfiltered"
		} else if strings.Contains("filtered", cmdResp) {
			cmdResp = "The port is filtered"
		} else if strings.Contains("closed", cmdResp) || strings.Contains("Host seems down", cmdResp) {
			cmdResp = "The port is closed"
		} else if strings.Contains("open", cmdResp) {
			cmdResp = "The port is open"
		}
	case 5:
		command := strings.ToLower(strings.TrimSpace(req.Command))
		parts := strings.Fields(command)
		if len(parts) == 0 || !isCommandAllowed(command) {
			resp = model.NewErrorResponse(http.StatusBadRequest, "invalid command", extras.ErrInvalidCommandInTroubleshoot)
			return resp
		}

		if parts[0] == "ping" {
			parts[0] = "/bin/ping"
			if len(parts) == 2 {
				parts = append(parts, parts[1])
				parts[1] = "-c 4"
			}
		} else if parts[0] == "nslookup" {
			parts[0] = "/usr/bin/nslookup"
		} else if parts[0] == "traceroute" {
			parts[0] = "/usr/sbin/traceroute"
		} else if parts[0] == "nmap" {
			parts[0] = "/usr/bin/nmap"
		} else if parts[0] == "systemctl" {
			parts[0] = "/bin/systemctl"
			if !validateSystemctlCommand(parts[1]) {
				resp = model.NewErrorResponse(http.StatusBadRequest, "invalid command", extras.ErrInvalidCommandInTroubleshoot)
				return resp
			}
		} else if parts[0] == "config" {
			if len(parts) == 3 && parts[1] == "show" && parts[2] == "supported-os" {
				return model.NewSuccessResponse(extras.ERR_SUCCESS, "Supported OS: Windows XP, Windows 7(default), Ubuntu 20")
			} else if len(parts) == 3 && parts[1] == "show" && parts[2] == "apt-sysinfo" {
				resp := " Files Per Hour - 12,100\n IMIX Files Per Hour - 3,000\n Virtual Machines - 8 \n Supported OS: Windows XP, Windows 7(default), Ubuntu 20 \n File size limit - 100 MB \n Supported File Types : 70+ (EXE, DLL, Archives (ISO, ZIP,7Z, RAR, etc.), PDF, scripts & more)"
				return model.NewSuccessResponse(extras.ERR_SUCCESS, resp)
			} else if len(parts) == 4 && parts[1] == "os" {
				// Expected format: 'config os {vm_no} {os_name}
				err := AlterVmOs(parts[2], parts[3])
				if err != nil {
					return model.NewErrorResponse(http.StatusBadRequest, err.Error(), extras.ErrInvalidCommandInTroubleshoot)
				}

				logger.LoggerFunc("info", logger.LoggerMessage("vmLog:OS updated successfully for VM: "+parts[2]))
				return model.NewSuccessResponse(extras.ERR_SUCCESS, "OS updated successfully for VM: "+parts[2])
			}
			return model.NewErrorResponse(http.StatusBadRequest, "invalid command", extras.ErrInvalidCommandInTroubleshoot)
		} else if parts[0] == "ip" {
			parts[0] = "/sbin/ip"
		} else {
			resp = model.NewErrorResponse(http.StatusBadRequest, "invalid command", extras.ErrInvalidCommandInTroubleshoot)
			return resp
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		cmdResp = processCommand(cmd)
	default:
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_TYPE_IN_TROUBLESHOOT, extras.ErrInvalidTypeInTroubleshoot)
		return resp
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, cmdResp)
	return resp
}

func processCommand(cmd *exec.Cmd) string {
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "Error while executing command: " + err.Error()
	}
	return string(output)
}

func validateDomainOrIPAddress(value string) bool {
	ipRegex := `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	domainRegex := `^([a-zA-Z0-9-]+\.){1,}([a-zA-Z]{2,})$`
	return regexp.MustCompile(ipRegex).MatchString(value) || regexp.MustCompile(domainRegex).MatchString(value)
}

func validateIPAddress(value string) bool {
	ipRegex := `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	return regexp.MustCompile(ipRegex).MatchString(value)
}

func validatePort(value string) bool {
	portRegex := `^([1-9]|[1-5][0-9]{0,4}|[6-9][0-9]{1,3}|[1-6][0-5][0-5][0-3][0-5])$`
	return regexp.MustCompile(portRegex).MatchString(value)
}

func validateDomain(value string) bool {
	domainRegex := `^([a-zA-Z0-9-]+\.){1,}([a-zA-Z]{2,})$`
	return regexp.MustCompile(domainRegex).MatchString(value)
}

func validateSystemctlCommand(value string) bool {
	return value == "status" || value == "stop" || value == "restart" || value == "is-active" || value == "is-enabled" || value == "is-failed" || value == "list-units" || value == "list-unit-files"
}

func ping(host string) (error, string) {
	if host == extras.EMPTY_STRING {
		return fmt.Errorf("please add the required fields"), extras.EMPTY_STRING
	}

	if !validateDomainOrIPAddress(host) {
		return fmt.Errorf("Invalid domain or IP address"), extras.EMPTY_STRING
	}

	cmd := exec.Command("ping", "-c", "4", host)
	return nil, processCommand(cmd)
}

func nslookup(host string) (error, string) {

	if host == extras.EMPTY_STRING {
		return fmt.Errorf("please add the required fields"), extras.EMPTY_STRING
	}

	if !validateDomain(host) {
		return fmt.Errorf("Invalid domain name"), extras.EMPTY_STRING
	}

	cmd := exec.Command("nslookup", host)
	return nil, processCommand(cmd)
}

func traceroute(host string) (error, string) {

	if host == extras.EMPTY_STRING {
		return fmt.Errorf("please add the required fields"), extras.EMPTY_STRING
	}

	if !validateDomainOrIPAddress(host) {
		return fmt.Errorf("Invalid domain or IP address"), extras.EMPTY_STRING
	}

	cmd := exec.Command("traceroute", host)
	return nil, processCommand(cmd)
}

func portScan(ipAddress, port string) (error, string) {

	if ipAddress == extras.EMPTY_STRING || port == extras.EMPTY_STRING {
		return fmt.Errorf("IP Address and Port are required"), extras.EMPTY_STRING
	}

	if !validateIPAddress(ipAddress) || !validatePort(port) {
		return fmt.Errorf("Invalid IP address or port"), extras.EMPTY_STRING

	}

	cmd := exec.Command("nmap", ipAddress, "-p", port)
	return nil, processCommand(cmd)
}

func AlterVmOs(vmNumberString string, osName string) error {

	vmNumber, err := strconv.Atoi(vmNumberString)
	if err != nil {
		return fmt.Errorf("invalid vm number")
	}

	if vmNumber < 1 || vmNumber > 8 {
		return fmt.Errorf("invalid vm number")
	}

	switch strings.ToLower(strings.TrimSpace(osName)) {
	case "windows_7", "windows-7", "windows7":
		osName = "Windows 7"
	case "ubuntu_20", "ubuntu-20", "ubuntu20":
		osName = "Ubuntu 20"
	default:
		return fmt.Errorf("invalid operating system. Supported operating systems are 'Windows 7' or 'Ubuntu 20'")
	}

	err = updateOSConfig(vmNumber, osName)
	if err != nil {
		return err
	}

	return nil
}

func updateOSConfig(vmNumber int, osName string) error {
	data, err := os.ReadFile(extras.PLATFORM_FILE_NAME)
	if os.IsNotExist(err) {
		data = []byte{}
	} else if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	configMap := make(map[int]string)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		vmNo, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		configMap[vmNo] = parts[1]
	}

	configMap[vmNumber] = osName

	var updatedConfig strings.Builder
	for vmNo, osName := range configMap {
		updatedConfig.WriteString(fmt.Sprintf("%d %s\n", vmNo, osName))
	}
	return os.WriteFile(extras.PLATFORM_FILE_NAME, []byte(updatedConfig.String()), 0644)
}

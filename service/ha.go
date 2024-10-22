package service

import (
	"anti-apt-backend/config/interface_config"
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"anti-apt-backend/model"
	"anti-apt-backend/model/interface_model"
	"anti-apt-backend/service/interfaces"
	"anti-apt-backend/util/interface_utils"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
	"gopkg.in/yaml.v2"
)

func CreateHa(request interface_model.CreateHaRequest) model.APIResponse {
	var resp model.APIResponse

	req := interface_utils.TrimStringsInStruct(request).(interface_model.CreateHaRequest)

	if err := validateHaRequest(req); err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		return resp
	}

	var respMsg string

	if req.RequestType == 1 {

		if strings.ToLower(req.ApplianceRole) == extras.BACKUP_STRING {
			req.MonitoredInterfaces = []interface_model.MonitoredInterface{}
			req.KeepAliveRequestInterval = 0
			req.KeepAliveAttempts = 0
		}

		config, err := interface_config.FetchConfig(interface_model.HA_STRING)
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

		haConfig := config.Ha

		var monitoredIntfs []interface_model.MonitoredInterface

		for _, intf := range req.MonitoredInterfaces {
			intfResp, err := interfaces.ListPhysicalInterfaces(intf.Interface, "HA")
			if err != nil {
				resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
				return resp
			}

			monitoredIntfs = append(monitoredIntfs, interface_model.MonitoredInterface{
				Interface:   intf.Interface,
				BaseIp:      intf.BaseIp,
				PeerIp:      intf.PeerIp,
				InterfaceIp: intfResp.PhysicalInterfaces[0].IpAddress.IpAddress + "/" + fmt.Sprint(intfResp.PhysicalInterfaces[0].IpAddress.Netmask),
			})
		}

		storeData := interface_model.Ha{
			ApplianceMode:            strings.ToLower(req.ApplianceMode),
			ApplianceRole:            strings.ToLower(req.ApplianceRole),
			Password:                 req.Password,
			DedicatedHaInterface:     req.DedicatedHaInterface,
			PeerIp:                   req.PeerIp,
			MonitoredInterfaces:      monitoredIntfs,
			KeepAliveRequestInterval: req.KeepAliveRequestInterval,
			KeepAliveAttempts:        req.KeepAliveAttempts,
			HaStatus: interface_model.HaStatus{
				LastSyncedAt: haConfig.HaStatus.LastSyncedAt,
				Status:       haConfig.HaStatus.Status,
			},
		}

		err = interface_config.UpdateConfig(interface_model.Config{Ha: storeData}, interface_model.HA_STRING, "", "CREATE HA")
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
			return resp
		}

		// for _, intf := range req.MonitoredInterfaces {
		// 	intfResp, err := interfaces.ListPhysicalInterfaces(intf.Interface)
		// 	if err != nil {
		// 		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		// 		return resp
		// 	}

		// 	interface_config.UpdateConfig(interface_model.Config{
		// 		PhysicalInterfaces: intfResp.PhysicalInterfaces,
		// 	}, "device", intf.Interface, "HA")
		// }

		err = createHaConfig(true, true)
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, "Error while creating HA Config", err)
			return resp
		}
		respMsg = "Successfully created HA Configuration"
	} else if req.RequestType == 2 {

		config, err := interface_config.FetchConfig(interface_model.HA_STRING)
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

		haConfig := config.Ha

		if haConfig.ApplianceRole == extras.EMPTY_STRING {
			resp = model.NewErrorResponse(http.StatusBadRequest, "HA is not enabled", extras.ErrInvalidActionType)
			return resp
		}

		// err = DisableHaInAnotherAppliance(haConfig.PeerIp)
		// if err != nil {
		// 	resp = model.NewErrorResponse(http.StatusBadRequest, "Error while disabling HA", err)
		// 	return resp
		// }
		resp := DisableHa()
		if resp.StatusCode != http.StatusOK {
			return resp
		}
		respMsg = "Successfully disabled HA"
	} else if req.RequestType == 3 {

		err := IntializeSyncBackup()
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, "Error while syncing HA backup", err)
			return resp
		}

		config, err := interface_config.FetchConfig(interface_model.HA_STRING)
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
			return resp
		}

		haConfig := config.Ha
		nowTime := time.Now().Format(extras.TIME_FORMAT)

		haConfig.HaStatus.LastSyncedAt = nowTime
		err = interface_config.UpdateConfig(interface_model.Config{Ha: haConfig}, interface_model.HA_STRING, "", "SYNC HA")
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
			return resp
		}

		err = sendLastSyncedAtToPeer(haConfig.PeerIp, nowTime)
		if err != nil {
			resp = model.NewErrorResponse(http.StatusBadRequest, "Error in sending last synced at to HA peer", err)
			return resp
		}

		go func() {
			time.Sleep(3 * time.Second)
			err = SetIpAfterSyncBackup()
			if err != nil {
				logger.LoggerFunc("error", logger.LoggerMessage("Error setting IP addresses in primary appliance after syncing with backup appliance"))
			}
			err = createHaConfig(true, true)
			if err != nil {
				logger.LoggerFunc("error", logger.LoggerMessage("Error while creating HA Config"))
			}
			// restartServices([]string{"keepalived", "backend"}, 5)

			ServiceActions([]string{extras.SERVICE_KEEPALIVED, extras.SERVICE_BACKEND}, extras.Restart, 5)

		}()

		respMsg = "Successfully Synced configuration with backup device"
	} else {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, fmt.Errorf("Invalid request type"))
		return resp

	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, respMsg)
	return resp
}

func validateHaRequest(req interface_model.CreateHaRequest) error {

	if req.RequestType != 1 && req.RequestType != 2 && req.RequestType != 3 {
		return fmt.Errorf("Invalid request type")
	}

	if req.RequestType == 1 {
		mode := strings.ToLower(strings.TrimSpace(req.ApplianceMode))

		if mode != "active-backup" && mode != "active-active" {
			return fmt.Errorf("Invalid appliance mode")
		}

		role := strings.ToLower(strings.TrimSpace(req.ApplianceRole))

		if mode == "active-active" && role == extras.BACKUP_STRING {
			return fmt.Errorf("Invalid appliance role")
		}

		if role != extras.PRIMARY_STRING && role != extras.BACKUP_STRING {
			return fmt.Errorf("Invalid appliance role")
		}

		dedicatedHaIntf := strings.TrimSpace(req.DedicatedHaInterface)

		if dedicatedHaIntf == extras.EMPTY_STRING {
			return fmt.Errorf("Invalid dedicated HA interface")
		}

		err := interface_utils.CheckInterfaceExists(dedicatedHaIntf)
		if err != nil {
			return fmt.Errorf("Invalid dedicated HA interface")
		}

		peerIp := strings.TrimSpace(req.PeerIp)

		if peerIp == extras.EMPTY_STRING {
			return fmt.Errorf("Invalid peer IP")
		}

		validIp, _ := interface_utils.ValidateIP(peerIp)

		if !validIp {
			return fmt.Errorf("Invalid peer IP")
		}

		monitoredIntfs := req.MonitoredInterfaces

		for _, intf := range monitoredIntfs {

			intfName := strings.TrimSpace(intf.Interface)

			if intfName == extras.EMPTY_STRING {
				return fmt.Errorf("Invalid monitored interface")
			}

			err := interface_utils.CheckInterfaceExists(intfName)
			if err != nil {
				return fmt.Errorf("Invalid monitored interface")
			}

			baseIp := strings.TrimSpace(intf.BaseIp)
			peerIp := strings.TrimSpace(intf.PeerIp)
			if baseIp != extras.EMPTY_STRING {
				validIp, _ := interface_utils.ValidateIP(baseIp)
				if !validIp {
					return fmt.Errorf("Invalid base IP for monitored interface : " + interfaces.PortMapping[intf.Interface])
				}

				if baseIp == peerIp {
					return fmt.Errorf("Base IP and Peer IP cannot be same for monitored interface : " + interfaces.PortMapping[intf.Interface])
				}

				if interface_utils.CheckIfIpAlreadyExists(baseIp) {
					return fmt.Errorf("Base IP already exists on another port")
				}
			}

			if peerIp != extras.EMPTY_STRING {
				validIp, _ := interface_utils.ValidateIP(peerIp)
				if !validIp {
					return fmt.Errorf("Invalid peer IP for monitored interface : " + interfaces.PortMapping[intf.Interface])
				}
			}
		}

	} else if req.RequestType == 3 {
		config, err := interface_config.FetchConfig(interface_model.HA_STRING)
		if err != nil {
			return fmt.Errorf("Error in fetching HA Configuration")
		}

		haConfig := config.Ha

		if haConfig.ApplianceRole == extras.EMPTY_STRING {
			return fmt.Errorf("HA is not enabled")
		}

		if haConfig.ApplianceRole != extras.PRIMARY_STRING {
			return fmt.Errorf("Sync backup can only be initiated from primary appliance")
		}
	} else if req.RequestType == 2 {
		config, err := interface_config.FetchConfig(interface_model.HA_STRING)
		if err != nil {
			return fmt.Errorf("Error in fetching HA Configuration")
		}

		haConfig := config.Ha

		if haConfig.ApplianceRole == extras.EMPTY_STRING {
			return fmt.Errorf("HA is not enabled")
		}

	}

	return nil
}

func DisableHa() model.APIResponse {
	var resp model.APIResponse

	config, err := interface_config.FetchConfig(interface_model.HA_STRING)
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	haConfig := config.Ha

	if haConfig.ApplianceRole == extras.EMPTY_STRING {
		resp = model.NewErrorResponse(http.StatusBadRequest, "HA is already in disabled state", extras.ErrInvalidActionType)
		return resp
	}

	err = interface_config.UpdateConfig(interface_model.Config{Ha: interface_model.Ha{}}, interface_model.HA_STRING, "", "DELETE HA")
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_SAVING_DATA, err)
		return resp
	}

	err = resetKeepAlived()
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, "Error while disabling HA", err)
		return resp
	}

	logger.LoggerFunc("info", logger.LoggerMessage("successfully disabled HA, system will reboot in 5 seconds"))

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "successfully disabled HA, now the system will reboot in 5 seconds")

	go func() {
		time.Sleep(7 * time.Second)
		cmd := exec.Command("reboot")
		err = cmd.Run()
		if err != nil {
			logger.LoggerFunc("error", logger.LoggerMessage("sysLog: Error in rebooting system"+err.Error()))
		}
	}()

	return resp

}

func DisableHaInAnotherAppliance(peerIp string) error {
	if peerIp == extras.EMPTY_STRING {
		log.Println("error", "Invalid HA peer IP")
		return fmt.Errorf("Invalid HA peer IP")
	}

	if !isReachableUsingPing(peerIp) {
		log.Println("error", "HA peer is not reachable")
		return fmt.Errorf("HA peer is not reachable")
	}

	url := "https://" + peerIp + ":444/ha/disable"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("Error creating HTTP request: %v", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request to HA peer: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v", err)
	}

	fmt.Println("Response from HA peer:", string(responseBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error while disabling HA: " + string(responseBody))
	}

	return nil
}

func GetHa() model.APIResponse {
	var resp model.APIResponse

	config, err := interface_config.FetchConfig(interface_model.HA_STRING)
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	ha := config.Ha

	bytes, err := os.ReadFile(extras.HA_STATE_FILE)
	if err != nil {
		resp = model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		return resp
	}

	haState := strings.TrimSpace(string(bytes))

	if haState == extras.PRIMARY_STRING {
		ha.HaStatus.Status = "Master"
	} else if haState == extras.BACKUP_STRING {
		ha.HaStatus.Status = "Slave"
	} else {
		ha.HaStatus.Status = "None"
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, ha)
	return resp
}

var Vmac int

func createHaConfig(write, service bool) error {

	write = true
	service = false

	log.Println("info", "Creating HA Configuration")

	config, err := interface_config.FetchConfig(interface_model.HA_STRING)
	if err != nil {
		log.Println("error", "Error in fetching HA Configuration")
		return err
	}

	haConfig := config.Ha

	if haConfig.DedicatedHaInterface == extras.EMPTY_STRING {
		log.Println("info", "Dedicated HA Interface not found")
		return fmt.Errorf("Dedicated HA Interface not found")
	}

	conf, err := interface_config.FetchConfig("device")
	if err != nil {
		log.Println("error", "Error in fetching interfaces config")
		return err
	}

	intfs := conf.PhysicalInterfaces

	interfaceMap := make(map[string]interface_model.ListPhysicalInterface)

	for _, inf := range intfs {
		interfaceMap[inf.Name] = inf
	}

	// if dedicatedIntf.IpAddress.IpAddress == "" || dedicatedIntf.IpAddress.Protocol != "static" {
	// 	log.Println("info", "Dedicated HA Interface not configured to a static IP")
	// 	return fmt.Errorf("Dedicated HA Interface not configured to a static IP")
	// }

	primary := haConfig.ApplianceRole == extras.PRIMARY_STRING

	var grpStr strings.Builder
	grpStr.WriteString("")

	grpStr.WriteString("global_defs {\n    dynamic_interfaces\n\n}\n\n")
	grpStr.WriteString("vrrp_sync_group wijungle-cluster {\n    group {\n")

	fDisplayName := haConfig.DedicatedHaInterface
	grpStr.WriteString("        wijungle-cluster-" + fDisplayName + "\n")

	var intfStr strings.Builder
	intfStr.WriteString("")

	intfStr.WriteString(fmt.Sprintf("vrrp_instance wijungle-cluster-%s {\n", fDisplayName))
	if primary {
		intfStr.WriteString("    state MASTER\n")
	} else {
		intfStr.WriteString("    state BACKUP\n")
	}

	intfStr.WriteString(fmt.Sprintf("    interface %s\n", haConfig.DedicatedHaInterface))
	intfStr.WriteString("    virtual_router_id 100\n")
	intfStr.WriteString("    dont_track_primary\n")

	if Vmac == 1 {
		intfStr.WriteString("    use_vmac\n")
		intfStr.WriteString("    vmac_xmit_base\n")
	}

	if primary {
		intfStr.WriteString("    priority 10\n")
	} else {
		intfStr.WriteString("    priority 9\n")
	}

	intfStr.WriteString("    authentication {\n")
	intfStr.WriteString("        auth_type PASS\n")
	intfStr.WriteString(fmt.Sprintf("        auth_pass %s\n", strings.TrimSpace(haConfig.Password)))
	intfStr.WriteString("    }\n")

	virtualAddress := ""

	flushInterfaces := []string{}

	unicastString := "\n    unicast_peer {\n"
	unicastString += fmt.Sprintf("        %s\n", haConfig.PeerIp)
	unicastString += "    }"

	intfStr.WriteString(unicastString)

	trackInterfaceString := "\n    track_interface {"
	for _, intf := range haConfig.MonitoredInterfaces {
		// intfConf := interfaceMap[intf.Interface]
		// if intfConf.IpAddress.Protocol != "static" {
		// 	log.Println("info", "Monitored Interface not configured to a static IP, skipping it in HA Configuration")
		// 	continue
		// }
		trackInterfaceString += fmt.Sprintf("\n        %s weight 2", intf.Interface)
		flushInterfaces = append(flushInterfaces, intf.Interface)
	}
	trackInterfaceString += "\n    }\n\n"

	log.Println("info", "Track Interface String: ", trackInterfaceString)

	// err = os.WriteFile("/var/www/html/data/track_interfaces", []byte(trackInterfaceString), 0644)
	// if err != nil {
	// 	log.Println("error", "Error in writing track_interfaces file")
	// 	return err
	// }

	intfStr.WriteString(trackInterfaceString)

	intfStr.WriteString("    virtual_ipaddress {\n")

	for _, intf := range haConfig.MonitoredInterfaces {
		intfConf := interfaceMap[intf.Interface]
		// if intfConf.IpAddress.Protocol != "static" {
		// 	log.Println("info", "Monitored Interface not configured to a static IP, skipping it in HA Configuration")
		// 	continue
		// }

		if intfConf.IpAddress.IpAddress == "" {
			log.Println("info", "Monitored Interface ip not configured, skipping it in HA Configuration")
			continue
		}

		log.Println("info", "Monitored Interface name: ", intf.Interface)
		log.Println("info", "Monitored Interface config: ", intfConf)

		ip := intfConf.IpAddress.IpAddress + "/" + fmt.Sprint(intfConf.IpAddress.Netmask)

		if interface_utils.IsHaMonitored(intfConf.Name, haConfig) {
			ip = interface_utils.FetchInterfaceIpFromHaConfig(intfConf.Name, haConfig)
		}

		virtualAddress = fmt.Sprintf("        %s dev %s no_track\n", ip, intfConf.Name)
		// aliases := intfConf.AliasList
		// for _, alias := range aliases {
		// 	if alias.IpAddress != "" && alias.Netmask != 0 {
		// 		virtualAddress += fmt.Sprintf("        %s/%d dev %s no_track\n", alias.IpAddress, alias.Netmask, intfConf.Name)
		// 	}
		// }
		intfStr.WriteString(virtualAddress + "\n")
	}

	intfStr.WriteString("    }\n")

	// TODO: wan tables

	// for _, intf := range haConfig.MonitoredInterfaces {

	// }

	// wanStr := ""
	// wanTables := make(map[string][]string)

	// for _, intf := range haConfig.MonitoredInterfaces {
	// 	intfConf := interfaceMap[intf.Interface]
	// 	if intfConf.IpAddress.Protocol != "static" {
	// 		log.Println("info", "Monitored Interface not configured to a static IP, skipping it in HA Configuration")
	// 		continue
	// 	}

	// 	ip := intfConf.IpAddress.IpAddress
	// 	name := intfConf.Name
	// 	if _, ok := wanTables[name]; !ok {
	// 		wanTables[name] = make([]string, 0)
	// 	}
	// 	wanTables[name] = append(wanTables[name], ip)
	// }

	intfStr.WriteString("}\n")

	grpStr.WriteString("    }\n")
	grpStr.WriteString("    notify_master \"/var/www/html/data/ha/scripts/ha.sh primary\"\n")
	grpStr.WriteString("    notify_backup \"/var/www/html/data/ha/scripts/ha.sh backup\"\n")
	grpStr.WriteString("    notify_fault \"/var/www/html/data/ha/scripts/ha.sh fault\"\n")
	grpStr.WriteString("}\n")

	str := grpStr.String() + intfStr.String()

	//flush interfaces
	// for _, intf := range flushInterfaces {
	// 	cmd := exec.Command("ip", "address", "flush", "dev", intf)
	// 	err := cmd.Run()
	// 	if err != nil {
	// 		log.Println("error", "Error in flushing interface: ", intf)
	// 		return err
	// 	}
	// }

	// if write {
	// 	writeHa(str, service)
	// }

	// log.Println("HA Configuration: ", str)

	err = writeHa(str, false)
	if err != nil {
		return err
	}

	return nil
}

func writeHa(str string, service bool) error {
	err := os.WriteFile("/etc/keepalived/keepalived.conf", []byte(str), 0644)
	if err != nil {
		fmt.Println("Error writing keepalived.conf file:", err)
		return fmt.Errorf("Error writing keepalived.conf file: %v", err)
	}

	cmd := exec.Command("/bin/systemctl", "restart", "keepalived")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error restarting keepalived: %v", err)
		return fmt.Errorf("Error restarting keepalived service: %v", err)
	}
	return nil
}

func resetKeepAlived() error {
	err := exec.Command("/bin/systemctl", "stop", "keepalived").Run()
	if err != nil {
		return err
	}

	err = os.WriteFile("/etc/keepalived/keepalived.conf", []byte(""), 0644)
	if err != nil {
		fmt.Println("Error writing keepalived.conf file:", err)
		return fmt.Errorf("error writing keepalived.conf file: %v", err)
	}
	return nil
}

func IntializeSyncBackup() error {

	config, err := interface_config.FetchConfig(interface_model.HA_STRING)
	if err != nil {
		log.Println("error", "Error in fetching HA Configuration")
		return err
	}

	peerIp := config.Ha.PeerIp

	if peerIp == extras.EMPTY_STRING {
		// logger.LoggerFunc("error", logger.LoggerMessage("Invalid HA peer IP"))
		fmt.Println("error", "Invalid HA peer IP")
		return fmt.Errorf("invalid HA peer IP")
	}

	err = CompareDeviceInfo(peerIp)
	if err != nil {
		logger.LoggerFunc("error", logger.LoggerMessage("Error in comparing device info with HA peer"))
		fmt.Println("Error in comparing device info with HA peer:", err)
		return err
	}

	err = SendConfigToOtherDevice(peerIp)
	if err != nil {
		logger.LoggerFunc("error", logger.LoggerMessage("Error in sending config to other device"))
		fmt.Println("Error in sending config to other device:", err)
		return err
	}

	err = SyncBackupInAnotherAppliance(peerIp)
	if err != nil {
		logger.LoggerFunc("error", logger.LoggerMessage("Error in syncing backup in other device"))
		fmt.Println("Error in syncing backup:", err)
		return err
	}

	err = CreateHaForBackup(peerIp)
	if err != nil {
		logger.LoggerFunc("error", logger.LoggerMessage("Error in creating HA for backup"))
		fmt.Println("Error in creating HA for backup:", err)
		return err
	}

	return nil
}

func sendLastSyncedAtToPeer(haPeerIP string, t string) error {
	url := "https://" + haPeerIP + ":444/ha/update-last-synced-at"

	json := fmt.Sprintf(`{"last_synced_at": "%s"}`, t)

	req, err := http.NewRequest("POST", url, strings.NewReader(json))
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request to HA peer:", err)
		return fmt.Errorf("error sending request to HA peer: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return fmt.Errorf("error reading response body: %v", err)
	}

	fmt.Println("Response from HA peer:", string(responseBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error in sending last synced at to HA peer: " + string(responseBody))
	}

	return nil
}

func UpdateLastSyncedAt(req model.LastSyncedAt) error {
	config, err := interface_config.FetchConfig(interface_model.HA_STRING)
	if err != nil {
		return err
	}

	haConfig := config.Ha
	haConfig.HaStatus.LastSyncedAt = req.LastSyncedAt

	err = interface_config.UpdateConfig(interface_model.Config{Ha: haConfig}, interface_model.HA_STRING, "", "SYNC HA")
	if err != nil {
		return err
	}

	// go restartServices([]string{"keepalived", "backend"}, 8)

	go ServiceActions([]string{extras.SERVICE_KEEPALIVED, extras.SERVICE_BACKEND}, extras.Restart, 5)

	return nil
}

func CreateHaForBackup(peerIp string) error {

	if peerIp == extras.EMPTY_STRING {
		fmt.Println("error", "Invalid HA peer IP")
		return fmt.Errorf("Invalid HA peer IP")
	}

	if !isReachableUsingPing(peerIp) {
		log.Println("error", "HA peer is not reachable")
		return fmt.Errorf("HA peer is not reachable")
	}

	url := "https://" + peerIp + ":444/ha/generate-keepalived-config-for-backup"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("Error creating HTTP request: %v", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request to HA peer: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v", err)
	}

	fmt.Println("Response from HA peer:", string(responseBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error in sending config to HA peer: " + string(responseBody))
	}

	return nil
}

func GenerateKeepalivedConfigForBackupNsetIp() error {

	err := SetIpAfterSyncBackup()
	if err != nil {
		return fmt.Errorf("Error setting IP addresses after sync backup in backup appliance: %v", err)
	}

	err = createHaConfig(true, false)
	if err != nil {
		return fmt.Errorf("Error creating HA config: %v", err)
	}

	return nil
}

func SetIpAfterSyncBackup() error {
	config, err := interface_config.FetchConfig(interface_model.HA_STRING)
	if err != nil {
		log.Println("error", "Error in fetching HA Configuration")
		return err
	}

	resp, err := interface_config.FetchConfig(interface_model.DEVICE)
	if err != nil {
		fmt.Println("Failed to fetch physical interfaces from config file : ", err.Error())
		return err
	}

	for _, intf := range resp.PhysicalInterfaces {

		link, err := netlink.LinkByName(intf.Name)
		if err != nil {
			return err
		}

		if link.Attrs().Name == interface_model.LOOP_BACK_DEVICE {
			continue
		}

		if !strings.HasPrefix(link.Attrs().Name, "e") {
			continue
		}

		if intf.Name == config.Ha.DedicatedHaInterface {
			continue
		}

		err = interface_utils.PerformIpFlush(intf.Name)
		if err != nil {
			return err
		}

		if intf.IpAddress.IpAddress != interface_model.EMPTY_STRING && intf.IpAddress.Netmask != 0 {
			err = interface_utils.SetIP(intf.Name, intf.IpAddress.IpAddress, intf.IpAddress.Netmask)
			if err != nil {
				return err
			}
		}

		if interface_utils.IsHaMonitored(intf.Name, config.Ha) {

			err = interface_utils.PerformIpFlush(intf.Name)
			if err != nil {
				return err
			}

			baseIp, peerIp := interface_utils.FetchHaIps(intf.Name, config.Ha)
			if interface_utils.IsHaPrimary(config.Ha) && baseIp != extras.EMPTY_STRING {
				err = interface_utils.SetIP(intf.Name, baseIp, 24)
				if err != nil {
					return err
				}
			}

			if interface_utils.IsHaBackup(config.Ha) && peerIp != extras.EMPTY_STRING {
				err = interface_utils.SetIP(intf.Name, peerIp, 24)
				if err != nil {
					return err
				}
			}
		}

		intfResp, err := interfaces.ListPhysicalInterfaces(intf.Name, "HA")
		if err != nil {
			return err
		}

		interface_config.UpdateConfig(interface_model.Config{
			PhysicalInterfaces: intfResp.PhysicalInterfaces,
		}, interface_model.DEVICE, intf.Name, interface_model.HA_STRING)

		err = netlink.LinkSetUp(link)
		if err != nil {
			logger.LoggerFunc("error", logger.LoggerMessage("Error setting up interface "+intf.Name+" "+err.Error()))
		}

	}

	return nil
}

func restartServices(services []string, sleepTime int) {

	if sleepTime > 0 {
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	for _, service := range services {
		cmd := exec.Command("/bin/systemctl", "restart", service)
		err := cmd.Run()
		if err != nil {
			logger.LoggerFunc("error", logger.LoggerMessage("Error restarting service "+service+" "+err.Error()))
		}
	}
}

func ServiceActions(services []string, action extras.ServicesActionType, sleepTime int) {
	if sleepTime > 0 {
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	for _, service := range services {
		cmd := exec.Command("/bin/systemctl", action.String(), service)
		err := cmd.Run()
		if err != nil {
			logger.LoggerFunc("error", logger.LoggerMessage("Error while executing systemctl %s %s: "+err.Error()), action.String(), service)
		}
	}
}

func SyncBackupInAnotherAppliance(peerIp string) error {

	if peerIp == extras.EMPTY_STRING {
		log.Println("error", "Invalid HA peer IP")
		return fmt.Errorf("Invalid HA peer IP")
	}

	if !isReachableUsingPing(peerIp) {
		log.Println("error", "HA peer is not reachable")
		return fmt.Errorf("HA peer is not reachable")
	}

	url := "https://" + peerIp + ":444/ha/sync-backup"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("Error creating HTTP request: %v", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request to HA peer: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v", err)
	}

	fmt.Println("Response from HA peer:", string(responseBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error in sending config to HA peer: " + string(responseBody))
	}

	return nil
}

func CompareDeviceInfo(haPeerIP string) error {

	if haPeerIP == extras.EMPTY_STRING {
		log.Println("error", "Invalid HA peer IP")
		return fmt.Errorf("Invalid HA peer IP")

	}

	if !isReachableUsingPing(haPeerIP) {
		log.Println("error", "HA peer is not reachable")
		return fmt.Errorf("HA peer is not reachable")
	}

	deviceInfoPath := extras.ROOT_DATA_DEVICE_CONFIG
	if _, err := os.Stat(deviceInfoPath); os.IsNotExist(err) {
		log.Println("error", "Device info file not found")
		return err
	}

	deviceInfoContent, err := os.ReadFile(deviceInfoPath)
	if err != nil {
		log.Println("error", "Error in reading device info file")
		return err
	}

	deviceInfo := parseDeviceInfo(string(deviceInfoContent))
	config, err := interface_config.FetchConfig(interface_model.HA_STRING)
	if err != nil {
		log.Println("error", "Error in fetching HA Configuration")
		return err
	}

	deviceInfo.Password = config.Ha.Password
	payload := createPayload(deviceInfo)
	resp := sendRequestToHA(haPeerIP, payload)
	if resp.StatusCode != http.StatusOK {
		log.Println("error", "Error in sending request to HA peer")
		return fmt.Errorf("Error in sending request to HA peer: " + resp.Error)
	}

	return nil
}

func SendConfigToOtherDevice(haPeerIP string) error {
	if haPeerIP == extras.EMPTY_STRING {
		return fmt.Errorf("Invalid HA peer IP")
	}

	if !isReachableUsingPing(haPeerIP) {
		return fmt.Errorf("HA peer is not reachable")
	}

	interfaceConfigDataInBytes, err := os.ReadFile(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		return fmt.Errorf("Error reading interface config file: %v", err)
	}

	var interfaceConfig interface_model.Config
	if err := yaml.Unmarshal(interfaceConfigDataInBytes, &interfaceConfig); err != nil {
		return fmt.Errorf("Error unmarshalling interface config: %v", err)
	}

	for i := range interfaceConfig.PhysicalInterfaces {
		interfaceConfig.PhysicalInterfaces[i].LinkStats = interface_model.LinkStats{}
	}

	interfaceConfigDataInBytes, err = yaml.Marshal(interfaceConfig)
	if err != nil {
		return fmt.Errorf("Error marshalling interface config: %v", err)
	}

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	fileWriter, err := writer.CreateFormFile("configFile", "merged_config.yaml")
	if err != nil {
		return fmt.Errorf("Error creating form file field: %v", err)
	}

	_, err = fileWriter.Write(interfaceConfigDataInBytes)
	if err != nil {
		return fmt.Errorf("Error writing merged config contents: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("Error closing multipart writer: %v", err)
	}

	url := "https://" + haPeerIP + ":444/ha/copy-config"
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return fmt.Errorf("Error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request to HA peer: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v", err)
	}

	fmt.Println("Response from HA peer:", string(responseBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error in sending config to HA peer: " + string(responseBody))
	}

	return nil
}

func SyncBackup() error {

	configPath := extras.MERGED_CONFIG_FILE_NAME

	configContent, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("Error reading merged config file: %v", err)
	}

	var interfaceConfig interface_model.Config

	err = yaml.Unmarshal(configContent, &interfaceConfig)
	if err != nil {
		return fmt.Errorf("Error unmarshalling merged config content: %v", err)
	}

	err = writeInterfaceConfigFile(extras.INTERFACE_CONFIG_FILE_NAME, interfaceConfig)
	if err != nil {
		return fmt.Errorf("Error writing interface config file: %v", err)
	}

	return nil
}

func writeInterfaceConfigFile(filePath string, config interface_model.Config) error {

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	var oldConfig interface_model.Config
	err = yaml.Unmarshal(content, &oldConfig)
	if err != nil {
		return fmt.Errorf("Error unmarshalling old config: %v", err)
	}

	config.Ha.ApplianceRole = oldConfig.Ha.ApplianceRole
	config.Ha.PeerIp = oldConfig.Ha.PeerIp
	config.Ha.DedicatedHaInterface = oldConfig.Ha.DedicatedHaInterface

	oldPhyIntfs := oldConfig.PhysicalInterfaces
	newPhyIntfs := config.PhysicalInterfaces

	if len(oldPhyIntfs) != len(newPhyIntfs) {
		return fmt.Errorf("Error in updating interface config: Physical interfaces count mismatch")
	}

	for i, newIntf := range newPhyIntfs {
		if newIntf.Name == config.Ha.DedicatedHaInterface {
			ipInfo := oldPhyIntfs[i].IpAddress
			config.PhysicalInterfaces[i].IpAddress = ipInfo
		}
		// for _, intf := range config.Ha.MonitoredInterfaces {
		// 	if newIntf.Name == intf.Interface {
		// 		ipInfo := interface_model.IpAddressResponse{}
		// 		config.PhysicalInterfaces[i].IpAddress = ipInfo
		// 	}
		// }
		config.PhysicalInterfaces[i].AliasList = []interface_model.IpAddressResponse{}
	}

	configContent, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("Error marshalling config: %v", err)
	}

	err = os.WriteFile(filePath, configContent, 0644)
	if err != nil {
		return fmt.Errorf("Error writing config file: %v", err)
	}

	return nil
}

func overwriteConfigFile(originalFilePath, receivedFilePath string) error {
	originalContent, err := os.ReadFile(originalFilePath)
	if err != nil {
		return fmt.Errorf("Error reading original config file: %v", err)
	}

	receivedContent, err := os.ReadFile(receivedFilePath)
	if err != nil {
		return fmt.Errorf("Error reading received config file: %v", err)
	}

	var originalConfig interface_model.Config
	if err := yaml.Unmarshal(originalContent, &originalConfig); err != nil {
		return fmt.Errorf("Error unmarshalling original config file: %v", err)
	}

	var receivedConfig interface_model.Config
	if err := yaml.Unmarshal(receivedContent, &receivedConfig); err != nil {
		return fmt.Errorf("Error unmarshalling received config file: %v", err)
	}

	receivedConfig.Ha.ApplianceRole = originalConfig.Ha.ApplianceRole
	receivedConfig.Ha.PeerIp = originalConfig.Ha.PeerIp

	if err := createBackup(originalFilePath); err != nil {
		return fmt.Errorf("Error creating backup of the original config file: %v", err)
	}

	modifiedContent, err := yaml.Marshal(receivedConfig)
	if err != nil {
		return fmt.Errorf("Error marshalling modified config content: %v", err)
	}

	if err := os.WriteFile(originalFilePath, modifiedContent, 0644); err != nil {
		return fmt.Errorf("Error overwriting the original config file: %v", err)
	}

	return nil
}

func createBackup(filePath string) error {
	fContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Error reading config file while ha sync backup: %v", err)
	}

	backupFilePath := filePath + ".bak"

	err = os.WriteFile(backupFilePath, fContent, 0644)
	if err != nil {
		return fmt.Errorf("Error creating backup file: %v", err)
	}

	return nil
}

func parseDeviceInfo(content string) model.HaDeviceInfo {
	lines := strings.Split(content, "\n")
	deviceInfo := model.HaDeviceInfo{}

	for _, line := range lines {
		fields := strings.Split(line, "=")
		if len(fields) == 2 {
			key := strings.TrimSpace(fields[0])
			value := strings.TrimSpace(fields[1])

			switch key {
			case "serial_no":
				deviceInfo.SerialNo = value
			case "model_no":
				deviceInfo.ModelNo = value
			case "password":
				deviceInfo.Password = value
			}
		}
	}

	return deviceInfo
}

func isReachable(ip string) bool {
	conn, err := net.DialTimeout("tcp", ip+":80", 2)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func isReachableUsingPing(ip string) bool {
	cmd := exec.Command("ping", "-c", "1", ip)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func createPayload(deviceInfo model.HaDeviceInfo) string {
	payload := fmt.Sprintf(`{"serial_no": "%s", "model_no": "%s", "password": "%s"}`, deviceInfo.SerialNo, deviceInfo.ModelNo, deviceInfo.Password)
	return payload
}

func sendRequestToHA(haPeerIP, payload string) model.APIResponse {

	url := "https://" + haPeerIP + ":444/ha/compareDeviceInfo"

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		resp := model.NewErrorResponse(http.StatusBadRequest, "Error creating HTTP request", err)
		return resp
	}

	req.Header.Set("Content-Type", "application/json")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request to HA peer:", err)
		resp := model.NewErrorResponse(http.StatusBadRequest, "Error sending request to HA peer", err)
		return resp
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		resp := model.NewErrorResponse(http.StatusBadRequest, "Error reading response body", err)
		return resp
	}

	fmt.Println("Response from HA peer:", string(responseBody))

	var response model.APIResponse

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		resp := model.NewErrorResponse(http.StatusBadRequest, "Error unmarshalling response", err)
		return resp
	}
	return response
}

func CompareDeviceInfoFromAnotherAppliance(req model.HaDeviceInfo) error {
	deviceInfoPath := extras.ROOT_DATA_DEVICE_CONFIG
	if _, err := os.Stat(deviceInfoPath); os.IsNotExist(err) {
		log.Println("error", "Device info file not found")
		return err
	}

	deviceInfoContent, err := os.ReadFile(deviceInfoPath)
	if err != nil {
		log.Println("error", "Error in reading device info file")
		return err
	}

	deviceInfo := parseDeviceInfo(string(deviceInfoContent))

	if !compareDevicesInfo(deviceInfo, req) {
		log.Println("error", "Device info does not match")
		return fmt.Errorf("Device info does not match")
	}

	config, err := interface_config.FetchConfig(interface_model.HA_STRING)
	if err != nil {
		log.Println("error", "Error in fetching HA Configuration")
		return err
	}

	haConfig := config.Ha

	if haConfig.ApplianceRole == extras.EMPTY_STRING {
		log.Println("error", "HA config not found in backup appliance")
		return fmt.Errorf("HA config not found in backup appliance")
	}

	if haConfig.ApplianceRole != extras.BACKUP_STRING {
		log.Println("error", "Appliance role is not backup")
		return fmt.Errorf("Appliance role is not set to backup in backup appliance")
	}

	if haConfig.DedicatedHaInterface == extras.EMPTY_STRING {
		log.Println("error", "Dedicated HA Interface not found")
		return fmt.Errorf("Dedicated HA Interface not found in backup appliance")
	}

	if haConfig.PeerIp == extras.EMPTY_STRING {
		log.Println("error", "Peer IP not found")
		return fmt.Errorf("Peer IP not found in backup appliance")
	}

	if haConfig.Password != req.Password {
		log.Println("error", "Password does not match")
		return fmt.Errorf("Ha password does not match")
	}

	return nil
}

func compareDevicesInfo(deviceInfo1, deviceInfo2 model.HaDeviceInfo) bool {

	if deviceInfo1.ModelNo != deviceInfo2.ModelNo {
		return false
	}

	// if deviceInfo1.Password != deviceInfo2.Password {
	// 	return false
	// }

	return true
}

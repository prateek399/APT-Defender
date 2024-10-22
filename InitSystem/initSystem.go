package main

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"anti-apt-backend/model/interface_model"
	"anti-apt-backend/service/interfaces"
	"anti-apt-backend/util/interface_utils"
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
	"gopkg.in/yaml.v2"
)

var services = []string{"apache2", "backend", "cuckoo-rooter", "cuckoo", "RestAPI"}

func main() {

	setDeviceRebootedFlagTo1()

	stopServices()

	enableService()
	// interfaceConfig, err := readInterfaceConfig(extras.INTERFACE_CONFIG_FILE_NAME)
	// if err != nil {
	// 	log.Println("Error reading interface_config.yaml:", err)
	// }

	// config, err := interface_config.FetchConfig(interface_model.HA_STRING)
	// if err != nil {
	// 	log.Println("Error fetching Ha:", err)
	// }

	// err = setIPAddresses(interfaceConfig, config.Ha)
	// if err != nil {
	// 	log.Fatal("Error setting IP addresses:", err)
	// }

	err := interfaces.RestoreInterfaceSettings("INIT_SYSTEM")
	if err != nil {
		logger.LoggerFunc("error", logger.LoggerMessage("Error restoring interface settings in init system : "+err.Error()))
	}

	readCuckooConf() // Read cuckoo configuration file

	configureCuckoo()

	startServices()
}

func readInterfaceConfig(filename string) (*interface_model.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config interface_model.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func setIPAddresses(interfaceConfig *interface_model.Config, haConfig interface_model.Ha) error {
	if interfaceConfig == nil {
		return fmt.Errorf("interface config is nil")
	}

	physicalIntfs := interfaceConfig.PhysicalInterfaces

	for _, physicalInterface := range physicalIntfs {

		if !strings.HasPrefix(physicalInterface.Name, "e") {
			continue
		}

		if physicalInterface.IpAddress.Protocol != "static" {
			continue
		}

		log.Println("Setting IP for interface:", physicalInterface.Name)

		err := interface_utils.PerformIpFlush(physicalInterface.Name)
		if err != nil {
			return err
		}

		if interface_utils.IsHaMonitored(physicalInterface.Name, haConfig) {

			baseIp, peerIp := interface_utils.FetchHaIps(physicalInterface.Name, haConfig)

			if interface_utils.IsHaPrimary(haConfig) && baseIp != extras.EMPTY_STRING {
				ip := baseIp
				err := interface_utils.SetIP(physicalInterface.Name, ip, physicalInterface.IpAddress.Netmask)
				if err != nil {
					return err
				}
			} else if interface_utils.IsHaBackup(haConfig) && peerIp != extras.EMPTY_STRING {
				ip := peerIp
				err := interface_utils.SetIP(physicalInterface.Name, ip, physicalInterface.IpAddress.Netmask)
				if err != nil {
					return err
				}
			}
		} else {
			ip := physicalInterface.IpAddress.IpAddress
			if ip == extras.EMPTY_STRING {
				continue
			}
			err := interface_utils.SetIP(physicalInterface.Name, ip, physicalInterface.IpAddress.Netmask)
			if err != nil {
				return err
			}
		}

		link, err := netlink.LinkByName(physicalInterface.Name)
		if err != nil {
			return err
		}
		err = netlink.LinkSetUp(link)
		if err != nil {
			return err
		}

	}
	return nil
}

func stopServices() {
	for _, service := range services {
		cmd := exec.Command("/bin/systemctl", "stop", service)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Println("Error stopping service:", err)
		}
		log.Println("Service stopped successfully:", service)
	}
}

func enableService() {
	cmd := exec.Command("systemctl", "enable", "backend")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Println("Error enabling service:", err)
	}
	log.Println("Service enabled successfully:", "backend")
}

func startServices() {
	for _, service := range services {
		cmd := exec.Command("/bin/systemctl", "start", service)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Println("Error starting service:", err)
		}
		log.Println("Service started successfully:", service)
	}
}

func readCuckooConf() {
	file, err := os.Open(extras.CUCKOO_CONF_FILE_PATH)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) >= 2 {
			if strings.Contains(line, "version_check") {
				extras.VERSION_CHECK = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "ignore_vulnerabilities") {
				extras.IGNORE_VULNERABLILITY = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "delete_original") {
				extras.DELETE_ORIGINAL = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "delete_bin_copy") {
				extras.DELETE_BIN_COPY = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "machinery") {
				extras.MACHINERY = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "memory_dump") {
				extras.MEMORY_DUMP = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "terminate_processes") {
				extras.TERMINATE_PROCESSES = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "reschedule") {
				extras.RESCHEDULE = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "process_results") {
				extras.PROCESS_RESULTS = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "max_analysis_count") {
				extras.MAX_ANALYSIS_COUNT, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err != nil {
					extras.MAX_ANALYSIS_COUNT = 10
				}
			}
			if strings.Contains(line, "max_machines_count") {
				extras.MAX_MACHINES_COUNT, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err != nil {
					extras.MAX_MACHINES_COUNT = 8
				}
			}
			if strings.Contains(line, "max_vmstartup_count") {
				extras.MAX_VM_STARTUP_COUNT, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err != nil {
					extras.MAX_VM_STARTUP_COUNT = 8
				}
			}
			if strings.Contains(line, "freespace") {
				extras.FREESPACE, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err != nil {
					extras.FREESPACE = 8
				}
			}
			if strings.Contains(line, "tmppath") {
				extras.TMP_PATH = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "rooter") {
				extras.ROOTER = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "ip") {
				extras.SOURCE_IP = strings.TrimSpace(parts[1])
			}
			if strings.Contains(line, "port") {
				extras.SOURCE_PORT, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err != nil {
					extras.SOURCE_PORT = 2042
				}
			}
			if strings.Contains(line, "upload_max_size") {
				extras.MAX_ALLOWED_FILE_SIZE, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err != nil {
					extras.MAX_ALLOWED_FILE_SIZE = 10485760
				}
			}
			if strings.Contains(line, "vm_state") {
				extras.VM_STATE, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err != nil {
					extras.VM_STATE = 60
				}
			}
			if strings.Contains(line, "critical") {
				extras.CRITICAL_TIMEOUT, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err != nil {
					extras.CRITICAL_TIMEOUT = 60
				}
			}
			if strings.Contains(line, "analysis_size_limit") {
				extras.ANALYSIS_SIZE_LIMIT, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				if err != nil {
					extras.ANALYSIS_SIZE_LIMIT = 134217728
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return
	}
}

func configureCuckoo() {

	// link, err := netlink.LinkByName("enp8s0")
	// if err != nil {
	// 	log.Fatal("Error fetching link in init script:", err)
	// }

	// err = netlink.LinkSetUp(link)
	// if err != nil {
	// 	log.Fatal("Error setting link up in init script:", err)
	// }

	lastInterface := getLastInterfaceName()
	if lastInterface == "" {
		logger.LoggerFunc("error", logger.LoggerMessage("No interface found which is not disabled"))
	}

	cmds := []string{
		// "/sbin/dhclient enp8s0",
		"/bin/systemctl restart systemd-resolved",
		"mount -o ro,loop /log/iso/win7ultimate.iso /mnt/win7",
		`sudo su wijungle -c "source /usr/share/virtualenvwrapper/virtualenvwrapper.sh && source /home/wijungle/.virtualenvs/cuckoo-test/bin/activate && bash -c 'vmcloak-vboxnet0'"`,
		"sysctl -w net.ipv4.conf.vboxnet0.forwarding=1",
		// run these commands only if network zone is WAN
		"sysctl -w net.ipv4.conf.vboxnet0.forwarding=1",
		"iptables -w -t nat -A POSTROUTING -o " + lastInterface + " -s 192.168.56.0/24 -j MASQUERADE",
		"iptables -w -P FORWARD DROP",
		"iptables -w -A FORWARD -m state --state RELATED,ESTABLISHED -j ACCEPT",
		"iptables -w -A FORWARD -s 192.168.56.0/24 -j ACCEPT",
		"service cuckoo-rooter start",
		"/bin/systemctl daemon-reload",
		"service cuckoo start",
		"service RestAPI start",
	}

	for _, cmd := range cmds {
		err := executeCommand(cmd)
		if err != nil {

			if strings.Contains(cmd, "mount") {
				log.Println("Error mounting iso in init system : " + err.Error())
			} else {
				log.Fatal("Error executing command:", err)
			}
		}
		log.Println("Command executed successfully:", cmd)
	}
}

func executeCommand(cmd string) error {
	command := exec.Command("bash", "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		log.Println("Command:", cmd)
		log.Println("Error executing command:", err)
		log.Println("Stdout:", command.Stdout)
		log.Println("Stderr:", command.Stderr)
	}
	return err
}

func getLastInterfaceName() string {
	interfaceConfig, err := readInterfaceConfig(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		logger.LoggerFunc("error", logger.LoggerMessage("Error reading interface settings in init system : "+err.Error()))
	}
	var lastInterface string
	if interfaceConfig == nil {
		links, err := netlink.LinkList()
		if err != nil {
			logger.LoggerFunc("error", logger.LoggerMessage("Error fetching link list in init system : "+err.Error()))
		}
		for _, link := range links {
			if !strings.HasPrefix(link.Attrs().Name, "e") || link.Attrs().Name == interface_model.LOOP_BACK_DEVICE {
				continue
			}
			lastInterface = link.Attrs().Name
		}
		return lastInterface
	}
	physicalIntfs := interfaceConfig.PhysicalInterfaces

	for _, physicalInterface := range physicalIntfs {
		if !strings.HasPrefix(physicalInterface.Name, "e") && !physicalInterface.IsDisabled {
			continue
		}
		lastInterface = physicalInterface.Name
	}

	return lastInterface
}

func setDeviceRebootedFlagTo1() {
	filePath := "/etc/device_rebooted"

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error creating file: %v\n", err)
			return
		}
		defer file.Close()
	} else if err != nil {
		fmt.Printf("Error checking file: %v\n", err)
		return
	}

	err = os.WriteFile(filePath, []byte("1"), 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
}

package interface_config

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model/interface_model"
	model "anti-apt-backend/model/interface_model"
	"anti-apt-backend/util/interface_utils"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/ghodss/yaml"
)

type RoutingConfigService interface {
	UpdateStaticRouteConfig(resp model.Config, operation string, caller string) error
	FetchRouteConfig(routeType string) (model.Config, error)
}

type routingService struct {
	ipsecService RoutingConfigService
}

func NewRoutingConfigService() RoutingConfigService {
	return &routingService{}
}

func (r *routingService) UpdateStaticRouteConfig(resp model.Config, operation string, caller string) error {
	if ok := mu.TryLock(LOCK_TIME_OUT); !ok {
		return fmt.Errorf("Timeout while acquiring lock for %s - %s", operation, caller)
	} else {
		fmt.Printf("Lock acquired for %s - %s\n", operation, caller)
	}
	defer func() {
		mu.Unlock()
		fmt.Printf("Lock released for %s - %s\n", operation, caller)
	}()

	yamlData, err := readConfigFile(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		return err
	}

	var config model.Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return err
	}

	switch operation {
	case model.Type_IPV4_UNICAST:
		var newData model.ListStaticRoutes
		yamlResp, err := yaml.Marshal(resp.StaticRoutes)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlResp, &newData); err != nil {
			return err
		}
		fmt.Println("newData", newData)
		config.StaticRoutes.Ipv4UnicastRoutes = newData.Ipv4UnicastRoutes
	case model.Type_IPV6_UNICAST:
		var newData model.ListStaticRoutes
		yamlResp, err := yaml.Marshal(resp.StaticRoutes)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlResp, &newData); err != nil {
			return err
		}
		config.StaticRoutes.Ipv6UnicastRoutes = newData.Ipv6UnicastRoutes
	case model.Type_MULTICAST:
		var newData model.ListStaticRoutes
		yamlResp, err := yaml.Marshal(resp.StaticRoutes)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlResp, &newData); err != nil {
			return err
		}
		config.StaticRoutes.MulticastRoutes = newData.MulticastRoutes
	case model.Type_STATIC:
		var newData model.ListStaticRoutes
		yamlResp, err := yaml.Marshal(resp.StaticRoutes)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlResp, &newData); err != nil {
			return err
		}
		config.StaticRoutes.GlobalConfig = newData.GlobalConfig
	}

	updatedYAMLData, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(extras.INTERFACE_CONFIG_FILE_NAME, updatedYAMLData, 0644); err != nil {
		return err
	}

	go initRoute()

	fmt.Println("Config file updated successfully")
	return nil
}

func (r *routingService) FetchRouteConfig(routeType string) (model.Config, error) {
	resp := model.Config{}

	yamlData, err := readConfigFile(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		return resp, err
	}

	var config model.Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return resp, err
	}

	switch routeType {
	case model.STATIC_ROUTE:
		resp.StaticRoutes = config.StaticRoutes
	case model.BGP:
		resp.BGP = config.BGP
	case model.OSPF:
		resp.OSPF = config.OSPF
	case model.OSPFV3:
		resp.OSPFV3 = config.OSPFV3
	}

	return resp, nil
}

func generateZebraConfig() error {

	rs := NewRoutingConfigService()
	routes, err := rs.FetchRouteConfig(model.STATIC_ROUTE)
	if err != nil {
		log.Println("Error fetching static routes config while generating zebra config")
		return err
	}

	log.Println("Dumping Static routes")

	var baseStr strings.Builder

	baseStr.WriteString("!")
	baseStr.WriteString("\npassword 1!3@A$%")
	baseStr.WriteString("\nlog file /var/log/quagga/zebra.log")
	baseStr.WriteString("\nservice advanced-vty")
	baseStr.WriteString("\n!")

	var interfaceStr strings.Builder
	interfaceStr.WriteString("")

	conf, err := FetchConfig(model.DEVICE)
	if err != nil {
		log.Println("Error fetching config while generating zebra config")
		return err
	}

	intfs := conf.PhysicalInterfaces

	config, err := FetchConfig(interface_model.HA_STRING)
	if err != nil {
		log.Println("Error fetching ha config while generating zebra config")
		return err
	}

	for _, intf := range intfs {
		if interface_utils.IsHaMonitored(intf.Name, config.Ha) {
			var tempIp string
			baseip, peerip := interface_utils.FetchHaIps(intf.Name, config.Ha)
			if interface_utils.IsHaPrimary(config.Ha) {
				tempIp = baseip
			} else if interface_utils.IsHaBackup(config.Ha) {
				tempIp = peerip
			}

			validIp, _ := interface_utils.ValidateIP(tempIp)

			if tempIp != "" && validIp {
				interfaceStr.WriteString(fmt.Sprintf("\ninterface %s", intf.Name))
			}

			if strings.Contains(tempIp, ".") {
				interfaceStr.WriteString(fmt.Sprintf("\nip address %s/%d", tempIp, intf.IpAddress.Netmask))
			}

			if strings.Contains(tempIp, ":") {
				interfaceStr.WriteString(fmt.Sprintf("\nipv6 address %s/%d", tempIp, intf.IpAddress.Netmask))
			}
		} else {
			interfaceStr.WriteString(fmt.Sprintf("\ninterface %s", intf.Name))

			ip := intf.IpAddress.IpAddress

			if strings.Contains(ip, ".") {
				interfaceStr.WriteString(fmt.Sprintf("\nip address %s/%d", ip, intf.IpAddress.Netmask))
			}

			if strings.Contains(ip, ":") {
				interfaceStr.WriteString(fmt.Sprintf("\nipv6 address %s/%d", ip, intf.IpAddress.Netmask))
			}

			aliases := intf.AliasList
			for _, alias := range aliases {
				ip := alias.IpAddress
				if ip == "" || alias.Netmask == 0 {
					continue
				}
				if strings.Contains(ip, ".") {
					interfaceStr.WriteString(fmt.Sprintf("\nip address %s/%d", ip, alias.Netmask))
				}

				if strings.Contains(ip, ":") {
					interfaceStr.WriteString(fmt.Sprintf("\nipv6 address %s/%d", ip, alias.Netmask))
				}
			}

		}
	}

	var routeStr strings.Builder

	routeStr.WriteString("")

	for _, route := range routes.StaticRoutes.Ipv4UnicastRoutes {
		routeAd := 1
		if route.AdministrativeDistance > 0 {
			routeAd = route.AdministrativeDistance
		}
		if route.Gateway != "" {
			routeStr.WriteString(fmt.Sprintf("\nip route %s/%s %s %d", route.DestinationIp, route.Netmask, route.Gateway, routeAd))
		} else {
			routeStr.WriteString(fmt.Sprintf("\nip route %s/%s %s %d", route.DestinationIp, route.Netmask, route.InterfaceName, routeAd))
		}
	}

	finalStr := baseStr.String() + interfaceStr.String() + routeStr.String()
	log.Println("Final Zebra Config: ", finalStr)
	os.WriteFile("/var/www/html/data/route/zebra.conf", []byte(finalStr), 0644)

	return nil
}

func writeZebraConfig(dump bool) {
	log.Println("Writing zebra config")
	if dump {
		configZebraFile()
	}

	exec.Command("sudo", "/bin/cp", "-f", "/var/www/html/data/route/daemons", "/etc/quagga/").Run()
	if _, err := os.Stat("/etc/quagga/zebra.conf"); os.IsNotExist(err) {
		exec.Command("touch", "/etc/quagga/zebra.conf").Run()
		exec.Command("chown", "quagga:quagga", "/etc/quagga/zebra.conf").Run()
	}
	os.WriteFile("/var/www/html/data/route/debian.conf", []byte("vtysh_enable=yes"), 0644)
	exec.Command("sudo", "/bin/cp", "-f", "/var/www/html/data/route/zebra.conf", "/etc/quagga/").Run()
	exec.Command("sudo", "/bin/cp", "-f", "/var/www/html/data/route/debian.conf", "/etc/quagga/").Run()

	exec.Command("sudo", "/usr/bin/pkill", "-9", "zebra").Run()
	exec.Command("sudo", "chown", "-R", "quagga:quagga", "/etc/quagga/").Run()
	exec.Command("sudo", "/bin/systemctl", "restart", "zebra").Run()
	exec.Command("sudo", "/bin/systemctl", "restart", "ospfd").Run()
	exec.Command("sudo", "/bin/systemctl", "restart", "bgpd").Run()
	exec.Command("sudo", "/bin/systemctl", "restart", "ripd").Run()
	exec.Command("sudo", "/bin/systemctl", "restart", "ospf6d").Run()
	exec.Command("sudo", "/bin/systemctl", "restart", "isisd").Run()
	exec.Command("sudo", "/bin/systemctl", "restart", "babeld").Run()
	exec.Command("sudo", "/bin/systemctl", "restart", "ripngd").Run()

}

func configZebraFile() {
	log.Println("Configuring zebra file")
	var str strings.Builder
	str.WriteString("")
	str.WriteString("\nospfd=no")
	str.WriteString("\nbgpd=no")
	str.WriteString("\nripd=no")
	str.WriteString("\nripngd=no")
	str.WriteString("\nospf6d=no")
	str.WriteString("\nisisd=no")
	str.WriteString("\nbabeld=no")
	finalStr := "zebra=yes" + str.String()
	log.Println("daemonstr: ", finalStr)
	os.WriteFile("/var/www/html/data/route/daemons", []byte(finalStr), 0644)
}

func initRoute() {
	log.Println("Initializing route")
	err := generateZebraConfig()
	if err != nil {
		log.Println("Error generating zebra config")
		return
	}
	writeZebraConfig(true)
}

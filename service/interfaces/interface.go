package interfaces

import (
	"anti-apt-backend/config/interface_config"
	config "anti-apt-backend/config/interface_config"
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"anti-apt-backend/model/interface_model"
	model "anti-apt-backend/model/interface_model"
	"anti-apt-backend/service/interfaces/bond"
	"anti-apt-backend/service/interfaces/bridge"
	"anti-apt-backend/service/interfaces/vlan"
	"anti-apt-backend/util/interface_utils"
	utils "anti-apt-backend/util/interface_utils"
	validations "anti-apt-backend/validation/interface_validations"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/safchain/ethtool"
	"github.com/vishvananda/netlink"
)

func FetchIps() ([]string, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, link := range links {
		if link.Type() == model.DEVICE {
			if link.Attrs().Name == model.LOOP_BACK_DEVICE {
				continue
			}

			if !strings.HasPrefix(link.Attrs().Name, "e") {
				continue
			}

			addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
			if err != nil {
				log.Println("Failed to fetch IP addresses for interface : ", link.Attrs().Name)
			}

			for _, addr := range addrs {
				ips = append(ips, addr.IP.String())
			}

		}
	}
	return ips, nil
}

func ListPhysicalInterfaces(intfName string, caller string) (resp *model.ListPhysicalInterfacesResponse, err error) {

	if intfName != model.EMPTY_STRING {
		link, err := netlink.LinkByName(intfName)
		if err != nil {
			return nil, fmt.Errorf("interface %s not found", intfName)
		}
		if link.Type() != model.DEVICE {
			return nil, fmt.Errorf("Interface %s is a type of %s, not a physical interface", intfName, link.Type())
		}
	}

	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	var respLinks []model.ListPhysicalInterface
	configFields, err := config.FetchConfigSpecificFields(model.DEVICE)
	if err != nil {
		fmt.Printf("Failed to fetch config specific fields from config file : %s", err.Error())
	}

	for _, link := range links {
		if link.Type() == model.DEVICE && (intfName == model.EMPTY_STRING || link.Attrs().Name == intfName) {
			if link.Attrs().Name == model.LOOP_BACK_DEVICE {
				continue
			}

			// skip if the interface name doesn't start with e
			if !strings.HasPrefix(link.Attrs().Name, "e") {
				continue
			}

			speed, duplex, autoneg, err := getLinkSpeedDuplex(link.Attrs().Name)
			if err != nil {
				return nil, err
			}
			attachedTo, err := fetchAttachedTo(link.Attrs().Name)
			if err != nil {
				return nil, err
			}

			physicalConfigFields, exists := configFields[link.Attrs().Name]
			if !exists {
				physicalConfigFields = model.ConfigSpecificFields{}
			}

			var isEditable bool
			if len(attachedTo) == 0 {
				isEditable = true
			}

			config, err := interface_config.FetchConfig(interface_model.HA_STRING)
			if err != nil {
				return nil, fmt.Errorf("Error in fetching HA data : %s", err.Error())
			}

			isMonitored := interface_utils.IsHaMonitored(link.Attrs().Name, config.Ha)
			isHaDedicated := interface_utils.IsHaDedicated(link.Attrs().Name, config.Ha)

			if isMonitored || isHaDedicated {
				isEditable = false
			}

			ipAddress := utils.GetPrimaryIPAddressV2(link)

			intfHaIp := interface_utils.FetchInterfaceIpFromHaConfig(link.Attrs().Name, config.Ha)
			if isMonitored && intfHaIp != model.EMPTY_STRING && caller == "MAIN" {
				ipAddress.IpAddress = strings.Split(intfHaIp, "/")[0]
				nm, err := strconv.Atoi(strings.Split(intfHaIp, "/")[1])
				if err != nil {
					return nil, fmt.Errorf("Error in converting string to int : %s", err.Error())
				}
				ipAddress.Netmask = nm
			}

			aliasList := utils.GetSecondaryIPAddressListV2(link)
			respLinks = append(respLinks, model.ListPhysicalInterface{
				Name:            link.Attrs().Name,
				MTU:             link.Attrs().MTU,
				HardwareAddress: link.Attrs().HardwareAddr.String(),
				IpAddress:       ipAddress,
				AliasList:       aliasList,
				LinkState:       link.Attrs().OperState.String(),
				LinkStats: model.LinkStats{
					TxPackets: link.Attrs().Statistics.TxPackets,
					TxBytes:   link.Attrs().Statistics.TxBytes,
					TxErrors:  link.Attrs().Statistics.TxErrors,
					TxDropped: link.Attrs().Statistics.TxDropped,
					RxPackets: link.Attrs().Statistics.RxPackets,
					RxBytes:   link.Attrs().Statistics.RxBytes,
					RxErrors:  link.Attrs().Statistics.RxErrors,
					RxDropped: link.Attrs().Statistics.RxDropped,
				},
				LinkSpeed:       speed,
				LinkDuplex:      duplex,
				LinkAutoneg:     autoneg,
				AttachedTo:      attachedTo,
				IsEditable:      isEditable,
				IsDeletable:     false,
				IsDisabled:      physicalConfigFields.IsDisabled,
				ServingLocation: physicalConfigFields.ServingLocation,
				DomainName:      physicalConfigFields.DomainName,
				NetworkZone:     physicalConfigFields.NetworkZone,
			})

			if intfName != model.EMPTY_STRING {
				break
			}
		}
	}

	return &model.ListPhysicalInterfacesResponse{
		PhysicalInterfacesCount: len(respLinks),
		PhysicalInterfaces:      respLinks,
	}, nil
}

func getLinkSpeedDuplex(ifname string) (string, string, string, error) {
	if ifname == model.LOOP_BACK_DEVICE {
		return "N/A", "N/A", "N/A", nil
	}
	e, err := ethtool.NewEthtool()
	if err != nil {
		return extras.EMPTY_STRING, extras.EMPTY_STRING, extras.EMPTY_STRING, err
	}
	defer e.Close()

	m, err := e.CmdGetMapped(ifname)
	if err != nil {
		return extras.EMPTY_STRING, extras.EMPTY_STRING, extras.EMPTY_STRING, err
	}

	// fmt.Printf("Map for %s is %v\n", ifname, m)

	var speed, duplex, autoneg string

	if speedInt, ok := m["Speed"]; !ok || speedInt == 65535 {
		speed = "Unknown"
	} else {
		speed = fmt.Sprintf("%v Mbps", speedInt)
	}

	if duplexInt, ok := m["Duplex"]; !ok {
		duplex = "Unknown"
	} else if duplexInt == 0 {
		duplex = "Half"
	} else if duplexInt == 1 {
		duplex = "Full"
	} else {
		duplex = "Unknown"
	}

	if autonegInt, ok := m["Autoneg"]; !ok {
		autoneg = "Unknown"
	} else if autonegInt == 0 {
		autoneg = "OFF"
	} else if autonegInt == 1 {
		autoneg = "ON"
	} else {
		autoneg = "Unknown"
	}
	return speed, duplex, autoneg, nil
}

func fetchAttachedTo(linkName string) (resp []string, err error) {
	bridges, err := bridge.ListBridgeInterfaces(model.ListBridgeInterfaceRequest{})
	if err != nil {
		return nil, fmt.Errorf("Error in fetching bridge interfaces")
	}
	for _, bridge := range bridges.BridgeInterfaces {
		for _, member := range bridge.MemberInterfaces {
			if member.InterfaceName == linkName {
				resp = append(resp, bridge.BridgeInterfaceName)
			}
		}
	}

	bonds, err := bond.ListBondInterfaces(model.ListBondInterfacesRequest{})
	if err != nil {
		return nil, fmt.Errorf("Error in fetching bond interfaces")
	}
	for _, bond := range bonds.BondInterfaces {
		for _, slave := range bond.SlaveInterfaces {
			if slave.InterfaceName == linkName {
				resp = append(resp, bond.BondInterfaceName)
			}
		}
	}

	vlans, err := vlan.ListVlanInterfaces(model.ListVlanInterfacesRequest{})
	if err != nil {
		return nil, fmt.Errorf("Error in fetching vlan interfaces")
	}
	for _, vlan := range vlans.VlanInterfaces {
		if vlan.ParentInterface.InterfaceName == linkName {
			resp = append(resp, vlan.VlanInterfaceName)
		}
	}
	return resp, nil
}

func UpdatePhysicalInterface(physicalInterfaceName string, request model.UpdatePhysicalInterfaceRequest, curUsr string) (*model.ListPhysicalInterfacesResponse, error) {

	req, ok := utils.TrimStringsInStruct(request).(model.UpdatePhysicalInterfaceRequest)
	if !ok {
		return nil, fmt.Errorf("Failed to parse JSON request")
	}

	physicalInterfaceName = strings.TrimSpace(physicalInterfaceName)

	link, err := netlink.LinkByName(physicalInterfaceName)
	if err != nil {
		return nil, fmt.Errorf("Physical interface %s not found", physicalInterfaceName)
	}

	if link.Type() != model.DEVICE {
		return nil, fmt.Errorf("Interface %s is a type of %s, not a physical interface", physicalInterfaceName, link.Type())
	}

	haConfig, err := config.FetchConfig(interface_model.HA_STRING)
	if err != nil {
		return nil, fmt.Errorf("Error in fetching HA data : %s", err.Error())
	}

	isMonitored := interface_utils.IsHaMonitored(physicalInterfaceName, haConfig.Ha)
	isHaDedicated := interface_utils.IsHaDedicated(physicalInterfaceName, haConfig.Ha)

	if isMonitored || isHaDedicated {
		return nil, fmt.Errorf("%s is a dedicated port or a port monitored by HA, cannot update", PortMapping[physicalInterfaceName])
	}

	err = validations.ValidateUpdatePhysicalInterfaceRequest(req, physicalInterfaceName)
	if err != nil {
		return nil, err
	}

	attachedTo, err := fetchAttachedTo(physicalInterfaceName)
	if err != nil {
		return nil, err
	}

	if len(attachedTo) > 0 {
		return nil, fmt.Errorf("Physical interface %s is attached to %v, cannot update", physicalInterfaceName, attachedTo)
	}

	hwaddr := req.HardwareAddress
	if strings.TrimSpace(hwaddr) != model.EMPTY_STRING && hwaddr != link.Attrs().HardwareAddr.String() {
		hwAddr, err := utils.ValidateAndGetHardwareAddress(hwaddr)
		if err != nil {
			return nil, fmt.Errorf("Invalid hardware address: %s", err.Error())
		}
		err = netlink.LinkSetHardwareAddr(link, hwAddr)
		if err != nil {
			return nil, fmt.Errorf("Failed to update hardware address for physical interface %s", PortMapping[physicalInterfaceName])
		}
	}

	if request.IPv4Details.IPv4 {
		if request.IPv4Details.IPv4AssignmentMode == model.STATIC_STRING {

			primaryIp := utils.GetPrimaryIPAddress(link)

			if request.IPv4Details.IPAddress+"/"+request.IPv4Details.Netmask != primaryIp {
				err = utils.PerformIpDel(link, primaryIp, req.IPv4Details.IPAddress, req.IPv4Details.Netmask)
				if err != nil {
					return nil, fmt.Errorf("Failed to delete IP address from physical interface %s with error : %v", PortMapping[physicalInterfaceName], err)
				}
				logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("sysLog:IP address on %s has been updated from %s to %s by %s", PortMapping[physicalInterfaceName], primaryIp, req.IPv4Details.IPAddress, curUsr)))
			}
		} else if request.IPv4Details.IPv4AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got update request for DHCP V4 for physical interface : ", physicalInterfaceName)
		}
	} else if request.IPv6Details.IPv6 {
		if request.IPv6Details.IPv6AssignmentMode == model.STATIC_STRING {

			primaryIp := utils.GetPrimaryIPAddress(link)

			if request.IPv6Details.IPAddress+"/"+request.IPv6Details.Prefix != primaryIp {
				err = utils.PerformIpDel(link, primaryIp, req.IPv6Details.IPAddress, req.IPv6Details.Prefix)
				if err != nil {
					return nil, fmt.Errorf("Failed to delete IP address from physical interface %s with error : %v", physicalInterfaceName, err)
				}
				logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("sysLog:IP address on %s has been updated from %s to %s by %s", PortMapping[physicalInterfaceName], primaryIp, req.IPv4Details.IPAddress, curUsr)))
			}
		} else if request.IPv6Details.IPv6AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got update request for DHCP V6 for physical interface : ", physicalInterfaceName)
		}
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		fmt.Println("Failed to bring up physical interface %s", PortMapping[physicalInterfaceName])
	}

	configFields, err := config.FetchConfigSpecificFields(model.DEVICE)
	if err != nil {
		fmt.Printf("Failed to fetch config specific fields from config file : %s", err.Error())
	}

	err = updatePhysicalConfig(physicalInterfaceName, "EDIT PHYSICAL INTERFACE")
	if err != nil {
		return nil, err
	}

	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("sysLog:configuration of %s has been updated by %s", PortMapping[physicalInterfaceName], curUsr)))

	physicalConfigFields, exists := configFields[physicalInterfaceName]
	if !exists {
		physicalConfigFields = model.ConfigSpecificFields{}
	}

	if req.ServingLocation != physicalConfigFields.ServingLocation || req.DomainName != physicalConfigFields.DomainName || req.NetworkZone != physicalConfigFields.NetworkZone || req.IsDisabled != physicalConfigFields.IsDisabled {
		err = config.UpdateConfigSpecificFields(model.DEVICE, physicalInterfaceName, model.ConfigSpecificFields{
			ServingLocation: req.ServingLocation,
			DomainName:      req.DomainName,
			NetworkZone:     req.NetworkZone,
			IsDisabled:      req.IsDisabled,
		}, "EDIT PHYSICAL INTERFACE (config specific fields)")
		if err != nil {
			return nil, fmt.Errorf("Physical interface has been updated but failed to update the config file : %s", err.Error())
		}
	}

	physicalInterfaces, err := ListPhysicalInterfaces(model.EMPTY_STRING, "UPDATE")
	if err != nil {
		return nil, fmt.Errorf("Physical interface has been updated but failed to fetch : %s", err.Error())
	}
	return physicalInterfaces, nil
}

func updatePhysicalConfig(physicalInterfaceName string, caller string) (err error) {
	var configResp model.Config
	if config.CheckIfInterfaceTypeIsEmpty(model.DEVICE) {
		InitPhysicalInterfacesConfig()
	} else {
		respForConfig, err := ListPhysicalInterfaces(physicalInterfaceName, "UPDATE")
		if err != nil {
			return fmt.Errorf("Physical interface has been created but failed to fetch the updated physical interfaces : %s", err.Error())
		}
		if len(respForConfig.PhysicalInterfaces) > 0 && respForConfig.PhysicalInterfaces[0].Name == physicalInterfaceName {
			configResp.PhysicalInterfaces = []model.ListPhysicalInterface{respForConfig.PhysicalInterfaces[0]}
			if caller == "LINK UP" {
				configResp.PhysicalInterfaces[0].LinkState = model.LINK_STATE_UP
			}
			err = config.UpdateConfig(configResp, model.DEVICE, physicalInterfaceName, caller)
			if err != nil {
				return fmt.Errorf("Physical interface has been created but failed to update the config file : %s", err.Error())
			}
		}
	}

	return nil
}

var PortMapping map[string]string

func InitPhysicalInterfacesConfig() {
	PortMapping = make(map[string]string)
	portMappingResp, err := GetPortMapping()
	if err != nil {
		fmt.Printf("Failed to fetch port mapping : %s", err.Error())
	}

	for _, port := range portMappingResp.PortMap {
		PortMapping[port.Value] = port.PortName
	}

	// fmt.Println(PortMapping)

	resp, err := ListPhysicalInterfaces(model.EMPTY_STRING, "INIT")
	if err != nil {
		fmt.Println("Failed to fetch physical interfaces : ", err.Error())
	}
	if resp != nil && resp.PhysicalInterfacesCount > 0 {
		err = config.UpdateConfig(model.Config{PhysicalInterfaces: resp.PhysicalInterfaces}, model.DEVICE, model.EMPTY_STRING, "INIT PHYSICAL INTERFACES")
		if err != nil {
			fmt.Println("Failed to create physical interafaces in config file : ", err.Error())
		}
	}
}

func GetPortMapping() (resp *model.PortMappingResponse, err error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	var portMap []model.PortMap
	for _, link := range links {
		if link.Type() == model.DEVICE {
			if link.Attrs().Name == model.LOOP_BACK_DEVICE {
				continue
			}

			if !strings.HasPrefix(link.Attrs().Name, "e") {
				continue
			}

			Ipresp := utils.GetPrimaryIPAddressV2(link)
			isStaticIP := false

			if Ipresp.IpAddress != model.EMPTY_STRING && Ipresp.Netmask != 0 && Ipresp.Protocol == model.STATIC_STRING {
				isStaticIP = true
			}

			portMap = append(portMap, model.PortMap{
				PortName:   "PORT " + string(len(portMap)+65),
				Value:      link.Attrs().Name,
				IsStaticIp: isStaticIP,
			})
		}
	}

	return &model.PortMappingResponse{
		PortMap: portMap,
	}, nil
}

func addIp(link netlink.Link, ip string, netmask int) error {
	netmaskStr := fmt.Sprintf("%d", netmask)
	_, ipnet := utils.ValidateIPNetmask(ip, netmaskStr)
	addr := &netlink.Addr{
		IPNet: ipnet,
	}
	err := netlink.AddrAdd(link, addr)
	if err != nil {
		return fmt.Errorf("failed to add IP address to physical interface : %s", err.Error())
	}
	return nil
}

func RestoreInterfaceSettings(caller string) error {

	resp, err := config.FetchConfig(model.DEVICE)
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage(fmt.Sprintf("sysLog:Failed to fetch physical interfaces from config file : %s", err.Error())))
		return err
	}

	for _, intf := range resp.PhysicalInterfaces {
		link, err := netlink.LinkByName(intf.Name)
		if err != nil {
			// logger.LoggerFunc("error", logger.LoggerMessage(fmt.Sprintf("sysLog:Interface %s not found", PortMapping[intf.Name])))
			continue
		}

		if link.Attrs().Name == model.LOOP_BACK_DEVICE {
			continue
		}

		if !strings.HasPrefix(link.Attrs().Name, "e") {
			continue
		}

		if intf.IpAddress.Protocol != model.STATIC_STRING {
			continue
		}

		err = utils.PerformIpFlush(link.Attrs().Name)
		if err != nil {
			// logger.LoggerFunc("error", logger.LoggerMessage(fmt.Sprintf("sysLog:Failed to flush IP addresses from interface : %s", PortMapping[intf.Name])))
			return err
		}

		if intf.IpAddress.IpAddress != model.EMPTY_STRING && intf.IpAddress.Netmask != 0 {
			err = addIp(link, intf.IpAddress.IpAddress, intf.IpAddress.Netmask)
			if err != nil {
				logger.LoggerFunc("error", logger.LoggerMessage(fmt.Sprintf("sysLog:Failed to set IP address for interface : %s", PortMapping[intf.Name])))
				return err
			}
		}

		if len(intf.AliasList) > 0 {
			for _, alias := range intf.AliasList {
				if alias.IpAddress == model.EMPTY_STRING {
					continue
				}
				err = addIp(link, alias.IpAddress, alias.Netmask)
				if err != nil {
					logger.LoggerFunc("error", logger.LoggerMessage(fmt.Sprintf("sysLog:Failed to set IP address for interface : %s", PortMapping[intf.Name])))
					return err
				}
			}
		}

		err = netlink.LinkSetUp(link)
		if err != nil {
			fmt.Println("Failed to bring up interface : ", intf.Name)
		}
	}
	logger.LoggerFunc("info", logger.LoggerMessage("sysLog:Physical interfaces have been restored successfully"))
	return nil
}

func ResetToFactoryDefaultSettingsForInterfaces() error {
	cfg, err := config.FetchConfig("ALL")
	if err != nil {
		return err
	}

	cfg.PhysicalInterfaces = []model.ListPhysicalInterface{}
	cfg.BridgeInterfaces = []model.ListBridgeInterface{}
	cfg.BondInterfaces = []model.ListBondDetails{}
	cfg.VLANInterfaces = []model.ListVlanInterface{}
	cfg.Ha = model.Ha{}
	cfg.StaticRoutes = model.ListStaticRoutes{}
	cfg.VirtualWires = []model.ListVirtualWire{}
	cfg.BGP = model.ListBgp{}
	cfg.OSPF = model.ListOspf{}
	cfg.OSPFV3 = model.ListOspfv3{}

	err = config.UpdateConfig(cfg, "ALL", model.EMPTY_STRING, "FACTORY DEFAULT SETTINGS")

	intfAt3, err := fetch3rdPortIfExists()
	if err != nil {
		return err
	}

	links, err := netlink.LinkList()
	if err != nil {
		return err
	}

	for _, intf := range links {

		intfName := intf.Attrs().Name

		link, err := netlink.LinkByName(intfName)
		if err != nil {
			continue
		}

		if intfName == model.LOOP_BACK_DEVICE {
			continue
		}

		if !strings.HasPrefix(intfName, "e") {
			continue
		}

		err = utils.PerformIpFlush(intfName)
		if err != nil {
			return err
		}

		err = netlink.LinkSetDown(link)
		if err != nil {
			return err
		}

		if intfName == intfAt3 {
			err = interface_utils.SetIP(intfName, "10.0.0.1", 23)
			if err != nil {
				return err
			}
			err = netlink.LinkSetUp(link)
			if err != nil {
				return err
			}

			resp, err := ListPhysicalInterfaces(intfName, "FACTORY DEFAULT SETTINGS")
			if err != nil {
				return err
			}

			if len(resp.PhysicalInterfaces) > 0 {
				resp.PhysicalInterfaces[0].LinkStats = model.LinkStats{}
			}

			err = config.UpdateConfig(model.Config{PhysicalInterfaces: resp.PhysicalInterfaces}, model.DEVICE, intfName, "FACTORY DEFAULT SETTINGS")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func fetch3rdPortIfExists() (string, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return extras.EMPTY_STRING, fmt.Errorf("Failed to fetch physical interfaces : %s", err.Error())
	}
	intfCount := 1
	if len(links) < 1 {
		return extras.EMPTY_STRING, fmt.Errorf("No physical ports found")
	} else if len(links) < 3 {
		intfCount = 1
	} else {
		intfCount = 3
	}

	c := 0

	for _, link := range links {
		if link.Type() == model.DEVICE {
			if link.Attrs().Name == model.LOOP_BACK_DEVICE {
				continue
			}

			if !strings.HasPrefix(link.Attrs().Name, "e") {
				continue
			}

			c++
			if c == intfCount {
				return link.Attrs().Name, nil
			}
		}
	}
	return extras.EMPTY_STRING, fmt.Errorf("No physical ports found")
}

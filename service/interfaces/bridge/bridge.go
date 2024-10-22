package bridge

import (
	config "anti-apt-backend/config/interface_config"
	"anti-apt-backend/logger"
	model "anti-apt-backend/model/interface_model"
	utils "anti-apt-backend/util/interface_utils"
	validations "anti-apt-backend/validation/interface_validations"
	"fmt"
	"strings"

	"github.com/vishvananda/netlink"
)

func CreateBridgeInterface(req model.CreateBridgeRequest) (*model.ListBridgeInterfacesResponse, error) {

	err := validations.ValidateInterfaceName(strings.TrimSpace(req.BridgeInterfaceName), model.BRIDGE_STRING)
	if err != nil {
		return nil, err
	}

	request, ok := utils.TrimStringsInStruct(req).(model.CreateBridgeRequest)
	if !ok {
		return nil, fmt.Errorf("Failed to parse JSON request")
	}

	err = utils.CheckInterfaceExists(strings.TrimSpace(req.BridgeInterfaceName))
	if err == nil {
		return nil, fmt.Errorf("Bridge interface %s already exists", strings.TrimSpace(req.BridgeInterfaceName))
	}

	err = validations.ValidateCreateBridgeRequest(request)
	if err != nil {
		return nil, err
	}

	for _, memIntf := range request.MemberInterfaces {
		hasMaster, err := utils.CheckIfMemberInterfaceHasMaster(memIntf)
		if err != nil {
			return nil, err
		}
		if hasMaster {
			return nil, err
		}
	}

	bridgeLink := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: request.BridgeInterfaceName,
		},
	}

	if err := netlink.LinkAdd(bridgeLink); err != nil {
		fmt.Println("Error while creating bridge interface : ", err.Error())
		return nil, fmt.Errorf("Failed to create bridge interface")
	}

	for _, memIntf := range request.MemberInterfaces {
		memIntfName := strings.TrimSpace(memIntf)
		if memIntfName != model.EMPTY_STRING {
			memIntfLink, err := netlink.LinkByName(memIntfName)
			if err != nil {
				return nil, fmt.Errorf("member interface %s not found", memIntfName)
			}
			if err := netlink.LinkSetMaster(memIntfLink, bridgeLink); err != nil {
				return nil, fmt.Errorf("Failed to add member interface %s to bridge", memIntfName)
			}
		}
	}

	if request.Description != model.EMPTY_STRING && request.BridgeNameByUser != model.EMPTY_STRING {
		if err := netlink.LinkSetAlias(bridgeLink, request.BridgeNameByUser+" : "+request.Description); err != nil {
			return nil, fmt.Errorf("Failed to set alias name and description for bridge interface %s", request.BridgeInterfaceName)
		}
	} else if request.BridgeNameByUser != model.EMPTY_STRING {
		if err := netlink.LinkSetAlias(bridgeLink, request.BridgeNameByUser); err != nil {
			return nil, fmt.Errorf("Failed to set alias name for bridge interface %s", request.BridgeInterfaceName)
		}
	} else if request.Description != model.EMPTY_STRING {
		if err := netlink.LinkSetAlias(bridgeLink, request.Description); err != nil {
			return nil, fmt.Errorf("Failed to set description as alias name for bridge interface %s", request.BridgeInterfaceName)
		}
	}

	hwaddr := request.HardwareAddress
	if strings.TrimSpace(hwaddr) != model.EMPTY_STRING {
		hwAddr, err := utils.ValidateAndGetHardwareAddress(hwaddr)
		if err != nil {
			return nil, fmt.Errorf("Invalid hardware address: %s", err.Error())
		}
		err = netlink.LinkSetHardwareAddr(bridgeLink, hwAddr)
		if err != nil {
			return nil, fmt.Errorf("Failed to set hardware address for bridge interface %s", request.BridgeInterfaceName)
		}
	}

	if request.EnableRouting {
		fmt.Println("Got request for enable routing for bridge interface : ", request.BridgeInterfaceName)
	} else {
		fmt.Println("Got request for disable routing for bridge interface : ", request.BridgeInterfaceName)
	}

	if request.IPv4Details.IPv4 {
		if request.IPv4Details.IPv4AssignmentMode == model.STATIC_STRING {
			_, ipnet := utils.ValidateIPNetmask(request.IPv4Details.IPAddress, request.IPv4Details.Netmask)
			addr := &netlink.Addr{
				IPNet: ipnet,
			}

			if err := netlink.AddrAdd(bridgeLink, addr); err != nil {
				return nil, fmt.Errorf("Failed to add IP address to bridge interface %s", request.BridgeInterfaceName)
			}
		}
		if request.IPv4Details.IPv4AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got request for DHCP for bridge interface : ", request.BridgeInterfaceName)
		}
	}
	if request.IPv6Details.IPv6 {
		if request.IPv6Details.IPv6AssignmentMode == model.STATIC_STRING {
			_, ipnet := utils.ValidateIPNetmask(request.IPv6Details.IPAddress, request.IPv6Details.Prefix)
			addr := &netlink.Addr{
				IPNet: ipnet,
			}

			if err := netlink.AddrAdd(bridgeLink, addr); err != nil {
				return nil, fmt.Errorf("Failed to add IP address to bridge interface %s", request.BridgeInterfaceName)
			}
		}
	}

	if request.PermitArpBroadcast {
		fmt.Println("Got request for permit arp broadcast for bridge interface : ", request.BridgeInterfaceName)
	} else {
		fmt.Println("Got request for deny arp broadcast for bridge interface : ", request.BridgeInterfaceName)
	}

	if request.MTU != 0 {
		err = netlink.LinkSetMTU(bridgeLink, request.MTU)
		if err != nil {
			return nil, fmt.Errorf("Failed to set MTU for bridge interface %s", request.BridgeInterfaceName)
		}
	}

	if request.MssDetails.OverRideMSS {
		fmt.Println("Got request for override MSS for bridge interface : ", request.BridgeInterfaceName)
	}

	if request.StpDetails.TurnOnStp {
		fmt.Println("Got request to turn on STP for bridge interface : ", request.BridgeInterfaceName)
	} else {
		fmt.Println("Got request to turn off STP for bridge interface : ", request.BridgeInterfaceName)
	}

	err = UpdateBridgeConfig(request.BridgeInterfaceName, "CREATE BRIDGE")
	if err != nil {
		return nil, err
	}

	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Bridge interface %s has been created", request.BridgeInterfaceName)))

	err = config.UpdateConfigSpecificFields(model.BRIDGE_STRING, request.BridgeInterfaceName, model.ConfigSpecificFields{
		ServingLocation: request.ServingLocation,
		DomainName:      request.DomainName,
		NetworkZone:     request.NetworkZone,
	}, "CREATE BRIDGE (config specific fields)")

	if err != nil {
		fmt.Println("Failed to update config specific fields for bridge interface : ", err.Error())
	}

	listInterfaces, err := ListBridgeInterfaces(model.ListBridgeInterfaceRequest{BridgeInterfaceName: model.EMPTY_STRING})
	if err != nil {
		return nil, fmt.Errorf("Bridge has been created but failed to fetch the updated bridge interfaces : %s", err.Error())
	}

	return listInterfaces, nil
}

func UpdateBridgeInterface(bridgeInterfaceName string, req model.UpdateBridgeRequest) (*model.ListBridgeInterfacesResponse, error) {

	request, ok := utils.TrimStringsInStruct(req).(model.UpdateBridgeRequest)
	if !ok {
		return nil, fmt.Errorf("Failed to parse JSON request")
	}

	bridgeInterfaceName = strings.TrimSpace(bridgeInterfaceName)
	bridgeLink, err := netlink.LinkByName(bridgeInterfaceName)
	if err != nil {
		return nil, fmt.Errorf("Bridge interface %s not found", bridgeInterfaceName)
	}

	if bridgeLink.Type() != model.BRIDGE_STRING {
		return nil, fmt.Errorf("Interface %s is a type of %s, not a bridge interface", bridgeInterfaceName, bridgeLink.Type())
	}

	memberInterfaces := getMemberInterfaces(bridgeLink, []netlink.Link{})
	request.AddToBridge, request.RemoveFromBridge = utils.FetchAddRemoveFromList(memberInterfaces, request.MemberInterfaces)

	err = validations.ValidateUpdateBridgeRequest(request, bridgeInterfaceName)
	if err != nil {
		return nil, err
	}

	if len(memberInterfaces)+len(request.AddToBridge)-len(request.RemoveFromBridge) < 2 {
		return nil, fmt.Errorf("Atleast 2 interfaces should remain as members for a bridge interface")
	}

	for _, memIntfToAdd := range request.AddToBridge {
		if memIntfToAdd != model.EMPTY_STRING {
			addLink, err := netlink.LinkByName(memIntfToAdd)
			if err != nil {
				return nil, fmt.Errorf("Member interface to add %s not found", memIntfToAdd)
			}
			if err := netlink.LinkSetMaster(addLink, bridgeLink.(*netlink.Bridge)); err != nil {
				return nil, fmt.Errorf("Failed to add interface %s to bridge", memIntfToAdd)
			}
			logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Member interface %s has been added to bridge %s", memIntfToAdd, bridgeInterfaceName)))
		}
	}

	for _, memIntfToRemove := range request.RemoveFromBridge {
		if memIntfToRemove != model.EMPTY_STRING {
			removeLink, err := netlink.LinkByName(memIntfToRemove)
			if err != nil {
				return nil, fmt.Errorf("Member interface to remove %s not found", memIntfToRemove)
			}
			if err := netlink.LinkSetNoMaster(removeLink); err != nil {
				return nil, fmt.Errorf("Failed to remove interface %s from bridge %s", memIntfToRemove, bridgeInterfaceName)
			}
			logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Member interface %s has been removed from bridge %s", memIntfToRemove, bridgeInterfaceName)))
		}
	}

	hwaddr := request.HardwareAddress
	if strings.TrimSpace(hwaddr) != model.EMPTY_STRING && hwaddr != bridgeLink.Attrs().HardwareAddr.String() {
		hwAddr, err := utils.ValidateAndGetHardwareAddress(hwaddr)
		if err != nil {
			return nil, fmt.Errorf("Invalid hardware address: %s", err.Error())
		}
		err = netlink.LinkSetHardwareAddr(bridgeLink, hwAddr)
		if err != nil {
			return nil, fmt.Errorf("Failed to update hardware address for bridge interface %s", bridgeInterfaceName)
		}
	}

	if request.IPv4Details.IPv4 {
		if request.IPv4Details.IPv4AssignmentMode == model.STATIC_STRING {

			primaryIp := utils.GetPrimaryIPAddress(bridgeLink)

			if request.IPv4Details.IPAddress+"/"+request.IPv4Details.Netmask != primaryIp {
				err = utils.PerformIpDel(bridgeLink, primaryIp, req.IPv4Details.IPAddress, req.IPv4Details.Netmask)
				if err != nil {
					return nil, fmt.Errorf("Failed to delete IP address from bridge interface %s with error : %v", bridgeInterfaceName, err)
				}
				logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("IP address %s has been updated to %s on interface %s", primaryIp, req.IPv4Details.IPAddress, bridgeInterfaceName)))
			}
		} else if request.IPv4Details.IPv4AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got update request for DHCP V4 for bridge interface : ", bridgeInterfaceName)
		}
	} else if request.IPv6Details.IPv6 {
		if request.IPv6Details.IPv6AssignmentMode == model.STATIC_STRING {

			primaryIp := utils.GetPrimaryIPAddress(bridgeLink)

			if request.IPv6Details.IPAddress+"/"+request.IPv6Details.Prefix != primaryIp {
				err = utils.PerformIpDel(bridgeLink, primaryIp, req.IPv6Details.IPAddress, req.IPv6Details.Prefix)
				if err != nil {
					return nil, fmt.Errorf("Failed to delete IP address from bridge interface %s with error : %v", bridgeInterfaceName, err)
				}
				logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("IP address %s has been updated to %s on interface %s", primaryIp, req.IPv6Details.IPAddress, bridgeInterfaceName)))
			}
		} else if request.IPv6Details.IPv6AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got update request for DHCP V6 for bridge interface : ", bridgeInterfaceName)
		}
	}

	if request.MTU != 0 && (request.MTU != bridgeLink.Attrs().MTU) {
		err = netlink.LinkSetMTU(bridgeLink, request.MTU)
		if err != nil {
			return nil, fmt.Errorf("Failed to update MTU for bridge interface %s", bridgeInterfaceName)
		}
		logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("MTU has been updated to %d on interface %s", request.MTU, bridgeInterfaceName)))
	}

	configFields, err := config.FetchConfigSpecificFields(model.BRIDGE_STRING)
	if err != nil {
		fmt.Printf("Failed to fetch config specific fields from config file : %s", err.Error())
	}

	err = UpdateBridgeConfig(bridgeInterfaceName, "UPDATE BRIDGE")
	if err != nil {
		return nil, err
	}

	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Bridge interface %s has been updated", bridgeInterfaceName)))

	bridgeConfigFields, exists := configFields[bridgeInterfaceName]
	if !exists {
		bridgeConfigFields = model.ConfigSpecificFields{}
	}

	if request.ServingLocation != bridgeConfigFields.ServingLocation || request.DomainName != bridgeConfigFields.DomainName || request.NetworkZone != bridgeConfigFields.NetworkZone || request.IsDisabled != bridgeConfigFields.IsDisabled {
		err = config.UpdateConfigSpecificFields(model.BRIDGE_STRING, bridgeInterfaceName, model.ConfigSpecificFields{
			ServingLocation: request.ServingLocation,
			DomainName:      request.DomainName,
			NetworkZone:     request.NetworkZone,
			IsDisabled:      request.IsDisabled,
		}, "UPDATE BRIDGE (config specific fields)")
		if err != nil {
			return nil, fmt.Errorf("Failed to update config specific fields for bridge interface : %s", err.Error())
		}
	}

	listInterfaces, err := ListBridgeInterfaces(model.ListBridgeInterfaceRequest{BridgeInterfaceName: model.EMPTY_STRING})
	if err != nil {
		return nil, fmt.Errorf("Bridge details has been updated but failed to fetch the updated bridge interfaces : %s", err.Error())
	}

	return listInterfaces, nil
}

func UpdateBridgeConfig(bridgeName string, caller string) (err error) {
	var respForConfig *model.ListBridgeInterfacesResponse
	var configResp model.Config
	if config.CheckIfInterfaceTypeIsEmpty(model.BRIDGE_STRING) {
		respForConfig, err = ListBridgeInterfaces(model.ListBridgeInterfaceRequest{BridgeInterfaceName: model.EMPTY_STRING})
		if err != nil {
			return fmt.Errorf("Bridge has been created but failed to fetch the updated bridge interfaces : %s", err.Error())
		}
		configResp.BridgeInterfaces = respForConfig.BridgeInterfaces
	} else {
		respForConfig, err = ListBridgeInterfaces(model.ListBridgeInterfaceRequest{BridgeInterfaceName: bridgeName})
		if err != nil {
			return fmt.Errorf("Bridge has been created but failed to fetch the updated bridge interfaces : %s", err.Error())
		}
		if len(respForConfig.BridgeInterfaces) > 0 && respForConfig.BridgeInterfaces[0].BridgeInterfaceName == bridgeName {
			configResp.BridgeInterfaces = []model.ListBridgeInterface{respForConfig.BridgeInterfaces[0]}
		}
	}

	err = config.UpdateConfig(configResp, model.BRIDGE_STRING, bridgeName, caller)
	if err != nil {
		return fmt.Errorf("Bridge has been created but failed to update the config file : %s", err.Error())
	}
	return nil
}

func ListBridgeInterfaces(req model.ListBridgeInterfaceRequest) (resp *model.ListBridgeInterfacesResponse, err error) {

	bridgeName := req.BridgeInterfaceName

	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch interfaces")
	}

	if bridgeName != model.EMPTY_STRING && bridgeName != "all" {
		bridgeLink, err := netlink.LinkByName(bridgeName)
		if err != nil {
			return nil, fmt.Errorf("Bridge interface %s not found", bridgeName)
		}
		if bridgeLink.Type() != model.BRIDGE_STRING {
			return nil, fmt.Errorf("Interface %s is a type of %s, not a Vlan interface", bridgeName, bridgeLink.Type())
		}
	}

	bridgeDetails := []model.ListBridgeInterface{}

	configFields, err := config.FetchConfigSpecificFields(model.BRIDGE_STRING)
	if err != nil {
		fmt.Printf("Failed to fetch config specific fields from config file : %s", err.Error())
	}

	for _, link := range links {
		if link.Type() == model.BRIDGE_STRING && (bridgeName == model.EMPTY_STRING || bridgeName == "all" || link.Attrs().Name == bridgeName) {

			memberInterfaces := getMemberInterfaces(link, links)
			memberDetails := []model.ListMemberInterface{}
			for _, memberName := range memberInterfaces {
				memberLink, err := netlink.LinkByName(memberName)
				if err != nil {
					fmt.Println("Error while fetching Member interface : ", memberName, " Error : ", err.Error())
					return nil, fmt.Errorf("Error while fetching Member interface %s", memberName)
				}
				memberInfo := model.ListMemberInterface{
					InterfaceName:   memberName,
					IpAddress:       utils.GetPrimaryIPAddress(memberLink),
					IpProtocol:      utils.GetIPProtocol(memberLink),
					IpVersion:       utils.GetIPVersion(memberLink),
					MTU:             memberLink.Attrs().MTU,
					HardwareAddress: memberLink.Attrs().HardwareAddr.String(),
				}
				memberDetails = append(memberDetails, memberInfo)
			}

			aliasList := utils.GetSecondaryIPAddressListV2(link)

			bridgeConfigFields, exists := configFields[link.Attrs().Name]
			if !exists {
				bridgeConfigFields = model.ConfigSpecificFields{}
			}

			bridgeInfo := model.ListBridgeInterface{
				BridgeInterfaceName:     link.Attrs().Name,
				CountOfMemberInterfaces: len(memberInterfaces),
				IpAddress:               utils.GetPrimaryIPAddressV2(link),
				AliasList:               aliasList,
				IpProtocol:              utils.GetIPProtocol(link),
				IpVersion:               utils.GetIPVersion(link),
				MTU:                     link.Attrs().MTU,
				HardwareAddress:         link.Attrs().HardwareAddr.String(),
				IsEditable:              true,
				IsDeletable:             true,
				MemberInterfaces:        memberDetails,
				ServingLocation:         bridgeConfigFields.ServingLocation,
				DomainName:              bridgeConfigFields.DomainName,
				NetworkZone:             bridgeConfigFields.NetworkZone,
			}

			bridgeDetails = append(bridgeDetails, bridgeInfo)

			if bridgeName != model.EMPTY_STRING && bridgeName != "all" {
				break
			}
		}
	}

	bridgeCount := len(bridgeDetails)

	return &model.ListBridgeInterfacesResponse{
		BridgeCount:      bridgeCount,
		BridgeInterfaces: bridgeDetails,
	}, nil
}

func DeleteBridgeInterface(request model.ListBridgeInterfaceRequest) (*model.ListBridgeInterfacesResponse, error) {

	bridgeInterfaceName := request.BridgeInterfaceName

	if bridgeInterfaceName == model.EMPTY_STRING {
		return nil, fmt.Errorf("Bridge interface name cannot be empty")
	}

	bridgeLink, err := netlink.LinkByName(bridgeInterfaceName)
	if err != nil {
		return nil, fmt.Errorf("Bridge interface %s not found", bridgeInterfaceName)
	}

	if bridgeLink.Type() != model.BRIDGE_STRING {
		return nil, fmt.Errorf("Interface %s is a type of %s, not a Vlan interface", bridgeInterfaceName, bridgeLink.Type())
	}

	if err := netlink.LinkDel(bridgeLink); err != nil {
		return nil, fmt.Errorf("Failed to delete bridge interface %s", bridgeInterfaceName)
	}

	err = config.DeleteConfig(model.BRIDGE_STRING, bridgeInterfaceName, "DELETE BRIDGE")
	if err != nil {
		fmt.Printf("Failed to delete bridge interface %s from config : %s", bridgeInterfaceName, err.Error())
	}

	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Bridge interface %s has been deleted", bridgeInterfaceName)))

	listInterfaces, err := ListBridgeInterfaces(model.ListBridgeInterfaceRequest{BridgeInterfaceName: model.EMPTY_STRING})
	if err != nil {
		return nil, fmt.Errorf("Bridge has been deleted but failed to fetch the updated bridge interfaces : %s", err.Error())
	}

	return listInterfaces, nil
}

func getMemberInterfaces(bridge netlink.Link, links []netlink.Link) []string {
	MemberInterfaces := []string{}

	if len(links) == 0 {
		links, _ = netlink.LinkList()
	}

	for _, link := range links {
		if link.Type() != model.BRIDGE_STRING && link.Attrs().MasterIndex == bridge.Attrs().Index {
			MemberInterfaces = append(MemberInterfaces, link.Attrs().Name)
		}
	}

	return MemberInterfaces
}

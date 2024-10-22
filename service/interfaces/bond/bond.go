package bond

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

func CreateBondedLink(request model.CreateBondRequest) (resp *model.ListBondInterfacesResponse, err error) {

	err = validations.ValidateInterfaceName(strings.TrimSpace(request.BondInterfaceName), model.BOND_STRING)
	if err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(request.BondMode)) == 0 {
		return nil, fmt.Errorf("Bond mode cannot be empty")
	}

	request, ok := utils.TrimStringsInStruct(request).(model.CreateBondRequest)
	if !ok {
		return nil, fmt.Errorf("Failed to parse JSON request")
	}

	err = utils.CheckInterfaceExists(strings.TrimSpace(request.BondInterfaceName))
	if err == nil {
		return nil, fmt.Errorf("Bond interface %s already exists", strings.TrimSpace(request.BondInterfaceName))
	}

	err = validations.ValidateCreateBondRequest(request)
	if err != nil {
		return nil, err
	}

	for _, slaveIntf := range request.SlaveInterfaces {
		hasMaster, err := utils.CheckIfMemberInterfaceHasMaster(slaveIntf)
		if err != nil {
			return nil, err
		}
		if hasMaster {
			return nil, err
		}
	}

	bondLinkObj := netlink.NewLinkBond(netlink.NewLinkAttrs())

	linkAttrs := netlink.LinkAttrs{
		Name: request.BondInterfaceName,
	}

	bondLinkObj.LinkAttrs = linkAttrs
	bondLinkObj.Mode = netlink.StringToBondMode(request.BondMode)

	err = netlink.LinkAdd(bondLinkObj)
	if err != nil {
		fmt.Printf("Failed to create bond interface, error: %+v\n", err)
		return nil, err
	}
	defer func() {
		if err != nil {
			fmt.Printf("Failed to create bond interface, error: %+v\n", err)
			if err := netlink.LinkDel(bondLinkObj); err != nil {
				fmt.Printf("Failed to delete bond interface, error: %+v\n", err)
			}
		}
	}()

	for _, slave := range request.SlaveInterfaces {
		if strings.TrimSpace(slave) != model.EMPTY_STRING {
			slaveLink, err := netlink.LinkByName(slave)
			if err != nil {
				fmt.Printf("slave link %s not found, error: %+v\n", slave, err)
			}

			if err := netlink.LinkSetBondSlave(slaveLink, bondLinkObj); err != nil {
				return nil, err
			}
		}

	}

	bondLink, err := netlink.LinkByName(request.BondInterfaceName)
	if err != nil {
		fmt.Printf("Failed to get bond link, error: %+v\n", err)
	}

	if request.BondNameByUser != model.EMPTY_STRING {
		if err := netlink.LinkSetAlias(bondLink, request.BondNameByUser); err != nil {
			return nil, fmt.Errorf("Failed to set alias name for bond interface %s", request.BondInterfaceName)
		}
	}

	hwaddr := request.HardwareAddress
	if strings.TrimSpace(hwaddr) != model.EMPTY_STRING {
		hwAddr, err := utils.ValidateAndGetHardwareAddress(hwaddr)
		if err != nil {
			return nil, fmt.Errorf("Invalid hardware address: %s", err.Error())
		}
		err = netlink.LinkSetHardwareAddr(bondLink, hwAddr)
		if err != nil {
			return nil, fmt.Errorf("Failed to set hardware address for bond interface %s", request.BondInterfaceName)
		}
	}

	if request.IPv4Details.IPv4 {
		if request.IPv4Details.IPv4AssignmentMode == model.STATIC_STRING {
			_, ipnet := utils.ValidateIPNetmask(request.IPv4Details.IPAddress, request.IPv4Details.Netmask)
			addr := &netlink.Addr{
				IPNet: ipnet,
			}

			if err := netlink.AddrAdd(bondLink, addr); err != nil {
				return nil, fmt.Errorf("Failed to add IP address to bond interface %s", request.BondInterfaceName)
			}
		}
		if request.IPv4Details.IPv4AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got request for DHCP for bond interface : ", request.BondInterfaceName)
		}
	}
	if request.IPv6Details.IPv6 {
		if request.IPv6Details.IPv6AssignmentMode == model.STATIC_STRING {
			_, ipnet := utils.ValidateIPNetmask(request.IPv6Details.IPAddress, request.IPv6Details.Prefix)
			addr := &netlink.Addr{
				IPNet: ipnet,
			}

			if err := netlink.AddrAdd(bondLink, addr); err != nil {
				return nil, fmt.Errorf("Failed to add IP address to bond interface %s", request.BondInterfaceName)
			}
		}
	}

	if request.MTU != 0 {
		err = netlink.LinkSetMTU(bondLink, request.MTU)
		if err != nil {
			return nil, fmt.Errorf("Failed to set MTU for bond interface %s", request.BondInterfaceName)
		}
	}

	err = UpdateBondConfig(request.BondInterfaceName, "CREATE BOND")
	if err != nil {
		return nil, err
	}

	err = config.UpdateConfigSpecificFields(model.BOND_STRING, request.BondInterfaceName, model.ConfigSpecificFields{
		ServingLocation: request.ServingLocation,
		DomainName:      request.DomainName,
		NetworkZone:     request.NetworkZone,
	}, "CREATE BOND (config specific fields)")

	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("LAG interface %s created successfully", request.BondInterfaceName)))

	if err != nil {
		return nil, fmt.Errorf("Bond created but failed to update the config file : %s", err.Error())
	}

	bondInterfaces, err := ListBondInterfaces(model.ListBondInterfacesRequest{BondInterfaceName: model.EMPTY_STRING})
	if err != nil {
		return nil, fmt.Errorf("Failed to list bond interfaces")
	}

	err = nil

	return bondInterfaces, nil
}

func UpdateBondInterface(bondInterfaceName string, request model.UpdateBondRequest) (resp *model.ListBondInterfacesResponse, err error) {
	req, ok := utils.TrimStringsInStruct(request).(model.UpdateBondRequest)
	if !ok {
		return nil, fmt.Errorf("Failed to parse JSON request")
	}

	bondInterfaceName = strings.TrimSpace(bondInterfaceName)
	bondLink, err := netlink.LinkByName(bondInterfaceName)
	if err != nil {
		return nil, fmt.Errorf("Bond interface %s not found", bondInterfaceName)
	}

	if bondLink.Type() != model.BOND_STRING {
		return nil, fmt.Errorf("Interface %s is a type of %s, not a bond interface", bondInterfaceName, bondLink.Type())
	}

	slaveInterfaces := getSlaveInterfaces(bondLink, []netlink.Link{})
	req.AddToBond, req.RemoveFromBond = utils.FetchAddRemoveFromList(slaveInterfaces, req.SlaveInterfaces)

	err = validations.ValidateUpdateBondRequest(req, bondInterfaceName)
	if err != nil {
		return nil, err
	}

	if len(slaveInterfaces)+len(req.AddToBond)-len(req.RemoveFromBond) < 2 {
		return nil, fmt.Errorf("Atleast 2 interfaces should remain as slaves for a bond interface")
	}

	for _, slaveIntfToAdd := range req.AddToBond {
		if slaveIntfToAdd != model.EMPTY_STRING {
			bondLinkObj := bondLink.(*netlink.Bond)
			slaveLink, err := netlink.LinkByName(slaveIntfToAdd)
			if err != nil {
				fmt.Printf("slave link %s not found, error: %+v\n", slaveIntfToAdd, err)
			}
			if err := netlink.LinkSetBondSlave(slaveLink, bondLinkObj); err != nil {
				return nil, err
			}
			logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Slave interface %s has been added to LAG %s", slaveIntfToAdd, bondInterfaceName)))
		}
	}

	for _, slaveIntfToRemove := range req.RemoveFromBond {
		if slaveIntfToRemove != model.EMPTY_STRING {
			removeLink, err := netlink.LinkByName(slaveIntfToRemove)
			if err != nil {
				return nil, fmt.Errorf("Slave interface to remove %s not found", slaveIntfToRemove)
			}
			if err := netlink.LinkSetNoMaster(removeLink); err != nil {
				return nil, fmt.Errorf("Failed to remove interface %s from bond %s", slaveIntfToRemove, bondInterfaceName)
			}
			logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Slave interface %s has been removed from LAG %s", slaveIntfToRemove, bondInterfaceName)))
		}
	}

	hwaddr := req.HardwareAddress
	if strings.TrimSpace(hwaddr) != model.EMPTY_STRING && hwaddr != bondLink.Attrs().HardwareAddr.String() {
		hwAddr, err := utils.ValidateAndGetHardwareAddress(hwaddr)
		if err != nil {
			return nil, fmt.Errorf("Invalid hardware address: %s", err.Error())
		}
		err = netlink.LinkSetHardwareAddr(bondLink, hwAddr)
		if err != nil {
			return nil, fmt.Errorf("Failed to update hardware address for bond interface %s", bondInterfaceName)
		}
	}

	if request.IPv4Details.IPv4 {
		if request.IPv4Details.IPv4AssignmentMode == model.STATIC_STRING {

			primaryIp := utils.GetPrimaryIPAddress(bondLink)

			if request.IPv4Details.IPAddress+"/"+request.IPv4Details.Netmask != primaryIp {
				err = utils.PerformIpDel(bondLink, primaryIp, req.IPv4Details.IPAddress, req.IPv4Details.Netmask)
				if err != nil {
					return nil, fmt.Errorf("Failed to delete IP address from bond interface %s with error : %v", bondInterfaceName, err)
				}
				logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("IP address %s has been updated to %s on LAG %s", primaryIp, req.IPv4Details.IPAddress, bondInterfaceName)))
			}
		} else if request.IPv4Details.IPv4AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got update request for DHCP V4 for bond interface : ", bondInterfaceName)
		}
	} else if request.IPv6Details.IPv6 {
		if request.IPv6Details.IPv6AssignmentMode == model.STATIC_STRING {

			primaryIp := utils.GetPrimaryIPAddress(bondLink)

			if request.IPv6Details.IPAddress+"/"+request.IPv6Details.Prefix != primaryIp {
				err = utils.PerformIpDel(bondLink, primaryIp, req.IPv6Details.IPAddress, req.IPv6Details.Prefix)
				if err != nil {
					return nil, fmt.Errorf("Failed to delete IP address from bond interface %s with error : %v", bondInterfaceName, err)
				}
				logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("IP address %s has been updated to %s on LAG %s", primaryIp, req.IPv6Details.IPAddress, bondInterfaceName)))
			}
		} else if request.IPv6Details.IPv6AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got update request for DHCP V6 for bond interface : ", bondInterfaceName)
		}
	}

	if req.MTU != 0 && req.MTU != bondLink.Attrs().MTU {
		err = netlink.LinkSetMTU(bondLink, req.MTU)
		if err != nil {
			return nil, fmt.Errorf("Failed to update MTU for bond interface %s", bondInterfaceName)
		}
		logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("MTU has been updated to %d on LAG %s", req.MTU, bondInterfaceName)))
	}

	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("LAG interface %s has been updated", bondInterfaceName)))

	configFields, err := config.FetchConfigSpecificFields(model.VLAN_STRING)
	if err != nil {
		fmt.Printf("Failed to fetch config specific fields from config file : %s", err.Error())
	}

	err = UpdateBondConfig(bondInterfaceName, "UPDATE BOND")
	if err != nil {
		return nil, err
	}

	bondConfigFields, exists := configFields[bondInterfaceName]
	if !exists {
		bondConfigFields = model.ConfigSpecificFields{}
	}

	if req.ServingLocation != bondConfigFields.ServingLocation || req.DomainName != bondConfigFields.DomainName || req.NetworkZone != bondConfigFields.NetworkZone || req.IsDisabled != bondConfigFields.IsDisabled {
		err = config.UpdateConfigSpecificFields(model.BOND_STRING, bondInterfaceName, model.ConfigSpecificFields{
			ServingLocation: req.ServingLocation,
			DomainName:      req.DomainName,
			NetworkZone:     req.NetworkZone,
			IsDisabled:      req.IsDisabled,
		}, "UPDATE BOND (config specific fields)")
		if err != nil {
			return nil, fmt.Errorf("Failed to update config specific fields for bond interface : %s", err.Error())
		}
	}

	bondInterfaces, err := ListBondInterfaces(model.ListBondInterfacesRequest{BondInterfaceName: model.EMPTY_STRING})
	if err != nil {
		return nil, fmt.Errorf("Bond updated but failed to list bond interfaces")
	}

	return bondInterfaces, nil
}

func UpdateBondConfig(bondName string, caller string) (err error) {
	var respForConfig *model.ListBondInterfacesResponse
	var configResp model.Config
	if config.CheckIfInterfaceTypeIsEmpty(model.BOND_STRING) {
		respForConfig, err = ListBondInterfaces(model.ListBondInterfacesRequest{BondInterfaceName: model.EMPTY_STRING})
		if err != nil {
			return fmt.Errorf("Bond has been updated but failed to fetch the updated bond interfaces : %s", err.Error())
		}
		configResp.BondInterfaces = respForConfig.BondInterfaces
	} else {
		respForConfig, err = ListBondInterfaces(model.ListBondInterfacesRequest{BondInterfaceName: bondName})
		if err != nil {
			return fmt.Errorf("Bond has been updated but failed to fetch the updated bond interfaces : %s", err.Error())
		}
		if len(respForConfig.BondInterfaces) > 0 && respForConfig.BondInterfaces[0].BondInterfaceName == bondName {
			configResp.BondInterfaces = []model.ListBondDetails{respForConfig.BondInterfaces[0]}
		}
	}

	err = config.UpdateConfig(configResp, model.BOND_STRING, bondName, caller)
	if err != nil {
		return fmt.Errorf("Bond has been updated but failed to update the config file : %s", err.Error())
	}
	return nil
}

func getSlaveInterfaces(bond netlink.Link, links []netlink.Link) []string {
	slaveInterfaces := []string{}

	if len(links) == 0 {
		links, _ = netlink.LinkList()
	}

	for _, link := range links {
		if link.Attrs().Slave != nil && link.Attrs().Slave.SlaveType() == model.BOND_STRING && link.Attrs().MasterIndex == bond.Attrs().Index {
			slaveInterfaces = append(slaveInterfaces, link.Attrs().Name)
		}
	}

	return slaveInterfaces
}

func ListBondInterfaces(req model.ListBondInterfacesRequest) (resp *model.ListBondInterfacesResponse, err error) {

	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch interfaces")
	}

	bondInterfaceName := strings.TrimSpace(req.BondInterfaceName)

	if bondInterfaceName != model.EMPTY_STRING && bondInterfaceName != "all" {
		bondLink, err := netlink.LinkByName(bondInterfaceName)
		if err != nil {
			return nil, fmt.Errorf("Bond interface %s not found", bondInterfaceName)
		}
		if bondLink.Type() != model.BOND_STRING {
			return nil, fmt.Errorf("Interface %s is a type of %s, not a bond interface", bondInterfaceName, bondLink.Type())
		}
	}

	bondInterfaces := []model.ListBondDetails{}

	configFields, err := config.FetchConfigSpecificFields(model.BOND_STRING)
	if err != nil {
		fmt.Printf("Failed to fetch config specific fields from config file : %s", err.Error())
	}

	for _, link := range links {
		if link.Type() == model.BOND_STRING && (bondInterfaceName == model.EMPTY_STRING || bondInterfaceName == "all" || link.Attrs().Name == bondInterfaceName) {

			slaveInterfaces := getSlaveInterfaces(link, links)
			slaveDetails := []model.SlaveDetails{}
			for _, slaveName := range slaveInterfaces {
				slaveLink, err := netlink.LinkByName(slaveName)
				if err != nil {
					fmt.Println("Error while fetching slave interface : ", slaveName, " Error : ", err.Error())
					return nil, fmt.Errorf("Error while fetching slave interface %s", slaveName)
				}
				slaveInfo := model.SlaveDetails{
					InterfaceName:   slaveName,
					IpAddress:       utils.GetPrimaryIPAddress(slaveLink),
					IpProtocol:      utils.GetIPProtocol(slaveLink),
					IpVersion:       utils.GetIPVersion(slaveLink),
					MTU:             slaveLink.Attrs().MTU,
					HardwareAddress: slaveLink.Attrs().HardwareAddr.String(),
				}
				if slaveLink.Attrs().Slave != nil {
					slaveState := fmt.Sprintf(slaveLink.Attrs().Slave.(*netlink.BondSlave).State.String())
					slaveInfo.SlaveState = slaveState
				}

				slaveDetails = append(slaveDetails, slaveInfo)
			}

			aliasList := utils.GetSecondaryIPAddressListV2(link)

			bondConfigFields, exists := configFields[link.Attrs().Name]
			if !exists {
				bondConfigFields = model.ConfigSpecificFields{}
			}

			bondInfo := model.ListBondDetails{
				BondInterfaceName:      link.Attrs().Name,
				BondMode:               link.(*netlink.Bond).Mode.String(),
				IpAddress:              utils.GetPrimaryIPAddressV2(link),
				AliasList:              aliasList,
				IpProtocol:             utils.GetIPProtocol(link),
				IpVersion:              utils.GetIPVersion(link),
				MTU:                    link.Attrs().MTU,
				HardwareAddress:        link.Attrs().HardwareAddr.String(),
				CountOfSlaveInterfaces: len(slaveInterfaces),
				SlaveInterfaces:        slaveDetails,
				IsEditable:             true,
				IsDeletable:            true,
				ServingLocation:        bondConfigFields.ServingLocation,
				DomainName:             bondConfigFields.DomainName,
				NetworkZone:            bondConfigFields.NetworkZone,
			}

			bondInterfaces = append(bondInterfaces, bondInfo)

			if bondInterfaceName != model.EMPTY_STRING && bondInterfaceName != "all" {
				break
			}
		}
	}

	return &model.ListBondInterfacesResponse{
		BondCount:      len(bondInterfaces),
		BondInterfaces: bondInterfaces,
	}, nil
}

func DeleteBondInterface(req model.ListBondInterfacesRequest) (resp *model.ListBondInterfacesResponse, err error) {

	bondInterfaceName := strings.TrimSpace(req.BondInterfaceName)

	if bondInterfaceName == model.EMPTY_STRING {
		return nil, fmt.Errorf("Bond interface name cannot be empty")
	}

	bondLink, err := netlink.LinkByName(bondInterfaceName)
	if err != nil {
		return nil, fmt.Errorf("Bond interface %s not found", bondInterfaceName)
	}

	if bondLink.Type() != model.BOND_STRING {
		return nil, fmt.Errorf("Interface %s is a type of %s, not a bond interface", bondInterfaceName, bondLink.Type())
	}

	if err := netlink.LinkDel(bondLink); err != nil {
		return nil, fmt.Errorf("Failed to delete bond interface %s", bondInterfaceName)
	}

	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("LAG interface %s has been deleted", bondInterfaceName)))

	err = config.DeleteConfig(model.BOND_STRING, bondInterfaceName, "DELETE BOND")
	if err != nil {
		fmt.Printf("Failed to delete bond interface %s from config : %s", bondInterfaceName, err.Error())
	}

	bondInterfaces, err := ListBondInterfaces(model.ListBondInterfacesRequest{BondInterfaceName: model.EMPTY_STRING})
	if err != nil {
		return nil, fmt.Errorf("Failed to list bond interfaces")
	}

	return bondInterfaces, nil
}

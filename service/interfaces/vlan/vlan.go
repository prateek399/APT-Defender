package vlan

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

func CreateVlanInterface(req model.CreateVlanRequest) (*model.ListVlanInterfacesResp, error) {

	err := validations.ValidateInterfaceName(strings.TrimSpace(req.VlanInterfaceName), model.VLAN_STRING)
	if err != nil {
		return nil, err
	}

	request, ok := utils.TrimStringsInStruct(req).(model.CreateVlanRequest)
	if !ok {
		return nil, fmt.Errorf("Failed to parse JSON request")
	}

	err = validations.ValidateCreateVlanRequest(request)
	if err != nil {
		return nil, err
	}

	parentLink, err := netlink.LinkByName(request.ParentInterface)
	if err != nil {
		return nil, fmt.Errorf("Parent interface %s not found", request.ParentInterface)
	}

	vlanLink := &netlink.Vlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:        request.VlanInterfaceName,
			ParentIndex: parentLink.Attrs().Index,
		},
		VlanId: request.VlanID,
	}

	if err := netlink.LinkAdd(vlanLink); err != nil {
		if err.Error() == "file exists" {
			return nil, fmt.Errorf("A Vlan interface with VlanId - %d already exists with parent interface - %s", request.VlanID, request.ParentInterface)
		}
		return nil, fmt.Errorf("Failed to create Vlan interface: %s", err.Error())
	}

	if request.IPv4Details.IPv4 {
		if request.IPv4Details.IPv4AssignmentMode == model.STATIC_STRING {
			_, ipnet := utils.ValidateIPNetmask(request.IPv4Details.IPAddress, request.IPv4Details.Netmask)
			addr := &netlink.Addr{
				IPNet: ipnet,
			}

			if err := netlink.AddrAdd(vlanLink, addr); err != nil {
				return nil, fmt.Errorf("Failed to add IP address to vlan interface %s", request.VlanInterfaceName)
			}
		} else if request.IPv4Details.IPv4AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got request for DHCP for vlan interface : ", request.VlanInterfaceName)
		} else {
			return nil, fmt.Errorf("Invalid IPv4 assignment mode: %s for vlan interface %s", request.IPv4Details.IPv4AssignmentMode, request.VlanInterfaceName)
		}
	}
	if request.IPv6Details.IPv6 {
		if request.IPv6Details.IPv6AssignmentMode == model.STATIC_STRING {
			_, ipnet := utils.ValidateIPNetmask(request.IPv6Details.IPAddress, request.IPv6Details.Prefix)
			addr := &netlink.Addr{
				IPNet: ipnet,
			}

			if err := netlink.AddrAdd(vlanLink, addr); err != nil {
				return nil, fmt.Errorf("Failed to add IP address to vlan interface %s", request.VlanInterfaceName)
			}
		} else if request.IPv6Details.IPv6AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got request for DHCP for vlan interface : ", request.VlanInterfaceName)
		} else {
			return nil, fmt.Errorf("Invalid IPv6 assignment mode: %s for vlan interface %s", request.IPv6Details.IPv6AssignmentMode, request.VlanInterfaceName)
		}
	}

	err = UpdateVlanConfig(request.VlanInterfaceName, "CREATE VLAN")
	if err != nil {
		return nil, err
	}

	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Vlan interface %s has been created", request.VlanInterfaceName)))

	err = config.UpdateConfigSpecificFields(model.VLAN_STRING, request.VlanInterfaceName, model.ConfigSpecificFields{
		ServingLocation: request.ServingLocation,
		DomainName:      request.DomainName,
		NetworkZone:     request.NetworkZone,
	}, "CREATE VLAN (config specific fields)")

	if err != nil {
		fmt.Println("Failed to update config specific fields for vlan interface : ", err.Error())
	}

	listVlanInterfaces, err := ListVlanInterfaces(model.ListVlanInterfacesRequest{VlanInterfaceName: model.EMPTY_STRING})
	if err != nil {
		return nil, fmt.Errorf("Vlan has been created succesfully but failed to list VLAN interfaces : %s", err.Error())
	}

	return listVlanInterfaces, nil
}

func ListVlanInterfaces(req model.ListVlanInterfacesRequest) (*model.ListVlanInterfacesResp, error) {

	vlanName := req.VlanInterfaceName

	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch interfaces")
	}

	if vlanName != model.EMPTY_STRING && vlanName != "all" {
		vlanLink, err := netlink.LinkByName(vlanName)
		if err != nil {
			return nil, fmt.Errorf("Vlan interface %s not found", vlanName)
		}
		if vlanLink.Type() != model.VLAN_STRING {
			return nil, fmt.Errorf("Interface %s is a type of %s, not a Vlan interface", vlanName, vlanLink.Type())
		}
	}

	vlanDetails := []model.ListVlanInterface{}

	configFields, err := config.FetchConfigSpecificFields(model.VLAN_STRING)
	if err != nil {
		fmt.Printf("Failed to fetch config specific fields from config file : %s", err.Error())
	}

	for _, link := range links {
		if link.Type() == model.VLAN_STRING && (vlanName == model.EMPTY_STRING || vlanName == "all" || link.Attrs().Name == vlanName) {
			parentIndex := link.Attrs().ParentIndex
			parentLink := FetchLinkFromIndex(links, parentIndex)
			if parentLink == nil {
				continue // add in no parent interface found
			}
			parentName := parentLink.Attrs().Name
			parentInfo := model.ListMemberInterface{
				InterfaceName:   parentName,
				HardwareAddress: parentLink.Attrs().HardwareAddr.String(),
				IpAddress:       utils.GetPrimaryIPAddress(parentLink),
				IpProtocol:      utils.GetIPProtocol(parentLink),
				IpVersion:       utils.GetIPVersion(parentLink),
				MTU:             parentLink.Attrs().MTU,
			}

			aliasList := utils.GetSecondaryIPAddressListV2(link)

			vlanConfigFields, exists := configFields[link.Attrs().Name]
			if !exists {
				vlanConfigFields = model.ConfigSpecificFields{}
			}

			vlanInterface := model.ListVlanInterface{
				VlanInterfaceName: link.Attrs().Name,
				HardwareAddress:   link.Attrs().HardwareAddr.String(),
				IpAddress:         utils.GetPrimaryIPAddressV2(link),
				AliasList:         aliasList,
				IpProtocol:        utils.GetIPProtocol(link),
				IpVersion:         utils.GetIPVersion(link),
				MTU:               link.Attrs().MTU,
				ParentInterface:   parentInfo,
				IsEditable:        true,
				IsDeletable:       true,
				ServingLocation:   vlanConfigFields.ServingLocation,
				DomainName:        vlanConfigFields.DomainName,
				NetworkZone:       vlanConfigFields.NetworkZone,
			}
			vlanDetails = append(vlanDetails, vlanInterface)

			if vlanName != model.EMPTY_STRING && vlanName != "all" {
				break
			}
		}
	}

	return &model.ListVlanInterfacesResp{
		VlanInterfacesCount: len(vlanDetails),
		VlanInterfaces:      vlanDetails,
	}, nil

}

func FetchLinkFromIndex(links []netlink.Link, index int) netlink.Link {
	for _, link := range links {
		if link.Attrs().Index == index {
			return link
		}
	}
	return nil
}

func UpdateVlanInterface(vlanName string, req model.UpdateVlanRequest) (*model.ListVlanInterfacesResp, error) {

	if vlanName == model.EMPTY_STRING {
		return nil, fmt.Errorf("Vlan interface name cannot be empty")
	}

	request, ok := utils.TrimStringsInStruct(req).(model.UpdateVlanRequest)
	if !ok {
		return nil, fmt.Errorf("Failed to parse JSON request")
	}

	vlanLink, err := netlink.LinkByName(vlanName)
	if err != nil {
		return nil, fmt.Errorf("Vlan interface %s not found", vlanName)
	}

	err = validations.ValidateUpdateVlanRequest(request, vlanLink)
	if err != nil {
		return nil, err
	}

	if vlanLink.Type() != model.VLAN_STRING {
		return nil, fmt.Errorf("Interface %s is a type of %s, not a Vlan interface", vlanName, vlanLink.Type())
	}

	if request.IPv4Details.IPv4 {
		if request.IPv4Details.IPv4AssignmentMode == model.STATIC_STRING {

			primaryIp := utils.GetPrimaryIPAddress(vlanLink)

			if request.IPv4Details.IPAddress+"/"+request.IPv4Details.Netmask != primaryIp {
				err = utils.PerformIpDel(vlanLink, primaryIp, req.IPv4Details.IPAddress, req.IPv4Details.Netmask)
				if err != nil {
					return nil, fmt.Errorf("Failed to delete IP address from vlan interface %s with error : %v", vlanName, err)
				}
				logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("IP address %s has been updated to %s on vlan interface %s", primaryIp, req.IPv4Details.IPAddress, vlanName)))
			}
		} else if request.IPv4Details.IPv4AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got update request for DHCP V4 for vlan interface : ", vlanName)
		}
	} else if request.IPv6Details.IPv6 {
		if request.IPv6Details.IPv6AssignmentMode == model.STATIC_STRING {

			primaryIp := utils.GetPrimaryIPAddress(vlanLink)

			if request.IPv6Details.IPAddress+"/"+request.IPv6Details.Prefix != primaryIp {
				err = utils.PerformIpDel(vlanLink, primaryIp, req.IPv6Details.IPAddress, req.IPv6Details.Prefix)
				if err != nil {
					return nil, fmt.Errorf("Failed to delete IP address from vlan interface %s with error : %v", vlanName, err)
				}
				logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("IP address %s has been updated to %s on vlan interface %s", primaryIp, req.IPv6Details.IPAddress, vlanName)))
			}
		} else if request.IPv6Details.IPv6AssignmentMode == model.DHCP_STRING {
			fmt.Println("Got update request for DHCP V6 for vlan interface : ", vlanName)
		}
	}

	configFields, err := config.FetchConfigSpecificFields(model.VLAN_STRING)
	if err != nil {
		fmt.Printf("Failed to fetch config specific fields from config file : %s", err.Error())
	}

	err = UpdateVlanConfig(vlanName, "UPDATE VLAN")
	if err != nil {
		return nil, err
	}

	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Vlan interface %s has been updated", vlanName)))

	vlanConfigFields, exists := configFields[vlanName]
	if !exists {
		vlanConfigFields = model.ConfigSpecificFields{}
	}

	if request.ServingLocation != vlanConfigFields.ServingLocation || request.DomainName != vlanConfigFields.DomainName || request.NetworkZone != vlanConfigFields.NetworkZone || request.IsDisabled != vlanConfigFields.IsDisabled {
		err = config.UpdateConfigSpecificFields(model.VLAN_STRING, vlanName, model.ConfigSpecificFields{
			ServingLocation: request.ServingLocation,
			DomainName:      request.DomainName,
			NetworkZone:     request.NetworkZone,
			IsDisabled:      request.IsDisabled,
		}, "UPDATE VLAN (config specific fields)")
		if err != nil {
			return nil, fmt.Errorf("Failed to update config specific fields for vlan interface : %s", err.Error())
		}
	}

	listVlanInterfaces, err := ListVlanInterfaces(model.ListVlanInterfacesRequest{VlanInterfaceName: model.EMPTY_STRING})
	if err != nil {
		return nil, fmt.Errorf("Vlan interface %s has been updated succesfully but failed to list VLAN interfaces : %s", vlanName, err.Error())
	}
	return listVlanInterfaces, nil
}

func UpdateVlanConfig(vlanName string, caller string) (err error) {
	var respForConfig *model.ListVlanInterfacesResp
	var configResp model.Config
	if config.CheckIfInterfaceTypeIsEmpty(model.VLAN_STRING) {
		respForConfig, err = ListVlanInterfaces(model.ListVlanInterfacesRequest{VlanInterfaceName: model.EMPTY_STRING})
		if err != nil {
			return fmt.Errorf("Vlan has been created but failed to fetch the updated vlan interfaces : %s", err.Error())
		}
		configResp.VLANInterfaces = respForConfig.VlanInterfaces
	} else {
		respForConfig, err = ListVlanInterfaces(model.ListVlanInterfacesRequest{VlanInterfaceName: vlanName})
		if err != nil {
			return fmt.Errorf("Vlan has been created but failed to fetch the updated vlan interfaces : %s", err.Error())
		}
		if len(respForConfig.VlanInterfaces) > 0 && respForConfig.VlanInterfaces[0].VlanInterfaceName == vlanName {
			configResp.VLANInterfaces = []model.ListVlanInterface{respForConfig.VlanInterfaces[0]}
		}
	}

	err = config.UpdateConfig(configResp, model.VLAN_STRING, vlanName, caller)
	if err != nil {
		return fmt.Errorf("Vlan has been created but failed to update the config file : %s", err.Error())
	}
	return nil
}

func DeleteVlanInterface(req model.ListVlanInterfacesRequest) (*model.ListVlanInterfacesResp, error) {

	vlanName := req.VlanInterfaceName

	if vlanName == model.EMPTY_STRING {
		return nil, fmt.Errorf("Vlan interface name cannot be empty")
	}

	vlanLink, err := netlink.LinkByName(vlanName)
	if err != nil {
		return nil, fmt.Errorf("Vlan interface %s not found", vlanName)
	}

	if vlanLink.Type() != model.VLAN_STRING {
		return nil, fmt.Errorf("Interface %s is a type of %s, not a Vlan interface", vlanName, vlanLink.Type())
	}

	if err := netlink.LinkDel(vlanLink); err != nil {
		return nil, fmt.Errorf("Failed to delete Vlan interface %s", vlanName)
	}

	err = config.DeleteConfig(model.VLAN_STRING, vlanName, "DELETE VLAN")
	if err != nil {
		fmt.Printf("Failed to delete vlan interface %s from config : %s", vlanName, err.Error())
	}

	logger.LoggerFunc("info", logger.LoggerMessage(fmt.Sprintf("Vlan interface %s has been deleted", vlanName)))

	listVlanInterfaces, err := ListVlanInterfaces(model.ListVlanInterfacesRequest{VlanInterfaceName: model.EMPTY_STRING})
	if err != nil {
		return nil, fmt.Errorf("Vlan interface %s has been deleted succesfully but failed to list VLAN interfaces : %s", vlanName, err.Error())
	}
	return listVlanInterfaces, nil
}

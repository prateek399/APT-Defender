package interface_validations

import (
	model "anti-apt-backend/model/interface_model"
	utils "anti-apt-backend/util/interface_utils"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
)

const (
	maxCharacters = 15
)

func ValidateInterfaceName(name string, prefix string) error {

	if name == model.EMPTY_STRING {
		return fmt.Errorf("Interface name is required")
	}

	if len(name) > maxCharacters {
		return fmt.Errorf("Interface name %s exceeds the maximum allowed characters %d", name, maxCharacters)
	}

	if prefix == model.BRIDGE_STRING {
		prefix = "BR_"
	}

	if prefix == model.BOND_STRING {
		prefix = "LAG_"
	}

	if prefix == model.VLAN_STRING {
		prefix = "VLAN_"
	}

	if len(name) < len(prefix) || name[:len(prefix)] != prefix {
		return fmt.Errorf("Interface name %s does not start with the specified prefix %s", name, prefix)
	}

	validPattern := regexp.MustCompile("^[A-Za-z0-9_]*$")
	if !validPattern.MatchString(name[len(prefix):]) {
		return fmt.Errorf("Interface name %s contains invalid characters, only [A-Za-z0-9_] are allowed", name)
	}

	return nil
}

func ValidateUpdatePhysicalInterfaceRequest(request model.UpdatePhysicalInterfaceRequest, intfName string) error {

	err := validateHardwareAddress(request.HardwareAddress)
	if err != nil {
		return err
	}

	_, err = netlink.LinkByName(intfName)
	if err != nil {
		return fmt.Errorf("Physical interface %s not found", intfName)
	}

	err = validateIPv4Details(request.IPv4Details)
	if err != nil {
		return err
	}

	err = validateIPv6Details(request.IPv6Details)
	if err != nil {
		return err
	}

	err = validateMTUDetails(request.MTU)
	if err != nil {
		return err
	}

	err = isValidNetworkZone(request.NetworkZone)
	if err != nil {
		return err
	}

	return nil
}

// bridge validations
func ValidateCreateBridgeRequest(request model.CreateBridgeRequest) error {

	if len(request.MemberInterfaces) < 2 {
		return fmt.Errorf("Atleast 2 member interfaces are required to create a bridge")
	}

	hasCommom, commonInterfaceName := utils.HasCommon(request.MemberInterfaces)
	if hasCommom {
		return fmt.Errorf("Cannot add the same interface %s to the bridge multiple times", commonInterfaceName)
	}

	for _, memIntf := range request.MemberInterfaces {
		memIntfName := strings.TrimSpace(memIntf)
		if memIntfName != model.EMPTY_STRING {
			memLink, err := netlink.LinkByName(memIntfName)
			if err != nil {
				return fmt.Errorf("Member interface %s not found", memIntfName)
			}
			if memLink.Type() != model.DEVICE {
				return fmt.Errorf("Member interface %s is not a physical interface", memIntfName)
			}
		}
	}

	err := validateHardwareAddress(request.HardwareAddress)
	if err != nil {
		return err
	}

	err = validateIPv4Details(request.IPv4Details)
	if err != nil {
		return err
	}

	err = validateIPv6Details(request.IPv6Details)
	if err != nil {
		return err
	}

	err = validateMTUDetails(request.MTU)
	if err != nil {
		return err
	}

	err = validateMSSDetails(request.MssDetails)
	if err != nil {
		return err
	}

	err = validateStpDetails(request.StpDetails)
	if err != nil {
		return err
	}

	err = isValidNetworkZone(request.NetworkZone)
	if err != nil {
		return err
	}

	return nil
}

func ValidateUpdateBridgeRequest(request model.UpdateBridgeRequest, bridgeInterfaceName string) error {

	commonInterfaceName, haveCommon := utils.FindCommonMemberInterfaces(request.AddToBridge, request.RemoveFromBridge)
	if haveCommon {
		return fmt.Errorf("Cannot add and remove the same interface %s to the bridge at the same time.", commonInterfaceName)
	}

	for _, memIntfToAdd := range request.AddToBridge {
		if memIntfToAdd != model.EMPTY_STRING {
			addLink, err := netlink.LinkByName(memIntfToAdd)
			if err != nil {
				return fmt.Errorf("Member interface to add %s not found", memIntfToAdd)
			}
			if addLink.Type() != model.DEVICE {
				return fmt.Errorf("Member Interface %s to add is not a physical interface", memIntfToAdd)
			}

			masterIndex := addLink.Attrs().MasterIndex
			masterLink, err := netlink.LinkByIndex(masterIndex)
			if err != nil {
				fmt.Printf("Master with index %d not found, with error : %s", masterIndex, err.Error())
			} else if masterIndex != 0 && masterLink.Type() == model.BRIDGE_STRING && masterLink.Attrs().Name == bridgeInterfaceName {
				return fmt.Errorf("Interface %s is already a member interface of bridge %s. ", memIntfToAdd, masterLink.Attrs().Name)
			} else if masterIndex != 0 && masterLink.Type() == model.BRIDGE_STRING {
				return fmt.Errorf("Interface %s is already a member interface of bridge %s. Please remove from there first to add it here. ", memIntfToAdd, masterLink.Attrs().Name)
			} else if masterIndex != 0 && masterLink.Type() == model.BOND_STRING {
				return fmt.Errorf("Interface %s is a slave of bond %s. Please unslave from there first.", memIntfToAdd, masterLink.Attrs().Name)
			}
		}
	}

	bridgeLink, err := netlink.LinkByName(bridgeInterfaceName)
	if err != nil {
		return fmt.Errorf("Bridge interface %s not found", bridgeInterfaceName)
	}

	for _, memIntfToRemove := range request.RemoveFromBridge {
		if memIntfToRemove != model.EMPTY_STRING {
			removeLink, err := netlink.LinkByName(memIntfToRemove)
			if err != nil {
				return fmt.Errorf("Interface to remove %s not found", memIntfToRemove)
			}

			if removeLink.Attrs().MasterIndex != bridgeLink.Attrs().Index {
				return fmt.Errorf("Interface %s is not a member interface of the specified bridge %s", memIntfToRemove, bridgeInterfaceName)
			}
		}
	}

	err = validateHardwareAddress(request.HardwareAddress)
	if err != nil {
		return err
	}

	err = validateIPv4Details(request.IPv4Details)
	if err != nil {
		return err
	}

	err = validateIPv6Details(request.IPv6Details)
	if err != nil {
		return err
	}

	err = validateMTUDetails(request.MTU)
	if err != nil {
		return err
	}

	err = validateMSSDetails(request.MssDetails)
	if err != nil {
		return err
	}

	err = validateStpDetails(request.StpDetails)
	if err != nil {
		return err
	}

	err = isValidNetworkZone(request.NetworkZone)
	if err != nil {
		return err
	}

	return nil
}

// vlan validations

func ValidateCreateVlanRequest(request model.CreateVlanRequest) error {

	if err := utils.ValidateVlanId(request.VlanID); err != nil {
		return err
	}

	if request.VlanInterfaceName == model.EMPTY_STRING {
		return fmt.Errorf("Vlan interface name is required")
	}

	if request.ParentInterface == model.EMPTY_STRING {
		return fmt.Errorf("Parent interface name is required")
	}

	_, err := netlink.LinkByName(request.VlanInterfaceName)
	if err == nil {
		return fmt.Errorf("Vlan interface %s already exists", request.VlanInterfaceName)
	}

	_, err = netlink.LinkByName(request.ParentInterface)
	if err != nil {
		return fmt.Errorf("Parent interface %s not found", request.ParentInterface)
	}

	err = validateIPv4Details(request.IPv4Details)
	if err != nil {
		return err
	}

	err = validateIPv6Details(request.IPv6Details)
	if err != nil {
		return err
	}

	if request.IPv4Details.IPv4 {
		if !utils.StringOfTypeStaticOrDhcp(request.IPv4Details.IPv4AssignmentMode) {
			return fmt.Errorf("Invalid IPv4 assignment mode: %s, only static & dhcp are allowed in this mode.", request.IPv4Details.IPv4AssignmentMode)
		}
	}

	if request.IPv6Details.IPv6 {
		if !utils.StringOfTypeStaticOrDhcp(request.IPv6Details.IPv6AssignmentMode) {
			return fmt.Errorf("Invalid IPv6 assignment mode: %s, only static & dhcp are allowed in this mode.", request.IPv6Details.IPv6AssignmentMode)
		}
	}

	err = isValidNetworkZone(request.NetworkZone)
	if err != nil {
		return err
	}

	return nil
}

func ValidateUpdateVlanRequest(request model.UpdateVlanRequest, vlanLink netlink.Link) error {

	err := validateIPv4Details(request.IPv4Details)
	if err != nil {
		return err
	}

	err = validateIPv6Details(request.IPv6Details)
	if err != nil {
		return err
	}

	err = isValidNetworkZone(request.NetworkZone)
	if err != nil {
		return err
	}

	return nil
}

// bond validations
func ValidateCreateBondRequest(request model.CreateBondRequest) error {

	if len(request.SlaveInterfaces) < 2 {
		return fmt.Errorf("Atleast 2 slave interfaces are required to create a bond")
	}

	hasCommom, commonInterfaceName := utils.HasCommon(request.SlaveInterfaces)
	if hasCommom {
		return fmt.Errorf("Cannot add the same interface %s to the bond multiple times", commonInterfaceName)
	}

	for _, memIntf := range request.SlaveInterfaces {
		memIntfName := strings.TrimSpace(memIntf)
		if memIntfName == model.EMPTY_STRING {
			return fmt.Errorf("Slave interface name cannot be empty")
		}
		memLink, err := netlink.LinkByName(memIntfName)
		if err != nil {
			return fmt.Errorf("Interface %s not found", memIntfName)
		}
		if memLink.Type() != model.DEVICE {
			return fmt.Errorf("Interface %s is not a physical interface", memIntfName)
		}
		if memLink.Attrs().OperState.String() == model.LINK_STATE_UP {
			return fmt.Errorf("Slave Interface %s is up, please bring it down first", memIntfName)
		}
	}

	err := validateBondMode(request.BondMode)
	if err != nil {
		return err
	}

	err = validateHardwareAddress(request.HardwareAddress)
	if err != nil {
		return err
	}

	err = validateIPv4Details(request.IPv4Details)
	if err != nil {
		return err
	}

	err = validateMTUDetails(request.MTU)
	if err != nil {
		return err
	}

	err = validateMSSDetails(request.MssDetails)
	if err != nil {
		return err
	}

	err = isValidNetworkZone(request.NetworkZone)
	if err != nil {
		return err
	}

	return nil
}

func ValidateUpdateBondRequest(request model.UpdateBondRequest, bondInterfaceName string) error {

	commonInterfaceName, haveCommon := utils.FindCommonMemberInterfaces(request.AddToBond, request.RemoveFromBond)
	if haveCommon {
		return fmt.Errorf("Cannot add and remove the same interface %s to the bond at the same time.", commonInterfaceName)
	}

	for _, slaveIntfToAdd := range request.AddToBond {
		if slaveIntfToAdd == model.EMPTY_STRING {
			return fmt.Errorf("Slave interface name cannot be empty")
		}
		addLink, err := netlink.LinkByName(slaveIntfToAdd)
		if err != nil {
			return fmt.Errorf("Member interface to add %s not found", slaveIntfToAdd)
		}
		if addLink.Type() != model.DEVICE {
			return fmt.Errorf("Interface %s is not a physical interface", slaveIntfToAdd)
		}
		masterIndex := addLink.Attrs().MasterIndex
		masterLink, err := netlink.LinkByIndex(masterIndex)
		if err != nil {
			fmt.Printf("Master with index %d not found, with error : %s", masterIndex, err.Error())
		} else if masterIndex != 0 && masterLink.Type() == model.BOND_STRING && masterLink.Attrs().Name == bondInterfaceName {
			return fmt.Errorf("Interface %s is already a slave interface of bond %s. ", slaveIntfToAdd, masterLink.Attrs().Name)
		} else if masterIndex != 0 && masterLink.Type() == model.BOND_STRING {
			return fmt.Errorf("Interface %s is already a slave interface of bond %s. Please unslave from there first. ", slaveIntfToAdd, masterLink.Attrs().Name)
		} else if masterIndex != 0 && masterLink.Type() == model.BRIDGE_STRING {
			return fmt.Errorf("Interface %s is a member of bridge %s. Please remove from there first.", slaveIntfToAdd, masterLink.Attrs().Name)
		}
		if addLink.Attrs().OperState.String() == model.LINK_STATE_UP {
			return fmt.Errorf("Slave Interface %s is already up, Please bring it down first", slaveIntfToAdd)
		}
	}

	bondLink, err := netlink.LinkByName(bondInterfaceName)
	if err != nil {
		return fmt.Errorf("Bond interface %s not found", bondInterfaceName)
	}

	for _, slaveIntfToRemove := range request.RemoveFromBond {
		if slaveIntfToRemove == model.EMPTY_STRING {
			return fmt.Errorf("Slave interface name to be removed cannot be empty")
		}
		removeLink, err := netlink.LinkByName(slaveIntfToRemove)
		if err != nil {
			return fmt.Errorf("Interface to remove %s not found", slaveIntfToRemove)
		}
		if removeLink.Attrs().MasterIndex != bondLink.Attrs().Index {
			return fmt.Errorf("Interface %s is not a slave of the specified bond %s", slaveIntfToRemove, bondInterfaceName)
		}
	}

	err = validateIPv4Details(request.IPv4Details)
	if err != nil {
		return err
	}

	err = validateIPv6Details(request.IPv6Details)
	if err != nil {
		return err
	}

	err = validateHardwareAddress(request.HardwareAddress)
	if err != nil {
		return err
	}

	err = validateMTUDetails(request.MTU)
	if err != nil {
		return err
	}

	err = isValidNetworkZone(request.NetworkZone)
	if err != nil {
		return err
	}

	return nil
}

func validateHardwareAddress(hwAddr string) error {
	if strings.TrimSpace(hwAddr) != model.EMPTY_STRING {
		_, err := utils.ValidateAndGetHardwareAddress(hwAddr)
		if err != nil {
			return fmt.Errorf("Invalid hardware address: %s", err.Error())
		}
	}
	return nil
}

func validateIPv4Details(IPv4Details model.IPv4Details) error {
	if IPv4Details.IPv4 {
		if !utils.StringOfTypeStaticOrDhcp(IPv4Details.IPv4AssignmentMode) {
			return fmt.Errorf("Invalid IPv4 assignment mode: %s", IPv4Details.IPv4AssignmentMode)
		}

		if IPv4Details.IPv4AssignmentMode == model.STATIC_STRING {
			if IPv4Details.IPAddress == model.EMPTY_STRING {
				return fmt.Errorf("IP address is required for static assignment mode")
			}
			validated, ipversion := utils.ValidateIP(IPv4Details.IPAddress)
			if !validated {
				return fmt.Errorf("Invalid IP address format")
			} else {
				if ipversion != model.IPV4_STRING {
					return fmt.Errorf("Expecting Ipv4 protocol version")
				}
			}
			if IPv4Details.Netmask == model.EMPTY_STRING {
				return fmt.Errorf("Netmask is required for static assignment mode")
			}
			validated, _ = utils.ValidateIPNetmask(IPv4Details.IPAddress, IPv4Details.Netmask)
			if !validated {
				return fmt.Errorf("Invalid Ip address/netmask")
			}
		}
	}
	return nil
}

func validateIPv6Details(IPv6Details model.IPv6Details) error {
	if IPv6Details.IPv6 {
		if !utils.StringOfTypeStaticOrDhcp(IPv6Details.IPv6AssignmentMode) {
			return fmt.Errorf("Invalid IPv6 assignment mode: %s", IPv6Details.IPv6AssignmentMode)
		}

		if IPv6Details.IPv6AssignmentMode == model.STATIC_STRING {
			if IPv6Details.IPAddress == model.EMPTY_STRING {
				return fmt.Errorf("IP address is required for static assignment mode")
			}
			validated, ipversion := utils.ValidateIP(IPv6Details.IPAddress)
			if !validated {
				return fmt.Errorf("Invalid IP address format")
			} else {
				if ipversion != model.IPV6_STRING {
					return fmt.Errorf("Expecting Ipv6 protocol version")
				}
			}
			if IPv6Details.Prefix == model.EMPTY_STRING {
				return fmt.Errorf("Prefix is required for ipv6 static assignment mode")
			}
			validated, _ = utils.ValidateIPNetmask(IPv6Details.IPAddress, IPv6Details.Prefix)
			if !validated {
				return fmt.Errorf("Invalid Ip address/netmask")
			}
		}
	}
	return nil
}

func validateBondMode(bondMode string) error {

	if bondMode == model.EMPTY_STRING {
		return fmt.Errorf("Bond mode cannot be empty")
	}

	if bondMode != "active-backup" && bondMode != "802.3ad" {
		return fmt.Errorf("Invalid bond mode: %s , Possible modes are: active-backup, 802.3ad", bondMode)
	}

	if bondMode != model.EMPTY_STRING && netlink.StringToBondMode(bondMode) == netlink.BOND_MODE_UNKNOWN {
		return fmt.Errorf("Invalid bond mode: %s", bondMode)
	}
	return nil
}

func validateMTUDetails(MTU int) error {
	if MTU != 0 {
		if err := utils.ValidateMTU(MTU); err != nil {
			return fmt.Errorf("Invalid MTU: %s", err.Error())
		}
	}
	return nil
}

func validateMSSDetails(MssDetails model.MssDetails) error {
	if MssDetails.OverRideMSS && MssDetails.MSS != 0 {
		if err := utils.ValidateMSS(MssDetails.MSS); err != nil {
			return fmt.Errorf("Invalid MSS: %s", err.Error())
		}
	}
	return nil
}

func validateStpDetails(StpDetails model.StpDetails) error {
	if StpDetails.TurnOnStp && StpDetails.StpMaxAge != 0 {
		if err := utils.ValidateStpMaxAge(StpDetails.StpMaxAge); err != nil {
			return fmt.Errorf("Invalid range for STP Max Age: %s", err.Error())
		}
	}
	return nil
}

func isValidNetworkZone(networkZone string) error {
	if networkZone == model.EMPTY_STRING {
		return fmt.Errorf("Network zone cannot be empty")
	}

	upperNetworkZone := strings.ToLower(networkZone)

	switch upperNetworkZone {
	case model.LAN_STRING, model.WAN_STRING, model.DMZ_STRING:
		return nil
	default:
		return fmt.Errorf("Invalid network zone: %s", networkZone)
	}
}

// routing validations

func ValidateCreateStaticRouteRequest(req model.CreateStaticRouteRequest) error {

	if req.Operation == model.Type_IPV4_UNICAST {
		if req.Ipv4UnicastRoute.InterfaceName == model.EMPTY_STRING {
			return fmt.Errorf("Interface name is required")
		}

		_, err := netlink.LinkByName(req.Ipv4UnicastRoute.InterfaceName)
		if err != nil {
			return fmt.Errorf("Interface %s not found", req.Ipv4UnicastRoute.InterfaceName)
		}

		if req.Ipv4UnicastRoute.DestinationIp == model.EMPTY_STRING {
			return fmt.Errorf("Destination IP is required")
		}

		if req.Ipv4UnicastRoute.Netmask == model.EMPTY_STRING {
			return fmt.Errorf("Netmask is required")
		}

		validated, ipnet := utils.ValidateIPNetmask(req.Ipv4UnicastRoute.DestinationIp, req.Ipv4UnicastRoute.Netmask)
		if !validated {
			return fmt.Errorf("Invalid destination IP/netmask")
		}

		if ipnet.IP.To4() == nil {
			return fmt.Errorf("Invalid destination IP, expecting an IPv4 address")
		}

		if req.Ipv4UnicastRoute.Gateway != model.EMPTY_STRING {
			validated, _ := utils.ValidateIP(req.Ipv4UnicastRoute.Gateway)
			if !validated {
				return fmt.Errorf("Invalid gateway IP")
			}
		}

		metric := req.Ipv4UnicastRoute.Metric
		if metric < 0 || metric > 255 {
			return fmt.Errorf("Invalid metric value, possible range is (0-255)")
		}

		ad := req.Ipv4UnicastRoute.AdministrativeDistance
		if ad != 0 && (ad < 1 || ad > 255) {
			return fmt.Errorf("Invalid administrative distance value, possible range is (1-255)")
		}
	} else if req.Operation == model.Type_IPV6_UNICAST {
		if req.Ipv6UnicastRoute.InterfaceName == model.EMPTY_STRING {
			return fmt.Errorf("Interface name is required")
		}

		_, err := netlink.LinkByName(req.Ipv6UnicastRoute.InterfaceName)
		if err != nil {
			return fmt.Errorf("Interface %s not found", req.Ipv6UnicastRoute.InterfaceName)
		}

		if req.Ipv6UnicastRoute.DestinationIp == model.EMPTY_STRING {
			return fmt.Errorf("Destination IP is required")
		}

		if req.Ipv6UnicastRoute.Prefix == model.EMPTY_STRING {
			return fmt.Errorf("Prefix is required")
		}

		validated, ipnet := utils.ValidateIPNetmask(req.Ipv6UnicastRoute.DestinationIp, req.Ipv6UnicastRoute.Prefix)
		if !validated {
			return fmt.Errorf("Invalid destination IP/prefix")
		}

		if ipnet.IP.To4() != nil {
			return fmt.Errorf("Invalid destination IP, expecting an IPv6 address")
		}

		if req.Ipv6UnicastRoute.Gateway != model.EMPTY_STRING {
			validated, _ := utils.ValidateIP(req.Ipv6UnicastRoute.Gateway)
			if !validated {
				return fmt.Errorf("Invalid gateway IP")
			}
		}

		metric := req.Ipv6UnicastRoute.Metric
		if metric < 1 || metric > 255 {
			return fmt.Errorf("Invalid metric value, possible range is (1-255)")
		}

	} else if req.Operation == model.Type_MULTICAST {

		if req.MulticastRoute.SourceInterface == model.EMPTY_STRING || req.MulticastRoute.DestinationInterface == model.EMPTY_STRING || req.MulticastRoute.SourceIpAddress == model.EMPTY_STRING || req.MulticastRoute.MulticastIpv4Address == model.EMPTY_STRING {
			return fmt.Errorf("All fields are mandatory")
		}

		if req.MulticastRoute.SourceInterface == req.MulticastRoute.DestinationInterface {
			return fmt.Errorf("Source and destination interface cannot be same")
		}

		_, err := netlink.LinkByName(req.MulticastRoute.SourceInterface)
		if err != nil {
			return fmt.Errorf("Source interface %s not found", req.MulticastRoute.SourceInterface)
		}

		_, err = netlink.LinkByName(req.MulticastRoute.DestinationInterface)
		if err != nil {
			return fmt.Errorf("Destination interface %s not found", req.MulticastRoute.DestinationInterface)
		}

		validated, _ := utils.ValidateIP(req.MulticastRoute.SourceIpAddress)
		if !validated {
			return fmt.Errorf("Invalid source IP")
		}

		validated, addrType := utils.ValidateIP(req.MulticastRoute.MulticastIpv4Address)
		if !validated {
			return fmt.Errorf("Invalid multicast group IP")
		}

		if addrType != model.IPV4_STRING {
			return fmt.Errorf("Invalid multicast IP, expecting an IPv4 address")
		}

		if !utils.IsMulticastIPv4(req.MulticastRoute.MulticastIpv4Address) {
			return fmt.Errorf("Invalid multicast IP, multicast IP range is (224.0.0.0 - 239.255.255.255)")
		}

	}

	return nil
}

func ValidateCreateBgpRequest(req model.CreateBgpRequest) error {

	if req.Operation == model.Type_NEIGHBOR {
		if strings.ToLower(req.Neighbor.IpVersion) != model.IPV4_STRING && strings.ToLower(req.Neighbor.IpVersion) != model.IPV6_STRING {
			return fmt.Errorf("Invalid IP version: %s", req.Neighbor.IpVersion)
		}

		if req.Neighbor.IpAddress == model.EMPTY_STRING {
			return fmt.Errorf("IP address is required")
		}

		validated, ipver := utils.ValidateIP(req.Neighbor.IpAddress)
		if !validated {
			return fmt.Errorf("Invalid IP address")
		}

		reqIpVersion := strings.ToLower(req.Neighbor.IpVersion)

		if reqIpVersion == model.IPV4_STRING && ipver != model.IPV4_STRING {
			return fmt.Errorf("Invalid IP address, expecting an IPv4 address")
		} else if reqIpVersion == model.IPV6_STRING && ipver != model.IPV6_STRING {
			return fmt.Errorf("Invalid IP address, expecting an IPv6 address")
		}

		if !validateAs(req.Neighbor.RemoteAs) {
			return fmt.Errorf("Invalid remote AS: %s, possible range is (1-4294967295)", req.Neighbor.RemoteAs)
		}

	} else if req.Operation == model.Type_NETWORK {

		if req.Network.IpAddress == model.EMPTY_STRING {
			return fmt.Errorf("IP address is required")
		}

		if strings.ToLower(req.Network.IpVersion) != model.IPV4_STRING && strings.ToLower(req.Network.IpVersion) != model.IPV6_STRING {
			return fmt.Errorf("Invalid IP version: %s", req.Network.IpVersion)
		}

		if req.Network.Netmask == model.EMPTY_STRING {
			return fmt.Errorf("Netmask/Prefix is required")
		}

		validated, ip := utils.ValidateIPNetmask(req.Network.IpAddress, req.Network.Netmask)
		if !validated {
			return fmt.Errorf("Invalid IP address/netmask")
		}

		reqIpver := strings.ToLower(req.Network.IpVersion)

		if reqIpver == model.IPV4_STRING && ip.IP.To4() == nil {
			return fmt.Errorf("Invalid IP address, expecting an IPv4 address")
		} else if reqIpver == model.IPV6_STRING && ip.IP.To16() == nil {
			return fmt.Errorf("Invalid IP address, expecting an IPv6 address")
		}

	}

	return nil
}

func ValidateCreateOspfRequest(req model.CreateOspfRequest) error {

	if req.Operation == model.Type_NETWORK {
		if req.OspfNetwork.IpAddress == model.EMPTY_STRING {
			return fmt.Errorf("IP address is required")
		}

		if req.OspfNetwork.Netmask == model.EMPTY_STRING {
			return fmt.Errorf("Netmask is required")
		}

		validated, ipnet := utils.ValidateIPNetmask(req.OspfNetwork.IpAddress, req.OspfNetwork.Netmask)
		if !validated {
			return fmt.Errorf("Invalid IP address/netmask")
		}

		ip := ipnet.IP.To4()
		if ip == nil {
			return fmt.Errorf("Invalid IP address, expecting an IPv4 address")
		}

		if req.OspfNetwork.Area == model.EMPTY_STRING {
			return fmt.Errorf("Area is required")
		}

		validated, _ = utils.ValidateIP(req.OspfNetwork.Area)
		if !validated {
			return fmt.Errorf("Invalid area")
		}

	} else if req.Operation == model.Type_AREA {
		if req.OspfArea.Area == model.EMPTY_STRING {
			return fmt.Errorf("Area is required")
		}

		validated, _ := utils.ValidateIP(req.OspfArea.Area)
		if !validated {
			return fmt.Errorf("Invalid area")
		}

		areaType := strings.ToLower(req.OspfArea.AreaType)
		authentication := strings.ToLower(req.OspfArea.Authentication)

		if areaType == model.EMPTY_STRING && authentication == model.EMPTY_STRING {
			return fmt.Errorf("Either area type or authentication is required")
		}

		if areaType != model.EMPTY_STRING && !validateOspfAreaType(areaType) {
			return fmt.Errorf("Invalid area type: %s", areaType)
		}

		if authentication != model.EMPTY_STRING && !validateOspfAuthentication(authentication) {
			return fmt.Errorf("Invalid authentication: %s", authentication)
		}

		if areaType == "normal" && (authentication == model.EMPTY_STRING && len(req.OspfArea.VirtualLinks) == 0) {
			return fmt.Errorf("Virtual links or Authentication is required in case of area type is normal")
		}

		if areaType != model.EMPTY_STRING && areaType != "normal" && req.OspfArea.AreaCost != model.EMPTY_STRING {
			areaCost, err := strconv.Atoi(req.OspfArea.AreaCost)
			if err != nil {
				return fmt.Errorf("Invalid area cost: %s", req.OspfArea.AreaCost)
			}
			if areaCost < 0 || areaCost > 16777214 {
				return fmt.Errorf("Invalid area cost, possible range is (0-16777214)")
			}
		}

	}

	return nil
}

func validateOspfAreaType(areaType string) bool {
	if areaType != "normal" && areaType != "stub" && areaType != "nssa" && areaType != "stub no-summary" && areaType != "nssa no-summary" {
		return false
	}
	return true
}

func validateOspfAuthentication(authentication string) bool {
	if authentication != "text" && authentication != "md5" {
		return false
	}
	return true
}

func ValidateCreateOspfv3Request(req model.CreateOspfv3Request) error {

	if req.Operation == model.Type_INTERFACE {
		if req.OspfInterface.InterfaceName == model.EMPTY_STRING {
			return fmt.Errorf("Interface name is required")
		}

		_, err := netlink.LinkByName(req.OspfInterface.InterfaceName)
		if err != nil {
			return fmt.Errorf("Interface %s not found", req.OspfInterface.InterfaceName)
		}

		if req.OspfInterface.Area == model.EMPTY_STRING {
			return fmt.Errorf("Area is required")
		}

		validated, _ := utils.ValidateIP(req.OspfInterface.Area)
		if !validated {
			return fmt.Errorf("Invalid area")
		}

		if req.OspfInterface.HelloInterval != 0 && (req.OspfInterface.HelloInterval < 1 || req.OspfInterface.HelloInterval > 65535) {
			return fmt.Errorf("Invalid hello interval, possible range is (1-65535)")
		}

		if req.OspfInterface.DeadInterval != 0 && (req.OspfInterface.DeadInterval < 1 || req.OspfInterface.DeadInterval > 65535) {
			return fmt.Errorf("Invalid dead interval, possible range is (1-65535)")
		}

		if req.OspfInterface.ReTransmitInterval != 0 && (req.OspfInterface.ReTransmitInterval < 1 || req.OspfInterface.ReTransmitInterval > 65535) {
			return fmt.Errorf("Invalid retransmit interval, possible range is (1-65535)")
		}

		if req.OspfInterface.TransmitDelay != 0 && (req.OspfInterface.TransmitDelay < 1 || req.OspfInterface.TransmitDelay > 3600) {
			return fmt.Errorf("Invalid transmit delay, possible range is (1-3600)")
		}

		if req.OspfInterface.InterfaceCost != 0 && (req.OspfInterface.InterfaceCost < 1 || req.OspfInterface.InterfaceCost > 65535) {
			return fmt.Errorf("Invalid interface cost, possible range is (1-65535)")
		}

		if req.OspfInterface.InstanceId < 0 || req.OspfInterface.InstanceId > 255 {
			return fmt.Errorf("Invalid instance id, possible range is (0-255)")
		}

		if req.OspfInterface.RouterPriority < 0 || req.OspfInterface.RouterPriority > 255 {
			return fmt.Errorf("Invalid priority, possible range is (0-255)")
		}

	} else if req.Operation == model.Type_AREA {
		if req.OspfArea.Area == model.EMPTY_STRING {
			return fmt.Errorf("Area is required")
		}

		validated, _ := utils.ValidateIP(req.OspfArea.Area)
		if !validated {
			return fmt.Errorf("Invalid area")
		}

		areaType := strings.ToLower(req.OspfArea.AreaType)

		if areaType == model.EMPTY_STRING {
			return fmt.Errorf("Area type is required")
		}

		if !validateOspfv3AreaType(areaType) {
			return fmt.Errorf("Invalid area type: %s", areaType)
		}

	}

	return nil
}

func validateOspfv3AreaType(areaType string) bool {
	if areaType != "stub" && areaType != "nssa" && areaType != "stub no-summary" && areaType != "nssa no-summary" {
		return false
	}
	return true
}

func ValidateBgpGlobalConfig(req model.BgpGlobalConfiguration) error {

	if req.LocalAs == model.EMPTY_STRING {
		return fmt.Errorf("Local AS is required")
	}

	if !validateAs(req.LocalAs) {
		return fmt.Errorf("Invalid local AS: %s, possible range is (1-4294967295)", req.LocalAs)
	}

	if !validateRouterIdAssignmentMode(req.RouterIdAssignment) {
		return fmt.Errorf("Invalid router id assignment mode: %s", req.RouterIdAssignment)
	}

	validated, addrType := utils.ValidateIP(req.RouterId)
	if !validated {
		return fmt.Errorf("Invalid router id")
	}

	if addrType != model.IPV4_STRING {
		return fmt.Errorf("Invalid router id, expecting an IPv4 address")
	}

	return nil
}

func validateRouterIdAssignmentMode(routerIdAssignmentMode string) bool {

	routerIdAssignmentMode = strings.ToLower(routerIdAssignmentMode)

	if routerIdAssignmentMode != "automatic" && routerIdAssignmentMode != "manual" {
		return false
	}
	return true
}

func validateAs(as string) bool {

	intAs, err := strconv.Atoi(as)
	if err != nil {
		return false
	}

	if intAs < 1 || intAs > 4294967295 {
		return false
	}

	return true
}

func ValidateOspfGlobalConfig(req model.OspfGlobalConfiguration) error {

	if req.RouterId != model.EMPTY_STRING {
		validated, addrType := utils.ValidateIP(req.RouterId)
		if !validated {
			return fmt.Errorf("Invalid router id")
		}

		if addrType != model.IPV4_STRING {
			return fmt.Errorf("Invalid router id, expecting an IPv4 address")
		}
	}

	if req.AbrType != model.EMPTY_STRING && !validateAbrType(req.AbrType) {
		return fmt.Errorf("Invalid ABR Type")
	}

	if req.DefaultMetric != model.EMPTY_STRING && !validateDefaultMetric(req.DefaultMetric) {
		return fmt.Errorf("Invalid Default Metric")
	}

	if req.Acrb != model.EMPTY_STRING && !validateAcrb(req.Acrb) {
		return fmt.Errorf("Invalid Auto-cost reference-bandwidth")
	}

	if req.DefInfoOriginate != model.EMPTY_STRING && !validateDefInfoOriginate(req.DefInfoOriginate) {
		return fmt.Errorf("Invalid Default-information originate")
	}

	defInfoOriginate := strings.ToLower(req.DefInfoOriginate)
	if (defInfoOriginate == "regular" || defInfoOriginate == "always") && !validateMetricDetails(req.DefInfoOriginateMetric) {
		return fmt.Errorf("Invalid Metric or Metric Type for Default-information originate")
	}

	if req.ReDistConnected && !validateMetricDetails(req.ReDistConnectedMetric) {
		return fmt.Errorf("Invalid Metric or Metric Type for Redistribute Connected")
	}

	if req.ReDistStatic && !validateMetricDetails(req.ReDistStaticMetric) {
		return fmt.Errorf("Invalid Metric or Metric Type for Redistribute Static")
	}

	if req.ReDistBgp && !validateMetricDetails(req.ReDistBgpMetric) {
		return fmt.Errorf("Invalid Metric or Metric Type for Redistribute BGP")
	}

	return nil
}

func ValidateOspfv3GlobalConfig(req model.Ospfv3GlobalConfiguration) error {

	if req.RouterId != model.EMPTY_STRING {
		validated, addrType := utils.ValidateIP(req.RouterId)
		if !validated {
			return fmt.Errorf("Invalid router id")
		}

		if addrType != model.IPV4_STRING {
			return fmt.Errorf("Invalid router id, expecting an IPv4 address")
		}
	}

	if req.DefaultMetric != model.EMPTY_STRING && !validateDefaultMetric(req.DefaultMetric) {
		return fmt.Errorf("Invalid Default Metric")
	}

	if req.AbrType != model.EMPTY_STRING && !validateAbrType(req.AbrType) {
		return fmt.Errorf("Invalid ABR Type")
	}

	if req.Acrb != model.EMPTY_STRING && !validateAcrb(req.Acrb) {
		return fmt.Errorf("Invalid Auto-cost reference-bandwidth")
	}

	if req.DefInfoOriginate != model.EMPTY_STRING && !validateDefInfoOriginate(req.DefInfoOriginate) {
		return fmt.Errorf("Invalid Default-information originate")
	}

	defInfoOriginate := strings.ToLower(req.DefInfoOriginate)
	if (defInfoOriginate == "regular" || defInfoOriginate == "always") && !validateMetricDetails(req.DefInfoOriginateMetric) {
		return fmt.Errorf("Invalid Metric or Metric Type for Default-information originate")
	}

	if req.ReDistConnected && !validateMetricDetails(req.ReDistConnectedMetric) {
		return fmt.Errorf("Invalid Metric or Metric Type for Redistribute Connected")
	}

	return nil
}

func validateAbrType(abrType string) bool {
	abrType = strings.ToLower(abrType)

	if abrType != "cisco" && abrType != "standard" && abrType != "ibm" && abrType != "shortcut" {
		return false
	}
	return true
}

func validateDefaultMetric(metric string) bool {
	metricInt, err := strconv.Atoi(metric)
	if err != nil {
		return false
	}

	if metricInt < 0 || metricInt > 16777214 {
		return false
	}

	return true
}

func validateMetricDetails(req model.MetricDetails) bool {
	metricType := strings.ToLower(req.MetricType)

	if metricType != "external type 1" && metricType != "external type 2" {
		return false
	}

	if req.Metric != model.EMPTY_STRING {
		metricInt, err := strconv.Atoi(req.Metric)
		if err != nil {
			return false
		}

		if metricInt < 0 || metricInt > 16777214 {
			return false
		}
	}

	return true
}

func validateAcrb(acrbS string) bool {

	acrb, err := strconv.Atoi(acrbS)
	if err != nil {
		return false
	}

	if acrb < 1 || acrb > 4294967 {
		return false
	}
	return true
}

func validateDefInfoOriginate(s string) bool {

	s = strings.ToLower(s)

	if s != "never" && s != "regular" && s != "always" {
		return false
	}

	return true
}

package interface_utils

import (
	"anti-apt-backend/extras"
	model "anti-apt-backend/model/interface_model"
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/vishvananda/netlink"
)

func CheckInterfaceExists(interfaceName string) error {
	_, err := netlink.LinkByName(strings.TrimSpace(interfaceName))
	return err
}

func CheckInterfacesExist(interfaceNames []string) (string, error) {
	for _, interfaceName := range interfaceNames {
		err := CheckInterfaceExists(interfaceName)
		if err != nil {
			return interfaceName, err
		}
	}
	return "", nil
}

func GetPrimaryIPAddressV2(link netlink.Link) (resp model.IpAddressResponse) {
	addrs, err := netlink.AddrList(link, model.FAMILY_ALL)
	if err != nil || len(addrs) == 0 {
		return resp
	}
	haveDhcp, addr := CheckForDynamicAddressV2(link)
	if haveDhcp {
		return addr
	}

	for _, addr := range addrs {
		if addr.Flags == 128 && addr.Label != model.EMPTY_STRING {
			ones, _ := addr.Mask.Size()
			resp.IpAddress = addr.IP.String()
			resp.Netmask = ones
			if addr.Flags == 0 {
				resp.Protocol = model.DHCP_STRING
			} else {
				resp.Protocol = model.STATIC_STRING
			}
			return resp
		}
	}

	return resp
}

func GetPrimaryIPAddress(link netlink.Link) string {
	addrs, err := netlink.AddrList(link, model.FAMILY_ALL)
	if err != nil || len(addrs) == 0 {
		return model.EMPTY_STRING
	}
	haveDhcp, addr := CheckForDynamicAddress(link)
	if haveDhcp {
		return addr
	}

	for _, addr := range addrs {
		if addr.Flags == 128 {
			ones, _ := addr.Mask.Size()
			return addr.IP.String() + "/" + fmt.Sprintf("%d", ones)
		}
	}

	return model.EMPTY_STRING
}

func PerformIpDel(link netlink.Link, deletingAddr string, newAddr string, newNetmask string) error {

	beforeDeleteList := GetIPAddressList(link)
	var found bool

	if deletingAddr != model.EMPTY_STRING {

		for _, beforeAddr := range beforeDeleteList {
			if beforeAddr == deletingAddr {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("Address %s not found on interface %s", deletingAddr, link.Attrs().Name)
		}

		fmt.Println("Deleting address : ", deletingAddr)
		ip := strings.Split(deletingAddr, "/")
		_, ipnet := ValidateIPNetmask(ip[0], ip[1])

		addrToRemove := &netlink.Addr{
			IPNet: ipnet,
		}

		fmt.Println("Address to remove : ", addrToRemove)
		if err := netlink.AddrDel(link, addrToRemove); err != nil {
			fmt.Println("error : ", err)
			return fmt.Errorf("Error while deleting address %s: %v", deletingAddr, err)
		}
	}

	if newAddr != model.EMPTY_STRING && newNetmask != model.EMPTY_STRING {
		_, ipnet := ValidateIPNetmask(newAddr, newNetmask)
		addrToAdd := &netlink.Addr{
			IPNet: ipnet,
		}
		if err := netlink.AddrAdd(link, addrToAdd); err != nil {
			return fmt.Errorf("Error while adding address %s: %v", newAddr, err)
		}
	}

	afterDeleteList := GetIPAddressList(link)

	beforeDeleteMap := make(map[string]bool)
	afterDeleteMap := make(map[string]bool)

	for _, beforeAddr := range beforeDeleteList {
		beforeDeleteMap[beforeAddr] = true
	}

	for _, afterAddr := range afterDeleteList {
		afterDeleteMap[afterAddr] = true
	}

	for beforeAddr := range beforeDeleteMap {
		if !afterDeleteMap[beforeAddr] && beforeAddr != deletingAddr {
			ip := strings.Split(beforeAddr, "/")
			_, ipnet := ValidateIPNetmask(ip[0], ip[1])
			addrToAdd := &netlink.Addr{
				IPNet: ipnet,
			}
			if err := netlink.AddrAdd(link, addrToAdd); err != nil {
				return fmt.Errorf("Error while adding address %s: %v", beforeAddr, err)
			}
		}
	}

	return nil
}

func FindProtocol(addr netlink.Addr) string {
	if addr.Flags == 0 {
		return model.DHCP_STRING
	}
	return model.STATIC_STRING
}

func GetSecondaryIPAddressListV2(link netlink.Link) (resp []model.IpAddressResponse) {

	primaryIp := GetPrimaryIPAddressV2(link)

	addrs, err := netlink.AddrList(link, model.FAMILY_ALL)
	if err != nil || len(addrs) == 0 {
		return resp
	}

	for _, addr := range addrs {
		ones, _ := addr.Mask.Size()
		if primaryIp.IpAddress+"/"+fmt.Sprintf("%d", primaryIp.Netmask) != addr.IP.String()+"/"+fmt.Sprintf("%d", ones) {
			resp = append(resp, model.IpAddressResponse{
				IpAddress: addr.IP.String(),
				Netmask:   ones,
				Protocol:  FindProtocol(addr),
			})
		}
	}

	return resp
}

func CheckForDynamicAddressV2(link netlink.Link) (bool, model.IpAddressResponse) {
	var resp model.IpAddressResponse
	addrs, err := netlink.AddrList(link, model.FAMILY_ALL)
	if err != nil || len(addrs) == 0 {
		return false, resp
	}
	for _, addr := range addrs {
		if addr.Flags == 0 || addr.Flags == 512 {
			ones, _ := addr.Mask.Size()
			resp.IpAddress = addr.IP.String()
			resp.Netmask = ones
			if addr.Flags == 0 || addr.Flags == 512 {
				resp.Protocol = model.DHCP_STRING
			} else {
				resp.Protocol = model.STATIC_STRING
			}
			return true, resp
		}
	}
	return false, resp
}

// func which checks if any of ip address in the address list have flag 0
func CheckForDynamicAddress(link netlink.Link) (bool, string) {
	addrs, err := netlink.AddrList(link, model.FAMILY_ALL)
	if err != nil || len(addrs) == 0 {
		return false, model.EMPTY_STRING
	}
	for _, addr := range addrs {
		if addr.Flags == 0 {
			ones, _ := addr.Mask.Size()
			return true, addr.IP.String() + "/" + fmt.Sprintf("%d", ones)
		}
	}
	return false, model.EMPTY_STRING
}

func GetIPAddressList(link netlink.Link) []string {
	addrs, err := netlink.AddrList(link, model.FAMILY_ALL)
	if err != nil || len(addrs) == 0 {
		return []string{}
	}
	ipList := []string{}
	for _, addr := range addrs {
		ones, _ := addr.Mask.Size()
		ipList = append(ipList, addr.IP.String()+"/"+fmt.Sprintf("%d", ones))
	}
	return ipList
}

func GetIPProtocol(link netlink.Link) string {
	addrs, err := netlink.AddrList(link, model.FAMILY_ALL)
	if err != nil || len(addrs) == 0 {
		return model.EMPTY_STRING
	}
	for _, addr := range addrs {
		if addr.Flags == 0 || addr.Flags == 512 {
			return model.DHCP_STRING
		}
	}
	return model.STATIC_STRING
}

func GetIPVersion(link netlink.Link) int {

	primaryAddr := GetPrimaryIPAddress(link)
	if primaryAddr != model.EMPTY_STRING {
		_, ipNet, _ := net.ParseCIDR(primaryAddr)
		if ipNet.IP.To4() != nil {
			return 4
		}
		return 6
	}

	return 0
}

func PerformIpFlush(name string) error {

	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("Error while fetching link while addr flush: %v", err)
	}

	addrs, err := netlink.AddrList(link, model.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("Error while fetching address list: %v", err)
	}

	for _, addr := range addrs {
		if err := netlink.AddrDel(link, &addr); err != nil {
			return fmt.Errorf("Error while deleting address %s: %v", addr.IP.String(), err)
		}
	}

	return nil
}

func SetIP(interfaceName, ip string, netmask int) error {
	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return err
	}

	netmaskStr := fmt.Sprintf("%d", netmask)

	if ip != model.EMPTY_STRING && netmask != 0 {
		_, ipnet := ValidateIPNetmask(ip, netmaskStr)
		addrToAdd := &netlink.Addr{
			IPNet: ipnet,
		}
		if err := netlink.AddrAdd(link, addrToAdd); err != nil {
			return fmt.Errorf("Error while adding address %s: %v", ip, err)
		}
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		return err
	}

	return nil
}

func IsHaMonitored(intfName string, haConfig model.Ha) bool {
	for _, haIntf := range haConfig.MonitoredInterfaces {
		if haIntf.Interface == intfName {
			return true
		}
	}
	return false
}

func IsHaPrimary(haConfig model.Ha) bool {
	applianceRole := strings.ToLower(strings.TrimSpace(haConfig.ApplianceRole))
	if applianceRole == extras.PRIMARY_STRING {
		return true
	}
	return false
}

func IsHaBackup(haConfig model.Ha) bool {
	applianceRole := strings.ToLower(strings.TrimSpace(haConfig.ApplianceRole))
	if applianceRole == extras.BACKUP_STRING {
		return true
	}
	return false
}

func IsHaDedicated(intfName string, haConfig model.Ha) bool {
	if haConfig.DedicatedHaInterface == intfName {
		return true
	}
	return false
}

func FetchHaIps(intfName string, haConfig model.Ha) (string, string) {
	for _, haIntf := range haConfig.MonitoredInterfaces {
		if haIntf.Interface == intfName {
			return haIntf.BaseIp, haIntf.PeerIp
		}
	}
	return extras.EMPTY_STRING, extras.EMPTY_STRING
}

func FetchInterfaceIpFromHaConfig(intfName string, haConfig model.Ha) string {
	for _, haIntf := range haConfig.MonitoredInterfaces {
		if haIntf.Interface == intfName {
			return haIntf.InterfaceIp
		}
	}
	return extras.EMPTY_STRING

}

func CheckIfIpAlreadyExists(ip string) bool {
	links, err := netlink.LinkList()
	if err != nil {
		return false
	}

	for _, link := range links {
		addrs, err := netlink.AddrList(link, model.FAMILY_ALL)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if addr.IP.String() == ip {
				return true
			}
		}
	}
	return false
}

func ValidateMTU(mtu int) error {
	if mtu < 576 || mtu > 9000 {
		return fmt.Errorf("MTU value is out of a valid range")
	}
	return nil
}

func ValidateMSS(mss int) error {
	if mss < 536 || mss > 8960 {
		return fmt.Errorf("MSS value is out of a valid range")
	}
	return nil
}

func ValidateStpMaxAge(sma int) error {
	if sma < 6 || sma > 40 {
		return fmt.Errorf("STP Max Age value is out of a valid range")
	}
	return nil
}

func ValidateVlanId(vid int) error {
	if vid < 1 || vid > 2094 {
		return fmt.Errorf("VLan Id value is out of a valid range")
	}
	return nil
}

func ValidateAndGetHardwareAddress(hwAddr string) (net.HardwareAddr, error) {
	hw, err := net.ParseMAC(hwAddr)
	if err != nil {
		return nil, fmt.Errorf("Invalid hardware address format")
	}
	return hw, nil
}

func IsMulticastIPv4(address string) bool {
	ip := net.ParseIP(address)
	if ip == nil {
		return false
	}

	if ip.To4() == nil {
		return false
	}

	return ip[12] >= 224 && ip[12] <= 239
}

func ValidateIP(ip string) (bool, string) {
	parsedIP := net.ParseIP(ip)
	if parsedIP.To4() != nil {
		return true, model.IPV4_STRING
	} else if parsedIP.To16() != nil {
		return true, model.IPV6_STRING
	}
	return parsedIP != nil, ""
}

func ValidateIPNetmask(ipAddress string, netmask string) (bool, *net.IPNet) {

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false, nil
	}

	_, ipnet, err := net.ParseCIDR(ipAddress + "/" + netmask)
	if err != nil {
		return false, nil
	}

	IPNet := &net.IPNet{
		IP:   ip,
		Mask: ipnet.Mask,
	}

	return true, IPNet
}

func StringOfTypeStaticOrDhcp(a string) bool {
	if a == model.STATIC_STRING || a == model.DHCP_STRING {
		return true
	}
	return false
}

func TrimStringsInStruct(input interface{}) interface{} {
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.Struct {
		return nil
	}

	out := reflect.New(inputValue.Type()).Elem()

	for i := 0; i < inputValue.NumField(); i++ {
		field := inputValue.Field(i)
		fieldType := field.Type()

		if fieldType.Kind() == reflect.String {
			trimmedValue := strings.TrimSpace(field.String())
			out.Field(i).SetString(trimmedValue)
		} else if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.String {
			outArray := make([]string, 0, field.Len())

			for j := 0; j < field.Len(); j++ {
				element := field.Index(j).String()
				trimmedElement := strings.TrimSpace(element)
				outArray = append(outArray, trimmedElement)
			}

			out.Field(i).Set(reflect.ValueOf(outArray))
		} else if fieldType.Kind() == reflect.Struct {
			nestedStruct := field.Interface()
			trimmedNestedStruct := TrimStringsInStruct(nestedStruct)
			out.Field(i).Set(reflect.ValueOf(trimmedNestedStruct))
		} else {
			out.Field(i).Set(field)
		}

	}

	return out.Interface()
}

func FindCommonMemberInterfaces(arr1, arr2 []string) (string, bool) {
	elementSet := make(map[string]bool)

	for _, val := range arr1 {
		if val != model.EMPTY_STRING {
			elementSet[val] = true
		}
	}

	for _, val := range arr2 {
		if val != model.EMPTY_STRING && elementSet[val] {
			return val, true
		}
	}

	return "", false
}

func HasCommon(arr []string) (bool, string) {
	m := make(map[string]bool)
	for _, s := range arr {
		if m[s] {
			return true, s
		}
		m[s] = true
	}
	return false, ""
}

func CheckIfMemberInterfaceHasMaster(interfaceName string) (bool, error) {
	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return false, err
	}

	masterIndex := link.Attrs().MasterIndex
	if masterIndex != 0 {
		masterLink, err := netlink.LinkByIndex(masterIndex)
		if err != nil {
			return false, err
		}

		masterType := masterLink.Type()
		switch masterType {
		case model.BOND_STRING:
			return true, fmt.Errorf("Interface %s is already a slave of a bond - "+masterLink.Attrs().Name, interfaceName)
		case model.BRIDGE_STRING:
			return true, fmt.Errorf("Interface %s is already a member of bridge - "+masterLink.Attrs().Name, interfaceName)
		default:
			return false, nil
		}
	}

	return false, nil
}

func FetchAddRemoveFromList(orig []string, req []string) (toAdd []string, toRemove []string) {

	origMap := make(map[string]bool)
	reqMap := make(map[string]bool)

	for _, item := range orig {
		origMap[item] = true
	}

	for _, item := range req {
		reqMap[item] = true
	}

	for _, item := range req {
		if !origMap[item] {
			toAdd = append(toAdd, item)
		}
	}

	for _, item := range orig {
		if !reqMap[item] {
			toRemove = append(toRemove, item)
		}
	}

	return toAdd, toRemove
}

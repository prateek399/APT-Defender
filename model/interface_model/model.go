package interface_model

import (
	"net/http"

	"golang.org/x/sys/unix"
)

// APIResponse is the response format for all the API's
type APIResponse struct {
	Data       interface{} `json:"data,omitempty"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message,omitempty"`
	Error      string      `json:"error,omitempty"`
}

func NewSuccessResponse(successMessage string, data interface{}) *APIResponse {
	return &APIResponse{
		StatusCode: http.StatusOK,
		Data:       data,
		Message:    successMessage,
	}
}

func NewErrorResponse(statusCode int, errorMessage string, err error) *APIResponse {
	return &APIResponse{
		StatusCode: statusCode,
		Message:    errorMessage + " : " + err.Error(),
	}
}

// physical interface API's request and response formats

type PortMappingResponse struct {
	PortMap []PortMap `json:"port_map"`
}

type PortMap struct {
	PortName   string `json:"port_name"`
	Value      string `json:"value"`
	IsStaticIp bool   `json:"is_static_ip"`
}

type ListPhysicalInterfacesResponse struct {
	PhysicalInterfacesCount int                     `json:"physical_interfaces_count"`
	PhysicalInterfaces      []ListPhysicalInterface `json:"physical_interfaces"`
}

type Config struct {
	PhysicalInterfaces []ListPhysicalInterface `json:"physical_interfaces" yaml:"physical_interfaces"`
	BridgeInterfaces   []ListBridgeInterface   `json:"bridge_interfaces" yaml:"bridge_interfaces"`
	VLANInterfaces     []ListVlanInterface     `json:"vlan_interfaces" yaml:"vlan_interfaces"`
	BondInterfaces     []ListBondDetails       `json:"bond_interfaces" yaml:"bond_interfaces"`
	VirtualWires       []ListVirtualWire       `json:"virtual_wires" yaml:"virtual_wires"`
	StaticRoutes       ListStaticRoutes        `json:"static_routes" yaml:"static_routes"`
	BGP                ListBgp                 `json:"bgp" yaml:"bgp"`
	OSPF               ListOspf                `json:"ospf" yaml:"ospf"`
	OSPFV3             ListOspfv3              `json:"ospfv3" yaml:"ospfv3"`
	Ha                 Ha                      `json:"ha" yaml:"ha"`
}

type MonitoredInterface struct {
	Interface   string `json:"interface" yaml:"interface"`
	BaseIp      string `json:"base_ip" yaml:"base_ip"`
	PeerIp      string `json:"peer_ip" yaml:"peer_ip"`
	InterfaceIp string `json:"interface_ip" yaml:"interface_ip"`
}

type Ha struct {
	ApplianceMode            string               `json:"appliance_mode" yaml:"appliance_mode"`
	ApplianceRole            string               `json:"appliance_role" yaml:"appliance_role"`
	Password                 string               `json:"password" yaml:"password"`
	DedicatedHaInterface     string               `json:"dedicated_ha_interface" yaml:"dedicated_ha_interface"`
	PeerIp                   string               `json:"peer_ip" yaml:"peer_ip"`
	MonitoredInterfaces      []MonitoredInterface `json:"monitored_interfaces" yaml:"monitored_interfaces"`
	KeepAliveRequestInterval int                  `json:"keep_alive_request_interval" yaml:"keep_alive_request_interval"`
	KeepAliveAttempts        int                  `json:"keep_alive_attempts" yaml:"keep_alive_attempts"`
	HaStatus                 HaStatus             `json:"ha_status" yaml:"ha_status"`
}

type HaStatus struct {
	Status       string `json:"status" yaml:"status"`
	LastSyncedAt string `json:"last_synced_at" yaml:"last_synced_at"`
}

type CreateHaRequest struct {
	RequestType              int                  `json:"request_type"`
	ApplianceMode            string               `json:"appliance_mode"`
	ApplianceRole            string               `json:"appliance_role"`
	Password                 string               `json:"password"`
	DedicatedHaInterface     string               `json:"dedicated_ha_interface"`
	PeerIp                   string               `json:"peer_ip"`
	MonitoredInterfaces      []MonitoredInterface `json:"monitored_interfaces"`
	KeepAliveRequestInterval int                  `json:"keep_alive_request_interval"`
	KeepAliveAttempts        int                  `json:"keep_alive_attempts"`
}

type ConfigSpecificFields struct {
	ServingLocation string `json:"serving_location"`
	DomainName      string `json:"domain_name"`
	NetworkZone     string `json:"network_zone"`
	IsDisabled      bool   `json:"is_disabled"`
}

type ConfigSpecificFieldsMap map[string]ConfigSpecificFields

type ListPhysicalInterface struct {
	Name            string              `json:"name" yaml:"name"`
	HardwareAddress string              `json:"hardware_address" yaml:"hardware_address"`
	IpAddress       IpAddressResponse   `json:"ip_address" yaml:"ip_address"`
	AliasList       []IpAddressResponse `json:"alias_list" yaml:"alias_list"`
	MTU             int                 `json:"mtu" yaml:"mtu"`
	LinkState       string              `json:"link_state" yaml:"link_state"`
	LinkSpeed       string              `json:"link_speed" yaml:"link_speed"`
	LinkDuplex      string              `json:"link_duplex" yaml:"link_duplex"`
	LinkAutoneg     string              `json:"link_autoneg" yaml:"link_autoneg"`
	LinkStats       LinkStats           `json:"link_stats" yaml:"link_stats"`
	AttachedTo      []string            `json:"attached_to" yaml:"attached_to"`
	IsEditable      bool                `json:"is_editable" yaml:"is_editable"`
	IsDeletable     bool                `json:"is_deletable" yaml:"is_deletable"`
	IsDisabled      bool                `json:"is_disabled" yaml:"is_disabled"`
	ServingLocation string              `json:"serving_location" yaml:"serving_location"` // lan, wan, dmz
	DomainName      string              `json:"domain_name" yaml:"domain_name"`
	NetworkZone     string              `json:"network_zone" yaml:"network_zone"`
}

type UpdatePhysicalInterfaceRequest struct {
	IPv4Details     IPv4Details `json:"ipv4_details"`
	IPv6Details     IPv6Details `json:"ipv6_details"`
	MTU             int         `json:"mtu"`
	HardwareAddress string      `json:"hardware_address"`
	ServingLocation string      `json:"serving_location"`
	DomainName      string      `json:"domain_name"`
	NetworkZone     string      `json:"network_zone"`
	IsDisabled      bool        `json:"is_disabled"`
}

// link API's
type ListLinksRequest struct {
	LinkType string `json:"link_type"`
	VlocName string `json:"vloc"`
}

type ListLinksResponse struct {
	LinksCount int        `json:"links_count"`
	Links      []LinkInfo `json:"links"`
}

type LinkInfo struct {
	Name            string    `json:"link_name"`
	Alias           string    `json:"alias"`
	HardwareAddress string    `json:"hardware_address"`
	IPAddresses     []string  `json:"ip_addresses"`
	IPProtocol      string    `json:"ip_protocol"`
	MTU             int       `json:"mtu"`
	LinkType        string    `json:"link_type"`
	LinkState       string    `json:"link_state"`
	LinkSpeed       string    `json:"speed"`
	LinkDuplex      string    `json:"duplex"`
	LinkAutoneg     string    `json:"autoneg"`
	LinkStats       LinkStats `json:"link_stats"`
}

type LinkStats struct {
	TxPackets uint64 `json:"tx_packets" yaml:"tx_packets"`
	TxBytes   uint64 `json:"tx_bytes" yaml:"tx_bytes"`
	TxErrors  uint64 `json:"tx_errors" yaml:"tx_errors"`
	TxDropped uint64 `json:"tx_dropped" yaml:"tx_dropped"`
	RxPackets uint64 `json:"rx_packets" yaml:"rx_packets"`
	RxBytes   uint64 `json:"rx_bytes" yaml:"rx_bytes"`
	RxErrors  uint64 `json:"rx_errors" yaml:"rx_errors"`
	RxDropped uint64 `json:"rx_dropped" yaml:"rx_dropped"`
}

type AddAliasRequest struct {
	InterfaceName string      `json:"interface_name" binding:"required"`
	IpVersion     int         `json:"ip_version" binding:"required"`
	IPv4Details   IPv4Details `json:"ipv4_details"`
	IPv6Details   IPv6Details `json:"ipv6_details"`
}

type GetAliasResponse struct {
	Alias []IpAddressResponse `json:"alias"`
}

type DeleteAliasRequest struct {
	InterfaceName string `json:"interface_name"`
	IpAddress     string `json:"ip_address"`
}

type CreateVirtualWireRequest struct {
	VirtualWireName string   `json:"virtual_wire_name"`
	ChildInterfaces []string `json:"child_interfaces"`
	ServingLocation string   `json:"serving_location"`
}

type ListVirtualWire struct {
	VirtualWireName string   `json:"virtual_wire_name"`
	ChildInterfaces []string `json:"child_interfaces"`
	ServingLocation string   `json:"serving_location"`
}

type ListVirtualWiresResponse struct {
	VirtualWireCount int               `json:"virtual_wire_count"`
	VirtualWires     []ListVirtualWire `json:"virtual_wires"`
}

// bridge API's request and response formats
type CreateBridgeRequest struct {
	BridgeInterfaceName string      `json:"bridge_interface_name" binding:"required"`
	BridgeNameByUser    string      `json:"bridge_name_by_user"`
	HardwareAddress     string      `json:"hardware_address"`
	Description         string      `json:"description"`
	EnableRouting       bool        `json:"enable_routing"`
	MemberInterfaces    []string    `json:"member_interfaces"`
	IPv4Details         IPv4Details `json:"ipv4_details"`
	IPv6Details         IPv6Details `json:"ipv6_details"`
	MTU                 int         `json:"mtu"`
	MssDetails          MssDetails  `json:"mss_details"`
	StpDetails          StpDetails  `json:"stp_details"`
	PermitArpBroadcast  bool        `json:"permit_arp_broadcast"`
	ServingLocation     string      `json:"serving_location"`
	DomainName          string      `json:"domain_name"`
	NetworkZone         string      `json:"network_zone"`
}

type UpdateBridgeRequest struct {
	MemberInterfaces   []string    `json:"member_interfaces"`
	AddToBridge        []string    `json:"add_to_bridge"`
	RemoveFromBridge   []string    `json:"remove_from_bridge"`
	HardwareAddress    string      `json:"hardware_address"`
	EnableRouting      bool        `json:"enable_routing"`
	IPv4Details        IPv4Details `json:"ipv4_details"`
	IPv6Details        IPv6Details `json:"ipv6_details"`
	MTU                int         `json:"mtu"`
	MssDetails         MssDetails  `json:"mss_details"`
	StpDetails         StpDetails  `json:"stp_details"`
	PermitArpBroadcast bool        `json:"permit_arp_broadcast"`
	ServingLocation    string      `json:"serving_location"`
	DomainName         string      `json:"domain_name"`
	NetworkZone        string      `json:"network_zone"`
	IsDisabled         bool        `json:"is_disabled"`
}

type IpAddressResponse struct {
	IpAddress string `json:"ip_address" yaml:"ip_address"`
	Netmask   int    `json:"netmask" yaml:"netmask"`
	Protocol  string `json:"protocol" yaml:"protocol"`
}

type IpAddressInfo struct {
	RemoveFlag bool   `json:"remove_flag"`
	IpAddress  string `json:"ip_address"`
	Netmask    string `json:"netmask"`
}

type ListBridgeInterfaceRequest struct {
	BridgeInterfaceName string `json:"bridge_interface_name"`
}

type ListBridgeInterfacesResponse struct {
	BridgeCount      int                   `json:"bridge_count"`
	BridgeInterfaces []ListBridgeInterface `json:"bridge_interfaces"`
}

type ListBridgeInterface struct {
	BridgeInterfaceName     string                `json:"bridge_interface_name"`
	CountOfMemberInterfaces int                   `json:"count_of_member_interfaces"`
	HardwareAddress         string                `json:"hardware_address"`
	IpAddress               IpAddressResponse     `json:"ip_address"`
	AliasList               []IpAddressResponse   `json:"alias_list"`
	IpProtocol              string                `json:"ip_protocol"`
	IpVersion               int                   `json:"ip_version"`
	MemberInterfaces        []ListMemberInterface `json:"member_interfaces"`
	MTU                     int                   `json:"mtu"`
	IsEditable              bool                  `json:"is_editable"`
	IsDeletable             bool                  `json:"is_deletable"`
	ServingLocation         string                `json:"serving_location"`
	DomainName              string                `json:"domain_name"`
	NetworkZone             string                `json:"network_zone"`
	IsDisabled              bool                  `json:"is_disabled"`
}

// vlan API's request and response formats
type CreateVlanRequest struct {
	VlanInterfaceName string      `json:"vlan_interface_name" binding:"required"`
	VlanID            int         `json:"vlan_id" binding:"required"`
	ParentInterface   string      `json:"parent_interface" binding:"required"`
	IPv4Details       IPv4Details `json:"ipv4_details"`
	IPv6Details       IPv6Details `json:"ipv6_details"`
	ServingLocation   string      `json:"serving_location"`
	DomainName        string      `json:"domain_name"`
	NetworkZone       string      `json:"network_zone"`
}

type ListVlanInterface struct {
	VlanInterfaceName string              `json:"vlan_interface_name"`
	ParentInterface   ListMemberInterface `json:"parent_interface"`
	HardwareAddress   string              `json:"hardware_address"`
	IpAddress         IpAddressResponse   `json:"ip_address"`
	AliasList         []IpAddressResponse `json:"alias_list"`
	IpProtocol        string              `json:"ip_protocol"`
	IpVersion         int                 `json:"ip_version"`
	MTU               int                 `json:"mtu"`
	IsEditable        bool                `json:"is_editable"`
	IsDeletable       bool                `json:"is_deletable"`
	ServingLocation   string              `json:"serving_location"`
	DomainName        string              `json:"domain_name"`
	NetworkZone       string              `json:"network_zone"`
	IsDisabled        bool                `json:"is_disabled"`
}

type ListVlanInterfacesResp struct {
	VlanInterfacesCount int                 `json:"vlan_interfaces_count"`
	VlanInterfaces      []ListVlanInterface `json:"vlan_interfaces"`
}

type UpdateVlanRequest struct {
	IPv4Details     IPv4Details `json:"ipv4_details"`
	IPv6Details     IPv6Details `json:"ipv6_details"`
	ServingLocation string      `json:"serving_location"`
	DomainName      string      `json:"domain_name"`
	NetworkZone     string      `json:"network_zone"`
	IsDisabled      bool        `json:"is_disabled"`
}

type ListVlanInterfacesRequest struct {
	VlanInterfaceName string `json:"vlan_interface_name"`
}

// bond API's request and response formats
type ListBondInterfacesRequest struct {
	BondInterfaceName string `json:"bond_interface_name"`
}

type CreateBondRequest struct {
	BondInterfaceName string      `json:"bond_interface_name" binding:"required"`
	BondNameByUser    string      `json:"bond_name_by_user"`
	HardwareAddress   string      `json:"hardware_address"`
	SlaveInterfaces   []string    `json:"slave_interfaces"`
	BondMode          string      `json:"bond_mode"`
	IPv4Details       IPv4Details `json:"ipv4_details"`
	IPv6Details       IPv6Details `json:"ipv6_details"`
	MTU               int         `json:"mtu"`
	MssDetails        MssDetails  `json:"mss_details"`
	ServingLocation   string      `json:"serving_location"`
	DomainName        string      `json:"domain_name"`
	NetworkZone       string      `json:"network_zone"`
}

type UpdateBondRequest struct {
	SlaveInterfaces []string    `json:"slave_interfaces"`
	AddToBond       []string    `json:"add_to_bond"`
	RemoveFromBond  []string    `json:"remove_from_bond"`
	HardwareAddress string      `json:"hardware_address"`
	IPv4Details     IPv4Details `json:"ipv4_details"`
	IPv6Details     IPv6Details `json:"ipv6_details"`
	MTU             int         `json:"mtu"`
	ServingLocation string      `json:"serving_location"`
	DomainName      string      `json:"domain_name"`
	NetworkZone     string      `json:"network_zone"`
	IsDisabled      bool        `json:"is_disabled"`
}

type ListBondInterfacesResponse struct {
	BondCount      int               `json:"bond_count"`
	BondInterfaces []ListBondDetails `json:"bond_interfaces"`
}

type ListBondDetails struct {
	BondInterfaceName      string              `json:"bond_interface_name"`
	BondMode               string              `json:"bond_mode"`
	HardwareAddress        string              `json:"hardware_address"`
	IpAddress              IpAddressResponse   `json:"ip_address"`
	AliasList              []IpAddressResponse `json:"alias_list"`
	IpProtocol             string              `json:"ip_protocol"`
	IpVersion              int                 `json:"ip_version"`
	CountOfSlaveInterfaces int                 `json:"count_of_slave_interfaces"`
	SlaveInterfaces        []SlaveDetails      `json:"slave_interfaces"`
	MTU                    int                 `json:"mtu"`
	IsEditable             bool                `json:"is_editable"`
	IsDeletable            bool                `json:"is_deletable"`
	ServingLocation        string              `json:"serving_location"`
	DomainName             string              `json:"domain_name"`
	NetworkZone            string              `json:"network_zone"`
	IsDisabled             bool                `json:"is_disabled"`
}

type SlaveDetails struct {
	SlaveState      string `json:"slave_state"`
	InterfaceName   string `json:"interface_name"`
	HardwareAddress string `json:"hardware_address"`
	IpAddress       string `json:"ip_address"`
	IpProtocol      string `json:"ip_protocol"`
	IpVersion       int    `json:"ip_version"`
	MTU             int    `json:"mtu"`
}

type ListMemberInterface struct {
	InterfaceName   string `json:"interface_name"`
	HardwareAddress string `json:"hardware_address"`
	IpAddress       string `json:"ip_address"`
	IpProtocol      string `json:"ip_protocol"`
	IpVersion       int    `json:"ip_version"`
	MTU             int    `json:"mtu"`
}

type IPv4Details struct {
	IPv4               bool   `json:"ipv4"`
	IPv4AssignmentMode string `json:"ipv4_assignment_mode"`
	IPAddress          string `json:"ip_address"`
	Netmask            string `json:"netmask"`
	GatewayDetail      string `json:"gateway_detail"`
	GatewayName        string `json:"gateway_name"`
	GatewayIP          string `json:"gateway_ip"`
}

type IPv6Details struct {
	IPv6               bool   `json:"ipv6"`
	IPv6AssignmentMode string `json:"ipv6_assignment_mode"`
	IPAddress          string `json:"ip_address"`
	Prefix             string `json:"prefix"`
	GatewayDetail      string `json:"gateway_detail"`
	GatewayName        string `json:"gateway_name"`
	GatewayIP          string `json:"gateway_ip"`
}

type MssDetails struct {
	OverRideMSS bool `json:"over_ride_mss"`
	MSS         int  `json:"mss"`
}

type StpDetails struct {
	TurnOnStp bool `json:"turn_on_stp"`
	StpMaxAge int  `json:"stp_max_age"`
}

// constants
const (
	FAMILY_ALL              = unix.AF_UNSPEC
	IP_ADDRESS_STRING       = "ip_address"
	IP_PROTOCOL_STRING      = "ip_protocol"
	IP_VERSION_STRING       = "ip_version"
	MTU_STRING              = "mtu"
	HARDWARE_ADDRESS_STRING = "hardware_address"
	STATIC_STRING           = "static"
	DHCP_STRING             = "dhcp"
	IPV4_STRING             = "ipv4"
	IPV6_STRING             = "ipv6"
	BRIDGE_STRING           = "bridge"
	VLAN_STRING             = "vlan"
	BOND_STRING             = "bond"
	EMPTY_STRING            = ""
	ERROR_STRING            = "error"
	MESSAGE_STRING          = "message"
	DEVICE                  = "device"
	LAN_STRING              = "lan"
	WAN_STRING              = "wan"
	DMZ_STRING              = "dmz"
	LINK_STATE_UP           = "up"
	LINK_STATE_DOWN         = "down"
	LOOP_BACK_DEVICE        = "lo"
	STATIC_ROUTE            = "static_route"
	BGP                     = "bgp"
	OSPF                    = "ospf"
	OSPFV3                  = "ospfv3"
	HA_STRING               = "ha"
)

const (
	Type_IPV4_UNICAST = "ipv4_unicast"
	Type_IPV6_UNICAST = "ipv6_unicast"
	Type_MULTICAST    = "multicast"
	Type_NEIGHBOR     = "neighbor"
	Type_NETWORK      = "network"
	Type_AREA         = "area"
	Type_INTERFACE    = "interface"
	Type_BGP          = "bgp"
	Type_OSPF         = "ospf"
	Type_OSPFV3       = "ospfv3"
	Type_STATIC       = "static"
)

// routing API's request and response formats
type CreateStaticRouteRequest struct {
	Operation        string           `json:"operation"`
	Ipv4UnicastRoute Ipv4UnicastRoute `json:"ipv4_unicast_route"`
	Ipv6UnicastRoute Ipv6UnicastRoute `json:"ipv6_unicast_route"`
	MulticastRoute   MulticastRoute   `json:"multicast_route"`
}

type StaticRoute struct {
	Ipv4UnicastRoute Ipv4UnicastRoute `json:"ipv4_unicast_route"`
	Ipv6UnicastRoute Ipv6UnicastRoute `json:"ipv6_unicast_route"`
	MulticastRoute   MulticastRoute   `json:"multicast_route"`
}

type UpdateStaticRouteRequest struct {
	OldRoute StaticRoute `json:"old_route"`
	NewRoute StaticRoute `json:"new_route"`
}

type Ipv4UnicastRoute struct {
	DestinationIp          string `json:"destination_ip" yaml:"destination_ip"`
	Netmask                string `json:"netmask" yaml:"netmask"`
	Gateway                string `json:"gateway" yaml:"gateway"`
	InterfaceName          string `json:"interface_name" yaml:"interface_name"`
	AdministrativeDistance int    `json:"administrative_distance" yaml:"administrative_distance"`
	Metric                 int    `json:"metric" yaml:"metric"`
}

type Ipv6UnicastRoute struct {
	DestinationIp string `json:"destination_ip" yaml:"destination_ip"`
	Prefix        string `json:"prefix" yaml:"prefix"`
	Gateway       string `json:"gateway" yaml:"gateway"`
	InterfaceName string `json:"interface_name" yaml:"interface_name"`
	Metric        int    `json:"metric" yaml:"metric"`
}

type MulticastRoute struct {
	SourceIpAddress      string `json:"source_ip_address" yaml:"source_ip_address"`
	SourceInterface      string `json:"source_interface" yaml:"source_interface"`
	MulticastIpv4Address string `json:"multicast_ipv4_address" yaml:"multicast_ipv4_address"`
	DestinationInterface string `json:"destination_interface" yaml:"destination_interface"`
}

type ListStaticRoutes struct {
	Ipv4UnicastRoutes []Ipv4UnicastRoute `json:"ipv4_unicast_routes" yaml:"ipv4_unicast_routes"`
	Ipv6UnicastRoutes []Ipv6UnicastRoute `json:"ipv6_unicast_routes" yaml:"ipv6_unicast_routes"`
	MulticastRoutes   []MulticastRoute   `json:"multicast_routes" yaml:"multicast_routes"`
	GlobalConfig      StaticGlobalConfig `json:"global_config" yaml:"global_config"`
}

type DeleteStaticRouteRequest struct {
	Operation        string           `json:"operation"`
	Ipv4UnicastRoute Ipv4UnicastRoute `json:"ipv4_unicast_route"`
	Ipv6UnicastRoute Ipv6UnicastRoute `json:"ipv6_unicast_route"`
	MulticastRoute   MulticastRoute   `json:"multicast_route"`
}

type UpdateGlobalConfigRequest struct {
	Operation                 string                    `json:"operation"`
	StaticGlobalConfig        StaticGlobalConfig        `json:"static_global_configuration"`
	BgpGlobalConfiguration    BgpGlobalConfiguration    `json:"bgp_global_configuration"`
	OspfGlobalConfiguration   OspfGlobalConfiguration   `json:"ospf_global_configuration"`
	Ospfv3GlobalConfiguration Ospfv3GlobalConfiguration `json:"ospfv3_global_configuration"`
}

type StaticGlobalConfig struct {
	ForwardMulticast bool `json:"forward_multicast"`
}

// BGP API's request and response formats

type BgpGlobalConfiguration struct {
	RouterIdAssignment string `json:"router_id_assignment"`
	RouterId           string `json:"router_id"`
	LocalAs            string `json:"local_as"`
}

type CreateBgpRequest struct {
	Operation string   `json:"operation"`
	Neighbor  Neighbor `json:"neighbor"`
	Network   Network  `json:"network"`
}

type Neighbor struct {
	IpVersion string `json:"ip_version"`
	IpAddress string `json:"ip_address"`
	RemoteAs  string `json:"remote_as"`
}

type Network struct {
	IpVersion string `json:"ip_version"`
	IpAddress string `json:"ip_address"`
	Netmask   string `json:"netmask"`
}

type ListBgp struct {
	GlobalConfig BgpGlobalConfiguration `json:"global_config"`
	Neighbors    []Neighbor             `json:"neighbors"`
	Networks     []Network              `json:"networks"`
}

type DeleteBgpRequest struct {
	Operation string   `json:"operation"`
	Neighbor  Neighbor `json:"neighbor"`
	Network   Network  `json:"network"`
}

type UpdateBgpRequest struct {
	OldRoute BgpRoute `json:"old_route"`
	NewRoute BgpRoute `json:"new_route"`
}

type BgpRoute struct {
	Neighbor Neighbor `json:"neighbor"`
	Network  Network  `json:"network"`
}

//  OSPF API's request and response formats

type OspfGlobalConfiguration struct {
	RouterId               string        `json:"router_id"`
	DefaultMetric          string        `json:"default_metric"`
	AbrType                string        `json:"abr_type"`
	Acrb                   string        `json:"ac_ref_bw"` // auto-cost reference-bandwidth
	DefInfoOriginate       string        `json:"def_info_originate"`
	DefInfoOriginateMetric MetricDetails `json:"def_info_originate_metric"`
	ReDistConnected        bool          `json:"re_dist_connected"`
	ReDistConnectedMetric  MetricDetails `json:"re_dist_connected_metric"`
	ReDistStatic           bool          `json:"re_dist_static"`
	ReDistStaticMetric     MetricDetails `json:"re_dist_static_metric"`
	ReDistBgp              bool          `json:"re_dist_bgp"`
	ReDistBgpMetric        MetricDetails `json:"re_dist_bgp_metric"`
}

type MetricDetails struct {
	Metric     string `json:"metric"`
	MetricType string `json:"metric_type"`
}

type CreateOspfRequest struct {
	Operation   string      `json:"operation"`
	OspfNetwork OspfNetwork `json:"ospf_network"`
	OspfArea    OspfArea    `json:"ospf_area"`
}

type OspfNetwork struct {
	IpAddress string `json:"ip_address"`
	Netmask   string `json:"netmask"`
	Area      string `json:"area"`
}

type OspfArea struct {
	Area           string   `json:"area"`
	AreaType       string   `json:"area_type"`
	Authentication string   `json:"authentication"`
	VirtualLinks   []string `json:"virtual_links"`
	AreaCost       string   `json:"area_cost"`
}

type ListOspf struct {
	GlobalConfig OspfGlobalConfiguration `json:"global_config"`
	OspfNetworks []OspfNetwork           `json:"ospf_networks"`
	OspfAreas    []OspfArea              `json:"ospf_areas"`
}

type DeleteOspfRequest struct {
	Operation   string      `json:"operation"`
	OspfNetwork OspfNetwork `json:"ospf_network"`
	OspfArea    OspfArea    `json:"ospf_area"`
}

type UpdateOspfRequest struct {
	OldRoute OspfRoute `json:"old_route"`
	NewRoute OspfRoute `json:"new_route"`
}

type OspfRoute struct {
	OspfNetwork OspfNetwork `json:"ospf_network"`
	OspfArea    OspfArea    `json:"ospf_area"`
}

// OSPFv3 API's request and response formats

type Ospfv3GlobalConfiguration struct {
	RouterId               string        `json:"router_id"`
	DefaultMetric          string        `json:"default_metric"`
	AbrType                string        `json:"abr_type"`
	Acrb                   string        `json:"ac_ref_bw"` // auto-cost reference-bandwidth
	DefInfoOriginate       string        `json:"def_info_originate"`
	DefInfoOriginateMetric MetricDetails `json:"def_info_originate_metric"`
	ReDistConnected        bool          `json:"re_dist_connected"`
	ReDistConnectedMetric  MetricDetails `json:"re_dist_connected_metric"`
}

type CreateOspfv3Request struct {
	Operation     string        `json:"operation"`
	OspfInterface OspfInterface `json:"ospf_interface"`
	OspfArea      Ospfv3Area    `json:"ospf_area"`
}

type Ospfv3Area struct {
	Area           string `json:"area"`
	AreaType       string `json:"area_type"`
	Authentication string `json:"authentication"`
}

type OspfInterface struct {
	InterfaceName      string `json:"interface_name"`
	Area               string `json:"area"`
	HelloInterval      int    `json:"hello_interval"`
	DeadInterval       int    `json:"dead_interval"`
	ReTransmitInterval int    `json:"retransmit_interval"`
	TransmitDelay      int    `json:"transmit_delay"`
	InterfaceCost      int    `json:"interface_cost"`
	InstanceId         int    `json:"instance_id"`
	RouterPriority     int    `json:"router_priority"`
}

type ListOspfv3 struct {
	GlobalConfig   Ospfv3GlobalConfiguration `json:"global_config"`
	OspfInterfaces []OspfInterface           `json:"ospf_interfaces"`
	OspfAreas      []Ospfv3Area              `json:"ospf_areas"`
}

type DeleteOspfv3Request struct {
	Operation     string        `json:"operation"`
	OspfInterface OspfInterface `json:"ospf_interface"`
	OspfArea      Ospfv3Area    `json:"ospf_area"`
}

type UpdateOspfv3Request struct {
	OldRoute Ospfv3Route `json:"old_route"`
	NewRoute Ospfv3Route `json:"new_route"`
}

type Ospfv3Route struct {
	OspfInterface OspfInterface `json:"ospf_interface"`
	OspfArea      Ospfv3Area    `json:"ospf_area"`
}

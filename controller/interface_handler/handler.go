package interface_handler

import (
	"anti-apt-backend/extras"
	model "anti-apt-backend/model/interface_model"
	"anti-apt-backend/service"
	routing "anti-apt-backend/service/interface_routing"
	"anti-apt-backend/service/interfaces"
	"anti-apt-backend/service/interfaces/bond"
	"anti-apt-backend/service/interfaces/bridge"
	"anti-apt-backend/service/interfaces/vlan"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var TempRouter *gin.Engine

const (
	BRIDGE_INTERFACE_NAME   = "bridge_interface_name"
	VLAN_INTERFACE_NAME     = "vlan_interface_name"
	BOND_INTERFACE_NAME     = "bond_interface_name"
	PHYSICAL_INTERFACE_NAME = "physical_interface_name"
	LinkType                = "link_type"
)

func GetPortMapping(ctx *gin.Context) {

	data, err := interfaces.GetPortMapping()
	if err != nil {
		resp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to fetch port mapping", err)
		ctx.JSON(resp.StatusCode, resp)
		return
	}

	resp := model.NewSuccessResponse(extras.ERR_SUCCESS, data)
	ctx.JSON(resp.StatusCode, resp)
}

func ListPhysicalInterfacesHandler(c *gin.Context) {

	physicalInterfaceName := strings.TrimSpace(c.Param(PHYSICAL_INTERFACE_NAME))

	physicalInterfaces, err := interfaces.ListPhysicalInterfaces(physicalInterfaceName, "MAIN")
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to fetch physical interfaces", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, physicalInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

func UpdatePhysicalInterfaceHandler(c *gin.Context) {

	curUsr, usrResp := service.GetCurUsr(c)
	if usrResp.StatusCode != http.StatusOK {
		c.JSON(usrResp.StatusCode, usrResp)
		return
	}

	physicalInterfaceName := strings.TrimSpace(c.Param(PHYSICAL_INTERFACE_NAME))
	var request model.UpdatePhysicalInterfaceRequest

	if err := c.BindJSON(&request); err != nil {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	physicalInterfaces, err := interfaces.UpdatePhysicalInterface(physicalInterfaceName, request, curUsr)
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, "Failed to update", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, physicalInterfaces)
	if successResp.StatusCode == http.StatusOK && request.IPv4Details.IPv4 {
		// ips, err := interfaces.FetchIps()
		// if err != nil {
		// 	fmt.Println("Error in fetching ips + ", err)
		// }
		// TempRouter.Use(middlewares.HandleCors(ips))
		go service.ServiceActions([]string{extras.SERVICE_BACKEND}, extras.Restart, 5)

	}
	c.JSON(successResp.StatusCode, successResp)
}

// Vlan Handlers
func CreateVlanInterfaceHandler(c *gin.Context) {

	var req model.CreateVlanRequest

	if err := c.BindJSON(&req); err != nil {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	vlanInterfaces, err := vlan.CreateVlanInterface(req)
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to create VLAN interface", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, vlanInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

func ListVlanInterfacesHandler(c *gin.Context) {

	vlanInterfaceName := strings.TrimSpace(c.Param(VLAN_INTERFACE_NAME))

	vlanInterfaces, err := vlan.ListVlanInterfaces(model.ListVlanInterfacesRequest{VlanInterfaceName: vlanInterfaceName})
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to fetch VLAN interfaces", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, vlanInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

func DeleteVlanInterfaceHandler(c *gin.Context) {

	vlanInterfaceName := strings.TrimSpace(c.Param(VLAN_INTERFACE_NAME))

	vlanInterfaces, err := vlan.DeleteVlanInterface(model.ListVlanInterfacesRequest{VlanInterfaceName: vlanInterfaceName})
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to delete VLAN interface", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, vlanInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

func UpdateVlanInterfaceHandler(c *gin.Context) {

	vlanInterfaceName := strings.TrimSpace(c.Param(VLAN_INTERFACE_NAME))
	var request model.UpdateVlanRequest

	if err := c.BindJSON(&request); err != nil {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	vlanInterfaces, err := vlan.UpdateVlanInterface(vlanInterfaceName, request)
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to update VLAN interface", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, vlanInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

// Bond Handlers
func CreateBondedLink(c *gin.Context) {

	var req model.CreateBondRequest

	if err := c.BindJSON(&req); err != nil {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	bondInterfaces, err := bond.CreateBondedLink(req)
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to create bonded link", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, bondInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

func UpdateBondInterface(c *gin.Context) {

	bondInterfaceName := strings.TrimSpace(c.Param(BOND_INTERFACE_NAME))
	var request model.UpdateBondRequest

	if err := c.BindJSON(&request); err != nil {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	bondInterfaces, err := bond.UpdateBondInterface(bondInterfaceName, request)
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to update bond interface", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, bondInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

func ListBondInterfaces(c *gin.Context) {

	bondInterfaceName := strings.TrimSpace(c.Param(BOND_INTERFACE_NAME))

	bondInterfaces, err := bond.ListBondInterfaces(model.ListBondInterfacesRequest{BondInterfaceName: bondInterfaceName})
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to fetch bond interfaces", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, bondInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

func DeleteBondInterface(c *gin.Context) {

	bondInterfaceName := strings.TrimSpace(c.Param(BOND_INTERFACE_NAME))

	bondInterfaces, err := bond.DeleteBondInterface(model.ListBondInterfacesRequest{BondInterfaceName: bondInterfaceName})
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to delete bond interface", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, bondInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

// Bridge Handlers
func CreateBridgeInterfaceHandler(c *gin.Context) {

	var req model.CreateBridgeRequest

	if err := c.BindJSON(&req); err != nil {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	bridgeInterfaces, err := bridge.CreateBridgeInterface(req)
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to create bridge interface", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, bridgeInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

func ListBridgeInterfacesHandler(c *gin.Context) {

	bridgeName := strings.TrimSpace(c.Param(BRIDGE_INTERFACE_NAME))

	bridgeInterfaces, err := bridge.ListBridgeInterfaces(model.ListBridgeInterfaceRequest{BridgeInterfaceName: bridgeName})
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to fetch bridge interfaces", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, bridgeInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

func UpdateBridgeInterfaceHandler(c *gin.Context) {

	bridgeInterfaceName := strings.TrimSpace(c.Param(BRIDGE_INTERFACE_NAME))
	var request model.UpdateBridgeRequest

	if err := c.BindJSON(&request); err != nil {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	bridgeInterfaces, err := bridge.UpdateBridgeInterface(bridgeInterfaceName, request)
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to update bridge interface", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, bridgeInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

func DeleteBridgeInterfaceHandler(c *gin.Context) {

	bridgeInterfaceName := strings.TrimSpace(c.Param(BRIDGE_INTERFACE_NAME))

	bridgInterfaces, err := bridge.DeleteBridgeInterface(model.ListBridgeInterfaceRequest{BridgeInterfaceName: bridgeInterfaceName})
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to delete bridge interface", err)
		c.JSON(errorResp.StatusCode, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, bridgInterfaces)
	c.JSON(successResp.StatusCode, successResp)
}

// static route handlers
func CreateStaticRouteHandler(c *gin.Context) {
	var req model.CreateStaticRouteRequest

	curUsr, usrResp := service.GetCurUsr(c)
	if usrResp.StatusCode != http.StatusOK {
		c.JSON(usrResp.StatusCode, usrResp)
		return
	}

	if err := c.BindJSON(&req); err != nil {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	resp, err := routing.CreateStaticRoute(req, curUsr)
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to create static route", err)
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, resp)
	c.JSON(successResp.StatusCode, successResp)
}

func ListStaticRoutesHandler(c *gin.Context) {
	routes, err := routing.ListStaticRoutes()
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to fetch static route configs", err)
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, routes)
	c.JSON(successResp.StatusCode, successResp)
}

func DeleteStaticRouteHandler(c *gin.Context) {
	operation := strings.TrimSpace(c.Param("operation"))

	curUsr, usrResp := service.GetCurUsr(c)
	if usrResp.StatusCode != http.StatusOK {
		c.JSON(usrResp.StatusCode, usrResp)
		return
	}

	if len(operation) == 0 {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, extras.ErrInvalidOperation)
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	var req model.DeleteStaticRouteRequest
	req.Operation = operation
	if operation == model.Type_IPV4_UNICAST {
		req.Ipv4UnicastRoute.InterfaceName = strings.TrimSpace(c.Query("interface"))
		req.Ipv4UnicastRoute.DestinationIp = strings.TrimSpace(c.Query("ip"))
		req.Ipv4UnicastRoute.Netmask = strings.TrimSpace(c.Query("netmask"))
	} else if operation == model.Type_IPV6_UNICAST {
		req.Ipv6UnicastRoute.InterfaceName = strings.TrimSpace(c.Query("interface"))
		req.Ipv6UnicastRoute.DestinationIp = strings.TrimSpace(c.Query("ip"))
		req.Ipv6UnicastRoute.Prefix = strings.TrimSpace(c.Query("prefix"))
	} else if operation == model.Type_MULTICAST {
		req.MulticastRoute.SourceInterface = strings.TrimSpace(c.Query("source"))
		req.MulticastRoute.SourceIpAddress = strings.TrimSpace(c.Query("ip"))
		req.MulticastRoute.DestinationInterface = strings.TrimSpace(c.Query("destination"))
		req.MulticastRoute.MulticastIpv4Address = strings.TrimSpace(c.Query("multicast"))
	} else {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, "only ipv4_unicast, ipv6_unicast & multicast operations are allowed", extras.ErrInvalidOperation)
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	resp, err := routing.DeleteStaticRoute(req, curUsr)
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to delete static route config", err)
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, resp)
	c.JSON(successResp.StatusCode, successResp)
}

func UpdateStaticRouteHandler(c *gin.Context) {

	var req model.UpdateStaticRouteRequest
	if err := c.BindJSON(&req); err != nil {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	curUsr, usrResp := service.GetCurUsr(c)
	if usrResp.StatusCode != http.StatusOK {
		c.JSON(usrResp.StatusCode, usrResp)
		return
	}

	operation := strings.TrimSpace(c.Param("operation"))
	if len(operation) == 0 {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, "Operation cannot be empty", extras.ErrInvalidOperation)
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	if operation != model.Type_IPV4_UNICAST && operation != model.Type_IPV6_UNICAST && operation != model.Type_MULTICAST {
		errorResp := model.NewErrorResponse(http.StatusBadRequest, "only ipv4_unicast, ipv6_unicast & multicast operations are allowed", extras.ErrInvalidOperation)
		c.JSON(http.StatusBadRequest, errorResp)
		return
	}

	resp, err := routing.UpdateStaticRoute(operation, req, curUsr)
	if err != nil {
		errorResp := model.NewErrorResponse(http.StatusInternalServerError, "Failed to update static route config", err)
		c.JSON(http.StatusInternalServerError, errorResp)
		return
	}

	successResp := model.NewSuccessResponse(extras.ERR_SUCCESS, resp)
	c.JSON(successResp.StatusCode, successResp)

}

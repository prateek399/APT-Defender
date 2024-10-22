package interface_routing

import (
	config "anti-apt-backend/config/interface_config"
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	model "anti-apt-backend/model/interface_model"
	"anti-apt-backend/service/interfaces"
	utils "anti-apt-backend/util/interface_utils"
	validations "anti-apt-backend/validation/interface_validations"

	"fmt"
)

func CreateStaticRoute(request model.CreateStaticRouteRequest, curUsr string) (model.ListStaticRoutes, error) {

	var resp model.ListStaticRoutes

	rc := config.NewRoutingConfigService()
	conf, err := rc.FetchRouteConfig(model.STATIC_ROUTE)
	if err != nil {
		return resp, err
	}

	currStaticRoutes := conf.StaticRoutes

	req := utils.TrimStringsInStruct(request).(model.CreateStaticRouteRequest)

	err = validations.ValidateCreateStaticRouteRequest(req)
	if err != nil {
		return resp, err
	}

	if req.Operation == model.Type_IPV4_UNICAST {
		curRoutes := currStaticRoutes.Ipv4UnicastRoutes
		for _, route := range curRoutes {
			if route.InterfaceName == req.Ipv4UnicastRoute.InterfaceName && route.DestinationIp == req.Ipv4UnicastRoute.DestinationIp && route.Netmask == req.Ipv4UnicastRoute.Netmask {
				return resp, extras.ErrRouteAlreadyExists
			}
		}

		curRoutes = append(curRoutes, req.Ipv4UnicastRoute)
		currStaticRoutes.Ipv4UnicastRoutes = curRoutes
		conf = model.Config{
			StaticRoutes: currStaticRoutes,
		}

		err = rc.UpdateStaticRouteConfig(conf, model.Type_IPV4_UNICAST, "STATIC_ROUTE_CREATE")
		if err != nil {
			return resp, err
		}

		logger.LoggerFunc("info", logger.LoggerMessage("sysLog:ipv4 unicast static route created for interface: "+interfaces.PortMapping[req.Ipv4UnicastRoute.InterfaceName]+" destination: "+req.Ipv4UnicastRoute.DestinationIp+"/"+req.Ipv4UnicastRoute.Netmask+" by "+curUsr))

	} else if req.Operation == model.Type_IPV6_UNICAST {

		curRoutes := currStaticRoutes.Ipv6UnicastRoutes
		for _, route := range curRoutes {
			if route.InterfaceName == req.Ipv6UnicastRoute.InterfaceName && route.DestinationIp == req.Ipv6UnicastRoute.DestinationIp && route.Prefix == req.Ipv6UnicastRoute.Prefix {
				return resp, extras.ErrRouteAlreadyExists
			}
		}

		curRoutes = append(curRoutes, req.Ipv6UnicastRoute)
		currStaticRoutes.Ipv6UnicastRoutes = curRoutes
		conf = model.Config{
			StaticRoutes: currStaticRoutes,
		}

		err = rc.UpdateStaticRouteConfig(conf, model.Type_IPV6_UNICAST, "STATIC_ROUTE_CREATE")
		if err != nil {
			return resp, err
		}

		logger.LoggerFunc("info", logger.LoggerMessage("sysLog:ipv6 static route created for interface: "+interfaces.PortMapping[req.Ipv6UnicastRoute.InterfaceName]+" destination: "+req.Ipv6UnicastRoute.DestinationIp+"/"+req.Ipv6UnicastRoute.Prefix+" by "+curUsr))

	} else if req.Operation == model.Type_MULTICAST {

		curRoutes := currStaticRoutes.MulticastRoutes
		for _, route := range curRoutes {
			if route.SourceInterface == req.MulticastRoute.SourceInterface && route.SourceIpAddress == req.MulticastRoute.SourceIpAddress && route.DestinationInterface == req.MulticastRoute.DestinationInterface && route.MulticastIpv4Address == req.MulticastRoute.MulticastIpv4Address {
				return resp, extras.ErrRouteAlreadyExists
			}
		}

		curRoutes = append(curRoutes, req.MulticastRoute)
		currStaticRoutes.MulticastRoutes = curRoutes
		conf = model.Config{
			StaticRoutes: currStaticRoutes,
		}

		err = rc.UpdateStaticRouteConfig(conf, model.Type_MULTICAST, "STATIC_ROUTE_CREATE")
		if err != nil {
			return resp, err
		}

		logger.LoggerFunc("info", logger.LoggerMessage("sysLog:multicast static route created for source interface: "+interfaces.PortMapping[req.MulticastRoute.SourceInterface]+" source ip: "+req.MulticastRoute.SourceIpAddress+" destination interface: "+interfaces.PortMapping[req.MulticastRoute.DestinationInterface]+" multicast ip: "+req.MulticastRoute.MulticastIpv4Address+" by "+curUsr))

	} else {
		return resp, fmt.Errorf("unknown operation, only ipv4_unicast, ipv6_unicast and multicast operations are allowed")
	}

	resp, err = ListStaticRoutes()
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func ListStaticRoutes() (model.ListStaticRoutes, error) {
	var resp model.ListStaticRoutes

	rc := config.NewRoutingConfigService()
	config, err := rc.FetchRouteConfig(model.STATIC_ROUTE)
	if err != nil {
		return resp, err
	}

	if config.StaticRoutes.Ipv4UnicastRoutes == nil {
		config.StaticRoutes.Ipv4UnicastRoutes = []model.Ipv4UnicastRoute{}
	}
	if config.StaticRoutes.Ipv6UnicastRoutes == nil {
		config.StaticRoutes.Ipv6UnicastRoutes = []model.Ipv6UnicastRoute{}
	}
	if config.StaticRoutes.MulticastRoutes == nil {
		config.StaticRoutes.MulticastRoutes = []model.MulticastRoute{}
	}

	resp = config.StaticRoutes

	return resp, nil
}

func DeleteStaticRoute(req model.DeleteStaticRouteRequest, curUsr string) (resp model.ListStaticRoutes, err error) {

	rc := config.NewRoutingConfigService()
	conf, err := rc.FetchRouteConfig(model.STATIC_ROUTE)
	if err != nil {
		return resp, err
	}

	routes := conf.StaticRoutes
	var found bool
	if req.Operation == model.Type_IPV4_UNICAST {

		route := req.Ipv4UnicastRoute
		if route.InterfaceName == model.EMPTY_STRING || route.DestinationIp == model.EMPTY_STRING || route.Netmask == model.EMPTY_STRING {
			return resp, fmt.Errorf("InterfaceName, DestinationIp and Netmask are mandatory fields")
		}

		found = false
		for _, r := range routes.Ipv4UnicastRoutes {
			if r.InterfaceName == route.InterfaceName && r.DestinationIp == route.DestinationIp && r.Netmask == route.Netmask {
				found = true
				continue
			}
			resp.Ipv4UnicastRoutes = append(resp.Ipv4UnicastRoutes, r)
		}

		if !found {
			return resp, extras.ErrRouteNotFound
		}

		err = rc.UpdateStaticRouteConfig(model.Config{StaticRoutes: resp}, model.Type_IPV4_UNICAST, "STATIC_ROUTE_DELETE")
		if err != nil {
			return resp, err
		}

		logger.LoggerFunc("info", logger.LoggerMessage("sysLog:ipv4 unicast static route deleted for interface: "+interfaces.PortMapping[route.InterfaceName]+" destination: "+route.DestinationIp+"/"+route.Netmask+" by "+curUsr))

	} else if req.Operation == model.Type_IPV6_UNICAST {

		route := req.Ipv6UnicastRoute
		if route.InterfaceName == model.EMPTY_STRING || route.DestinationIp == model.EMPTY_STRING || route.Prefix == model.EMPTY_STRING {
			return resp, fmt.Errorf("InterfaceName, DestinationIp and Prefix are mandatory fields")
		}

		found = false
		for _, r := range routes.Ipv6UnicastRoutes {
			if r.InterfaceName == route.InterfaceName && r.DestinationIp == route.DestinationIp && r.Prefix == route.Prefix {
				found = true
				continue
			}
			resp.Ipv6UnicastRoutes = append(resp.Ipv6UnicastRoutes, r)
		}

		if !found {
			return resp, extras.ErrRouteNotFound
		}

		err = rc.UpdateStaticRouteConfig(model.Config{StaticRoutes: resp}, model.Type_IPV6_UNICAST, "STATIC_ROUTE_DELETE")
		if err != nil {
			return resp, err
		}

		logger.LoggerFunc("info", logger.LoggerMessage("sysLog:ipv6 unicast static route deleted for interface: "+interfaces.PortMapping[route.InterfaceName]+" destination: "+route.DestinationIp+"/"+route.Prefix+" by "+curUsr))

	} else if req.Operation == model.Type_MULTICAST {

		route := req.MulticastRoute
		if route.SourceInterface == model.EMPTY_STRING || route.SourceIpAddress == model.EMPTY_STRING || route.DestinationInterface == model.EMPTY_STRING || route.MulticastIpv4Address == model.EMPTY_STRING {
			return resp, fmt.Errorf("SourceInterface, SourceIpAddress, DestinationInterface and MulticastIpv4Address are mandatory fields")
		}

		found = false
		for _, r := range routes.MulticastRoutes {
			if r.SourceInterface == route.SourceInterface && r.SourceIpAddress == route.SourceIpAddress && r.DestinationInterface == route.DestinationInterface && r.MulticastIpv4Address == route.MulticastIpv4Address {
				found = true
				continue
			}
			resp.MulticastRoutes = append(resp.MulticastRoutes, r)
		}

		if !found {
			return resp, extras.ErrRouteNotFound
		}

		err = rc.UpdateStaticRouteConfig(model.Config{StaticRoutes: resp}, model.Type_MULTICAST, "STATIC_ROUTE_DELETE")
		if err != nil {
			return resp, err
		}

		logger.LoggerFunc("info", logger.LoggerMessage("sysLog:multicast static route deleted for source interface: "+interfaces.PortMapping[route.SourceInterface]+" source ip: "+route.SourceIpAddress+" destination interface: "+interfaces.PortMapping[route.DestinationInterface]+" multicast ip: "+route.MulticastIpv4Address+" by "+curUsr))

	} else {
		return resp, fmt.Errorf("invalid operation, only ipv4 unicast, ipv6 unicast and multicast delete operations are supported")
	}

	resp, err = ListStaticRoutes()
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func UpdateStaticRoute(operation string, req model.UpdateStaticRouteRequest, curUsr string) (resp model.ListStaticRoutes, err error) {

	rc := config.NewRoutingConfigService()
	conf, err := rc.FetchRouteConfig(model.STATIC_ROUTE)
	if err != nil {
		return resp, err
	}

	existingRoutes := conf.StaticRoutes
	var found bool
	oldRoute := req.OldRoute
	newRoute := req.NewRoute

	err = validations.ValidateCreateStaticRouteRequest(model.CreateStaticRouteRequest{
		Operation:        operation,
		Ipv4UnicastRoute: newRoute.Ipv4UnicastRoute,
		Ipv6UnicastRoute: newRoute.Ipv6UnicastRoute,
		MulticastRoute:   newRoute.MulticastRoute,
	})

	if err != nil {
		return resp, err
	}

	if operation == model.Type_IPV4_UNICAST {

		route := oldRoute.Ipv4UnicastRoute
		if route.InterfaceName == model.EMPTY_STRING || route.DestinationIp == model.EMPTY_STRING || route.Netmask == model.EMPTY_STRING {
			return resp, fmt.Errorf("InterfaceName, DestinationIp and Netmask are mandatory fields")
		}

		found = false
		for _, r := range existingRoutes.Ipv4UnicastRoutes {
			if r.InterfaceName == route.InterfaceName && r.DestinationIp == route.DestinationIp && r.Netmask == route.Netmask {
				found = true
				r = newRoute.Ipv4UnicastRoute

			}
			resp.Ipv4UnicastRoutes = append(resp.Ipv4UnicastRoutes, r)
		}

		if !found {
			return resp, extras.ErrRouteNotFound
		}

		err = rc.UpdateStaticRouteConfig(model.Config{StaticRoutes: resp}, model.Type_IPV4_UNICAST, "STATIC_ROUTE_UPDATE")
		if err != nil {
			return resp, err
		}

		logger.LoggerFunc("info", logger.LoggerMessage("sysLog:ipv4 unicast static route updated for interface: "+interfaces.PortMapping[route.InterfaceName]+" by "+curUsr))

	} else if operation == model.Type_IPV6_UNICAST {

		route := oldRoute.Ipv6UnicastRoute
		if route.InterfaceName == model.EMPTY_STRING || route.DestinationIp == model.EMPTY_STRING || route.Prefix == model.EMPTY_STRING {
			return resp, fmt.Errorf("InterfaceName, DestinationIp and Prefix are mandatory fields")
		}

		found = false
		for _, r := range existingRoutes.Ipv6UnicastRoutes {
			if r.InterfaceName == route.InterfaceName && r.DestinationIp == route.DestinationIp && r.Prefix == route.Prefix {
				found = true
				r = newRoute.Ipv6UnicastRoute
			}
			resp.Ipv6UnicastRoutes = append(resp.Ipv6UnicastRoutes, r)
		}

		if !found {
			return resp, extras.ErrRouteNotFound
		}

		err = rc.UpdateStaticRouteConfig(model.Config{StaticRoutes: resp}, model.Type_IPV6_UNICAST, "STATIC_ROUTE_UPDATE")
		if err != nil {
			return resp, err
		}

		logger.LoggerFunc("info", logger.LoggerMessage("sysLog:ipv6 unicast static route updated for interface: "+interfaces.PortMapping[route.InterfaceName]+" by "+curUsr))

	} else if operation == model.Type_MULTICAST {

		route := oldRoute.MulticastRoute
		if route.SourceInterface == model.EMPTY_STRING || route.SourceIpAddress == model.EMPTY_STRING || route.DestinationInterface == model.EMPTY_STRING || route.MulticastIpv4Address == model.EMPTY_STRING {
			return resp, fmt.Errorf("SourceInterface, SourceIpAddress, DestinationInterface and MulticastIpv4Address are mandatory fields")
		}

		found = false
		for _, r := range existingRoutes.MulticastRoutes {
			if r.SourceInterface == route.SourceInterface && r.SourceIpAddress == route.SourceIpAddress && r.DestinationInterface == route.DestinationInterface && r.MulticastIpv4Address == route.MulticastIpv4Address {
				found = true
				r = newRoute.MulticastRoute
			}
			resp.MulticastRoutes = append(resp.MulticastRoutes, r)
		}

		if !found {
			return resp, extras.ErrRouteNotFound
		}

		err = rc.UpdateStaticRouteConfig(model.Config{StaticRoutes: resp}, model.Type_MULTICAST, "STATIC_ROUTE_UPDATE")
		if err != nil {
			return resp, err
		}

		logger.LoggerFunc("info", logger.LoggerMessage("sysLog:multicast static route updated for source interface: "+interfaces.PortMapping[route.SourceInterface]+" by "+curUsr))

	} else {
		return resp, fmt.Errorf("invalid operation, only ipv4 unicast, ipv6 unicast and multicast update operations are supported")
	}

	resp, err = ListStaticRoutes()
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func UpdateStaticGlobalConfig(req model.StaticGlobalConfig) (resp model.ListStaticRoutes, err error) {

	rc := config.NewRoutingConfigService()
	conf, err := rc.FetchRouteConfig(model.STATIC_ROUTE)
	if err != nil {
		return resp, err
	}

	static := conf.StaticRoutes

	// err = validations.ValidateStaticGlobalConfig(req)
	// if err != nil {
	// 	return resp, err
	// }

	static.GlobalConfig = req

	err = rc.UpdateStaticRouteConfig(model.Config{StaticRoutes: static}, model.Type_STATIC, "STATIC_GLOBAL_CONFIG_UPDATE")
	if err != nil {
		return resp, err
	}

	resp, err = ListStaticRoutes()
	if err != nil {
		return resp, err
	}

	return resp, nil

}

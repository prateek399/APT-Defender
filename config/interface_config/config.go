package interface_config

import (
	model "anti-apt-backend/model/interface_model"
	"fmt"
	"os"
	"time"

	"github.com/ghodss/yaml"
	lock "github.com/subchen/go-trylock"

	"anti-apt-backend/extras"
)

const (
	LOCK_TIME_OUT = 3 * time.Second
)

var mu = lock.New()

func readConfigFile(fName string) ([]byte, error) {
	if _, err := os.Stat(fName); os.IsNotExist(err) {
		err = os.WriteFile(fName, []byte{}, 0644)
		if err != nil {
			return nil, fmt.Errorf("error creating config file: %v", err)
		}
	}

	yamlData, err := os.ReadFile(fName)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	return yamlData, nil
}

func UpdateConfig(resp model.Config, interfaceType string, interfaceName string, caller string) error {

	if ok := mu.TryLock(LOCK_TIME_OUT); !ok {
		return fmt.Errorf("Timeout while acquiring lock for %s - %s", caller, interfaceName)
	} else {
		fmt.Printf("Lock acquired for %s - %s\n", caller, interfaceName)
	}
	defer func() {
		mu.Unlock()
		fmt.Printf("Lock released for %s - %s\n", caller, interfaceName)
	}()

	yamlData, err := readConfigFile(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		return err
	}

	var config model.Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return err
	}

	switch interfaceType {
	case "ALL":
		var newData model.Config
		yamlResp, err := yaml.Marshal(resp)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlResp, &newData); err != nil {
			return err
		}
		config = newData
	case model.DEVICE:
		var newData []model.ListPhysicalInterface
		yamlResp, err := yaml.Marshal(resp.PhysicalInterfaces)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlResp, &newData); err != nil {
			return err
		}

		found := false
		if caller == "INIT PHYSICAL INTERFACES" {
			config.PhysicalInterfaces = newData
			found = true
		} else {
			for i, physical := range config.PhysicalInterfaces {
				if physical.Name == interfaceName {
					found = true
					config.PhysicalInterfaces[i] = newData[0]
					break
				}
			}
		}

		if !found || config.PhysicalInterfaces == nil {
			config.PhysicalInterfaces = append(config.PhysicalInterfaces, newData...)
		}
	case model.VLAN_STRING:
		var newData []model.ListVlanInterface
		yamlResp, err := yaml.Marshal(resp.VLANInterfaces)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlResp, &newData); err != nil {
			return err
		}

		found := false
		for i, vlan := range config.VLANInterfaces {
			if vlan.VlanInterfaceName == interfaceName {
				found = true
				config.VLANInterfaces[i] = newData[0]
				break
			}
		}

		if !found || config.VLANInterfaces == nil {
			config.VLANInterfaces = append(config.VLANInterfaces, newData...)
		}
	case model.BRIDGE_STRING:
		var newData []model.ListBridgeInterface
		yamlResp, err := yaml.Marshal(resp.BridgeInterfaces)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlResp, &newData); err != nil {
			return err
		}

		found := false
		for i, bridge := range config.BridgeInterfaces {
			if bridge.BridgeInterfaceName == interfaceName {
				found = true
				config.BridgeInterfaces[i] = newData[0]
				break
			}
		}

		if !found || config.BridgeInterfaces == nil {
			config.BridgeInterfaces = append(config.BridgeInterfaces, newData...)
		}
	case model.BOND_STRING:
		var newData []model.ListBondDetails
		yamlResp, err := yaml.Marshal(resp.BondInterfaces)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlResp, &newData); err != nil {
			return err
		}

		found := false
		for i, bond := range config.BondInterfaces {
			if bond.BondInterfaceName == interfaceName {
				found = true
				config.BondInterfaces[i] = newData[0]
				break
			}
		}

		if !found || config.BondInterfaces == nil {
			config.BondInterfaces = append(config.BondInterfaces, newData...)
		}
	case model.HA_STRING:
		var newData model.Ha
		yamlResp, err := yaml.Marshal(resp.Ha)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlResp, &newData); err != nil {
			return err
		}
		config.Ha = newData
	}

	updatedYAMLData, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(extras.INTERFACE_CONFIG_FILE_NAME, updatedYAMLData, 0644); err != nil {
		return err
	}

	var updated model.Config
	if err := yaml.Unmarshal(updatedYAMLData, &updated); err != nil {
		fmt.Println("error unmarshalling updatedYAMLData to resp")
	}

	fmt.Println("Config file updated successfully")
	return nil
}

func FetchConfig(interfaceType string) (model.Config, error) {

	resp := model.Config{}

	yamlData, err := readConfigFile(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		return resp, err
	}

	var config model.Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return resp, err
	}

	switch interfaceType {
	case "ALL":
		resp = config
	case model.DEVICE:
		resp.PhysicalInterfaces = config.PhysicalInterfaces
	case model.VLAN_STRING:
		resp.VLANInterfaces = config.VLANInterfaces
	case model.BRIDGE_STRING:
		resp.BridgeInterfaces = config.BridgeInterfaces
	case model.BOND_STRING:
		resp.BondInterfaces = config.BondInterfaces
	case model.HA_STRING:
		resp.Ha = config.Ha
	}

	return resp, nil
}

func UpdateConfigSpecificFields(interfaceType string, interfaceName string, fields model.ConfigSpecificFields, caller string) error {

	if ok := mu.TryLock(LOCK_TIME_OUT); !ok {
		return fmt.Errorf("Timeout while acquiring lock for %s - %s", caller, interfaceName)
	} else {
		fmt.Printf("Lock acquired for %s - %s\n", caller, interfaceName)
	}
	defer func() {
		mu.Unlock()
		fmt.Printf("Lock released for %s - %s\n", caller, interfaceName)
	}()

	yamlData, err := readConfigFile(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		return err
	}

	var config model.Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return err
	}

	switch interfaceType {
	case model.DEVICE:
		for i, physical := range config.PhysicalInterfaces {
			if physical.Name == interfaceName {
				config.PhysicalInterfaces[i].ServingLocation = fields.ServingLocation
				config.PhysicalInterfaces[i].DomainName = fields.DomainName
				config.PhysicalInterfaces[i].NetworkZone = fields.NetworkZone
				config.PhysicalInterfaces[i].IsDisabled = fields.IsDisabled
			}
		}
	case model.VLAN_STRING:
		for i, vlan := range config.VLANInterfaces {
			if vlan.VlanInterfaceName == interfaceName {
				config.VLANInterfaces[i].ServingLocation = fields.ServingLocation
				config.VLANInterfaces[i].DomainName = fields.DomainName
				config.VLANInterfaces[i].NetworkZone = fields.NetworkZone
				config.VLANInterfaces[i].IsDisabled = fields.IsDisabled
			}
		}
	case model.BRIDGE_STRING:
		for i, bridge := range config.BridgeInterfaces {
			if bridge.BridgeInterfaceName == interfaceName {
				config.BridgeInterfaces[i].ServingLocation = fields.ServingLocation
				config.BridgeInterfaces[i].DomainName = fields.DomainName
				config.BridgeInterfaces[i].NetworkZone = fields.NetworkZone
				config.BridgeInterfaces[i].IsDisabled = fields.IsDisabled
			}
		}
	case model.BOND_STRING:
		for i, bond := range config.BondInterfaces {
			if bond.BondInterfaceName == interfaceName {
				config.BondInterfaces[i].ServingLocation = fields.ServingLocation
				config.BondInterfaces[i].DomainName = fields.DomainName
				config.BondInterfaces[i].NetworkZone = fields.NetworkZone
				config.BondInterfaces[i].IsDisabled = fields.IsDisabled
			}
		}
	}

	updatedYAMLData, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(extras.INTERFACE_CONFIG_FILE_NAME, updatedYAMLData, 0644); err != nil {
		return err
	}

	fmt.Println("Config file updated successfully")
	return nil
}

func DeleteConfig(interfaceType string, interfaceName string, caller string) error {

	if ok := mu.TryLock(LOCK_TIME_OUT); !ok {
		return fmt.Errorf("Timeout while acquiring lock for %s - %s", caller, interfaceName)
	} else {
		fmt.Printf("Lock acquired for %s - %s\n", caller, interfaceName)
	}
	defer func() {
		mu.Unlock()
		fmt.Printf("Lock released for %s - %s\n", caller, interfaceName)
	}()

	yamlData, err := readConfigFile(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		return err
	}

	var config model.Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return err
	}

	switch interfaceType {
	case model.VLAN_STRING:
		for i, vlan := range config.VLANInterfaces {
			if vlan.VlanInterfaceName == interfaceName {
				if i < len(config.VLANInterfaces)-1 {
					config.VLANInterfaces = append(config.VLANInterfaces[:i], config.VLANInterfaces[i+1:]...)
				} else {
					config.VLANInterfaces = config.VLANInterfaces[:i]
				}
				break
			}
		}
	case model.BRIDGE_STRING:
		for i, bridge := range config.BridgeInterfaces {
			if bridge.BridgeInterfaceName == interfaceName {
				if i < len(config.BridgeInterfaces)-1 {
					config.BridgeInterfaces = append(config.BridgeInterfaces[:i], config.BridgeInterfaces[i+1:]...)
				} else {
					config.BridgeInterfaces = config.BridgeInterfaces[:i]
				}
				break
			}
		}
	case model.BOND_STRING:
		for i, bond := range config.BondInterfaces {
			if bond.BondInterfaceName == interfaceName {
				if i < len(config.BondInterfaces)-1 {
					config.BondInterfaces = append(config.BondInterfaces[:i], config.BondInterfaces[i+1:]...)
				} else {
					config.BondInterfaces = config.BondInterfaces[:i]
				}
				break
			}
		}
	case model.HA_STRING:
		config.Ha = model.Ha{}
	}

	updatedYAMLData, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(extras.INTERFACE_CONFIG_FILE_NAME, updatedYAMLData, 0644); err != nil {
		return err
	}

	fmt.Println("Config file updated successfully")
	return nil
}

func FetchConfigSpecificFields(interfaceType string) (model.ConfigSpecificFieldsMap, error) {

	resp := make(model.ConfigSpecificFieldsMap)

	yamlData, err := readConfigFile(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		return resp, err
	}

	var config model.Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return resp, err
	}

	switch interfaceType {
	case model.DEVICE:
		for _, physical := range config.PhysicalInterfaces {
			resp[physical.Name] = model.ConfigSpecificFields{
				ServingLocation: physical.ServingLocation,
				DomainName:      physical.DomainName,
				NetworkZone:     physical.NetworkZone,
				IsDisabled:      physical.IsDisabled,
			}
		}
	case model.VLAN_STRING:
		for _, vlan := range config.VLANInterfaces {
			resp[vlan.VlanInterfaceName] = model.ConfigSpecificFields{
				ServingLocation: vlan.ServingLocation,
				DomainName:      vlan.DomainName,
				NetworkZone:     vlan.NetworkZone,
			}
		}
	case model.BRIDGE_STRING:
		for _, bridge := range config.BridgeInterfaces {
			resp[bridge.BridgeInterfaceName] = model.ConfigSpecificFields{
				ServingLocation: bridge.ServingLocation,
				DomainName:      bridge.DomainName,
				NetworkZone:     bridge.NetworkZone,
			}
		}
	case model.BOND_STRING:
		for _, bond := range config.BondInterfaces {
			resp[bond.BondInterfaceName] = model.ConfigSpecificFields{
				ServingLocation: bond.ServingLocation,
				DomainName:      bond.DomainName,
				NetworkZone:     bond.NetworkZone,
			}
		}
	}

	return resp, nil
}

func CheckIfInterfaceTypeIsEmpty(interfaceType string) bool {

	yamlData, err := readConfigFile(extras.INTERFACE_CONFIG_FILE_NAME)
	if err != nil {
		return false
	}

	var config model.Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return false
	}

	switch interfaceType {
	case model.DEVICE:
		if len(config.PhysicalInterfaces) == 0 {
			return true
		}
	case model.VLAN_STRING:
		if len(config.VLANInterfaces) == 0 {
			return true
		}
	case model.BRIDGE_STRING:
		if len(config.BridgeInterfaces) == 0 {
			return true
		}
	case model.BOND_STRING:
		if len(config.BondInterfaces) == 0 {
			return true
		}
	}

	return false
}

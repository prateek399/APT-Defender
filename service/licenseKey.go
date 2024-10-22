package service

import (
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"anti-apt-backend/logger"
	"anti-apt-backend/model"
	"anti-apt-backend/util"
	"anti-apt-backend/util/interface_utils"
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func CreateLicenseKey(applianceKey string, email string) model.APIResponse {
	var resp model.APIResponse
	var licenseKey model.KeysTable
	var activated bool = false
	applianceKey = strings.TrimSpace(applianceKey)

	Model := ""
	serial := ""
	file, err := os.Open(extras.ROOT_DATA_DEVICE_CONFIG)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			if strings.Contains(line, "serial_no") {
				serial = parts[1]
			}
			if strings.Contains(line, "model_no") {
				Model = parts[1]
			}
		}
	}

	if strings.HasPrefix(applianceKey, "hcwj") {
		decryptedKey, err := util.Decrypt(strings.TrimPrefix(applianceKey, "hcwj"))
		if err != nil {
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERROR_IN_DECRYPTING_KEY, err)
		}
		keyData := strings.Split(string(decryptedKey), ";")
		if len(keyData) != 5 {
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_LICENSE_KEY, extras.ErrInvalidLicenseKey)
		}
		serialNo := keyData[0]
		date := keyData[1]
		macAddr := strings.ToLower(keyData[2])
		license := keyData[3]
		modelName := keyData[4]

		licenseKey.ApplianceKey = license
		licenseKey.DeviceSerialId = serialNo
		licenseKey.ModelNo = modelName

		lKey, err := dao.FetchLicenseKeyProfile()
		if err != nil {
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_IN_FETCHING_DATA, err)
		}

		licenseKey.ShippingDate = lKey.ShippingDate
		licenseKey.RegisteredEmail = lKey.RegisteredEmail

		if !util.IsInPermanentInterfaces(macAddr) {
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_LICENSE_KEY, extras.ErrInvalidLicenseKey)
		}

		if strings.Contains(date, "months") {
			date = strings.Replace(date, " months", "", -1)
			val, err := strconv.Atoi(date)
			if err != nil {
				return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_LICENSE_KEY, extras.ErrInvalidLicenseKey)
			}

			licenseKey.ExpiryTime = licenseKey.ExpiryTime.AddDate(0, val, 0).Local()
		} else if strings.Contains(date, "days") {
			date = strings.Replace(date, " days", "", -1)
			val, err := strconv.Atoi(date)
			if err != nil {
				return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_LICENSE_KEY, extras.ErrInvalidLicenseKey)
			}

			licenseKey.ExpiryTime = licenseKey.ExpiryTime.AddDate(0, 0, val).Local()
		}

		activated = true
	}

	if !activated {
		// Retrieve device serial number
		if applianceKey == strings.TrimSpace(extras.LICENSEKEY_ONE_YEAR) {
			licenseKey.ShippingDate = time.Now().Local()
			licenseKey.ExpiryTime = time.Now().AddDate(1, 0, 0).Local()
			licenseKey.RegisteredEmail = email
			licenseKey.ApplianceKey = extras.LICENSEKEY_ONE_YEAR
			licenseKey.DeviceSerialId = serial
			licenseKey.ModelNo = Model

		} else if applianceKey == strings.TrimSpace(extras.LICENSEKEY_THREE_YEAR) {
			licenseKey.ShippingDate = time.Now().Local()
			licenseKey.ExpiryTime = time.Now().AddDate(3, 0, 0).Local()
			licenseKey.RegisteredEmail = email
			licenseKey.ApplianceKey = extras.LICENSEKEY_THREE_YEAR
			licenseKey.DeviceSerialId = serial
			licenseKey.ModelNo = Model
		} else if applianceKey == strings.TrimSpace(extras.LICENSEKEY_FIVE_YEAR) {
			licenseKey.ShippingDate = time.Now().Local()
			licenseKey.ExpiryTime = time.Now().AddDate(5, 0, 0).Local()
			licenseKey.RegisteredEmail = email
			licenseKey.ApplianceKey = extras.LICENSEKEY_FIVE_YEAR
			licenseKey.DeviceSerialId = serial
			licenseKey.ModelNo = Model

		} else {
			return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_INVALID_LICENSE_KEY, extras.ErrInvalidLicenseKey)
		}
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, licenseKey)
	return resp
}

func ApplyDeviceConfigFile(request model.DeviceConfigFile) model.APIResponse {
	var resp model.APIResponse

	req := interface_utils.TrimStringsInStruct(request).(model.DeviceConfigFile)

	err := validateApplyDeviceConfigRequest(req)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_CLIENT_SIDE, err)
	}

	err = writeDeviceConfigToFile(req)
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	err = resetSystemToFactoryDefaults()
	if err != nil {
		return model.NewErrorResponse(http.StatusBadRequest, extras.ERR_FROM_SERVER_SIDE, err)
	}

	resp = model.NewSuccessResponse(extras.ERR_SUCCESS, "Device config applied successfully.")
	return resp
}

func writeDeviceConfigToFile(deviceConfig model.DeviceConfigFile) error {

	var config strings.Builder
	config.WriteString("device_name = " + deviceConfig.DeviceName + "\n")
	config.WriteString("serial_no = " + deviceConfig.DeviceID + "\n")
	config.WriteString("model_name = " + deviceConfig.ModelName + "\n")
	config.WriteString("mac_address = " + deviceConfig.DeviceMAC + "\n")
	config.WriteString("license_1year = " + deviceConfig.ApplianceKey1yr + "\n")
	config.WriteString("license_3year = " + deviceConfig.ApplianceKey3yr + "\n")
	config.WriteString("license_5year = " + deviceConfig.ApplianceKey5yr + "\n")
	config.WriteString("org_name = " + deviceConfig.OrganizationName + "\n")

	// configData := "device_name =" + deviceConfig.DeviceName + "\n" + "serial_no = " + deviceConfig.DeviceID + "\n" + "model_name = " + deviceConfig.ModelName + "\n" + "mac_address = " + deviceConfig.DeviceMAC + "\n" + "license_1year = " + deviceConfig.ApplianceKey1yr + "\n" + "license_3year = " + deviceConfig.ApplianceKey3yr + "\n" + "license_5year = " + deviceConfig.ApplianceKey5yr + "\n" + "org_name = " + deviceConfig.OrganizationName + "\n"

	err := os.WriteFile(extras.ROOT_DATA_DEVICE_CONFIG, []byte(config.String()), 0777)
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("Error writing device config to file: "+err.Error()))
		return err
	}

	return nil
}

func resetSystemToFactoryDefaults() error {

	resp := EraseConfig(extras.INITIAL_SETUP, "IT Team")
	if resp.StatusCode != http.StatusOK {
		logger.LoggerFunc("error", logger.LoggerMessage("Error while factory default settings"))
		return fmt.Errorf("err : %s", resp.Error)
	}

	return nil
}

func validateApplyDeviceConfigRequest(req model.DeviceConfigFile) error {
	if req.DeviceName == "" || req.DeviceID == "" || req.ModelName == "" || req.DeviceMAC == "" || req.ApplianceKey1yr == "" || req.ApplianceKey3yr == "" || req.ApplianceKey5yr == "" || req.OrganizationName == "" {
		return fmt.Errorf("all fields are required")
	}

	return nil
}

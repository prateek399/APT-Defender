package cuckoo

import (
	"anti-apt-backend/extras"
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var ErrNotAuthorized = fmt.Errorf("not authorized")

var API_KEY string

func readAPIKeyFromFile() (string, error) {
	filePath := extras.CUCKOO_CONF_FILE_PATH
	file, err := os.Open(filePath)
	if err != nil {
		return extras.EMPTY_STRING, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) >= 2 {
			if strings.Contains(line, "api_token") {
				apiToken := strings.TrimSpace(parts[1])
				return apiToken, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return extras.EMPTY_STRING, err
	}

	return extras.EMPTY_STRING, fmt.Errorf("API token not found in the configuration file")
}

func (c *Client) MakeRequest(req *http.Request) (*http.Response, error) {

	if API_KEY == extras.EMPTY_STRING {
		// fmt.Println("Setting API key from config file")
		apiKey, _ := readAPIKeyFromFile()
		// if err != nil {
		// 	// fmt.Println("error in reading API key from file: ", err)
		// 	return nil, err
		// }
		API_KEY = apiKey
	}

	// if BASE_URL == extras.EMPTY_STRING {
	// 	fmt.Println("Setting Base URL from config file")
	// 	baseUrl, err := readBaseURLFromFile()
	// 	if err != nil {
	// 		fmt.Println("Failed to set base URL from conf file: ", err)
	// 		return nil, err
	// 	}
	// 	BASE_URL = baseUrl
	// }
	// log.Println("API KEY: ", API_KEY)

	if API_KEY == extras.EMPTY_STRING {
		// fmt.Println("API key not found")
		return nil, fmt.Errorf("API key not found")
	}

	if BASE_URL == extras.EMPTY_STRING {
		fmt.Println("Base URL not found")
		return nil, fmt.Errorf("base URL not found")
	}

	c.APIKey = API_KEY
	c.BaseURL = BASE_URL

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", API_KEY))

	resp, err := c.Client.Do(req)
	if err != nil {
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:couldn't make request"))
		// fmt.Println("error in making request: ", err)
		return nil, err
	}

	if resp.StatusCode == 401 {
		// fmt.Println("not authorized")
		// logger.LoggerFunc("error", logger.LoggerMessage("taskLog:not authorized"))
		return resp, ErrNotAuthorized
	}

	return resp, err
}

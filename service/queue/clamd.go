package queues

import (
	"anti-apt-backend/extras"
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dutchcoders/go-clamd"
)

const (
	RES_OK          = "OK"
	RES_FOUND       = "FOUND"
	RES_ERROR       = "ERROR"
	RES_PARSE_ERROR = "PARSE ERROR"
)

func extensionsToIgnore() (map[string]struct{}, error) {
	ignoreExtensions := make(map[string]struct{})
	ignoreFilePath := extras.IGNORE_EXTENSIONS_FILE_PATH
	file, err := os.Open(ignoreFilePath)
	if err != nil {
		return ignoreExtensions, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ext := strings.TrimSpace(scanner.Text())
		if ext != "" {
			ignoreExtensions[ext] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return ignoreExtensions, err
	}

	return ignoreExtensions, nil
}

func getClamdAddress() string {

	// address := "/opt/local/var/run/clamav/clamd.socket"
	// if _, err := os.Stat(address); os.IsNotExist(err) {
	return "/var/run/clamav/clamd.ctl"
	// }
	// if _, err := os.Stat(address); os.IsNotExist(err) {
	// 	address = "/run/clamav/clamd.ctl"
	// }
	// return address
}

func scanResultThroughClamd(clamd *clamd.Clamd, filePath string) (bool, error) {
	// slog.Println("Scanning file: ", filePath)
	response, err := clamd.ScanFile(filePath)
	if err != nil {
		return false, err
	}

	// slog.Println("Response: ", response)

	for s := range response {
		if s.Status == RES_FOUND {
			// slog.Println("Found malicious file: ", filePath)
			return true, nil
		}
		// slog.Println("Status: ", s.Description, s.Status, s.Hash, s.Path, s.Raw, s.Size)
	}
	return false, nil

}

func ScanFileThroughClamd(task Task) (bool, error) {

	clamdAddress := getClamdAddress()
	clamd := clamd.NewClamd(clamdAddress)

	// ignoreExtensions, _ := extensionsToIgnore()

	filePath := fmt.Sprintf(extras.SANDBOX_FILE_PATHS+"%d", task.Id)

	response, err := clamd.ScanFile(filePath)
	if err != nil {
		return false, err
	}

	// slog.Println("Response: ", response)

	for s := range response {
		if s.Status == RES_FOUND {
			// slog.Println("Found malicious file: ", filePath)
			return true, nil
		}
		// slog.Println("Status: ", s.Description, s.Status, s.Hash, s.Path, s.Raw, s.Size)
	}
	return false, nil

}

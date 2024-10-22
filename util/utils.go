package util

import (
	"anti-apt-backend/extras"
	"anti-apt-backend/model"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/ttacon/libphonenumber"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func AuthenticateToken(bearer_token []string, access_token string) error {
	if len(bearer_token) == 2 && bearer_token[1] == access_token {
		return nil
	}

	return fmt.Errorf("invalid access token found in request header")
}

func CompareDataType(x, y interface{}) error {
	typeOfX := reflect.TypeOf(x)
	typeOfY := reflect.TypeOf(y)

	if typeOfX.Kind() != typeOfY.Kind() {
		return fmt.Errorf("data type of %v (%v) and %v (%v) is different", x, typeOfX, y, typeOfY)
	}

	return nil
}

func TrimString(name string) string {
	arr := strings.Split(name, " ")
	name = ""
	for _, val := range arr {
		val = strings.TrimSpace(val)
		if len(val) > 0 {
			name = name + " " + val
		}
	}
	name = strings.TrimSpace(name)
	return name
}

func IsValidName(name string) bool {
	if len(name) == 0 {
		return false
	}
	validName := regexp.MustCompile(`^[A-Za-z0-9_][-A-Za-z0-9_]{2,35}$`)
	return validName.MatchString(name)
}

func IsValidEmail(email string) bool {
	validEmailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,12}$`)
	return validEmailPattern.MatchString(email)
}

func GetLocalIP() string {
	var emptyString string = ""
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return emptyString
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return emptyString
}

func ValidateLicenseKey(key string) error {
	for _, subPartofKey := range strings.Split(key, "-") {
		if len(subPartofKey) != 5 {
			return fmt.Errorf("invalid license key")
		}
	}

	return nil
}

func IsValidIp(ip string) bool {
	validIp := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	return validIp.MatchString(ip)
}

func IsValidPhone(phoneString string, countryCode interface{}) bool {
	var countrycode string
	switch v := countryCode.(type) {
	case int:
		countrycode = libphonenumber.GetRegionCodeForCountryCode(v)
	case string:
		countrycode = v
	}

	parsedNumber, err := libphonenumber.Parse(phoneString, countrycode)

	if err != nil {
		fmt.Println("Error in parsing the Phone number")
		return false
	}

	return libphonenumber.IsValidNumberForRegion(parsedNumber, countrycode)
}

func IsValidCountryCode(countryCode interface{}) bool {
	switch v := countryCode.(type) {
	case int:
		var regionCodes []string = libphonenumber.CountryCodeToRegion[v]
		if len(regionCodes) == 0 {
			return false
		}
	case string:
		if libphonenumber.GetCountryCodeForRegion(v) == 0 {
			return false
		}
	}
	return true
}

func GetVerdict(score float32) model.Verdict {
	switch {
	case score == 0:
		return model.Clean
	case score > 0 && score <= 3:
		return model.LowRisk
	case score > 3 && score <= 5:
		return model.MediumRisk
	case score > 5 && score <= 7:
		return model.HighRisk
	case score > 7:
		return model.Critical
	}
	return model.Unknown
}

func CheckForLockedVMs(machines []*model.Machine) map[string]*model.Machine {
	machinesMap := make(map[string]*model.Machine)
	for _, machine := range machines {
		machinesMap[machine.Name] = machine
	}
	return machinesMap
}

func ValidateContentType(bytes []byte, scanProfiles model.ScanProfile) string {
	// Check the file type
	kind, err := filetype.Match(bytes)
	if err != nil {
		fmt.Println("Unknown file type")
		return "unknown"
	}

	// fmt.Printf("File type: %s, MIME: %s\n", kind.Extension, kind.MIME.Value)
	return kind.Extension
}

func ValidateContentTypeOfFile(form *multipart.Form, scanProfiles model.ScanProfile) error {
	file, err := form.File["filename"][0].Open()
	if err != nil {
		return err
	}
	defer file.Close()

	ext := filepath.Ext(form.File["filename"][0].Filename)
	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	ext1 := ValidateContentType(buf.Bytes(), scanProfiles)

	var t = reflect.TypeOf(scanProfiles)
	var v = reflect.ValueOf(scanProfiles)
	var found = false
	for i := 0; i < t.NumField(); i++ {
		// fmt.Println("Ext: ", ext)
		if t.Field(i).Name != "UserAuthenticationKey" && v.Field(i).Bool() {
			if slices.Contains(extras.FileExt[t.Field(i).Name], ext) || slices.Contains(extras.FileExt[t.Field(i).Name], ext1) {
				found = true
				break
			}
		}
	}

	if !found {
		return extras.ErrInvalidContentType
	}

	return nil
}

func IsEmpty(value interface{}) bool {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		zero := reflect.Zero(v.Type())
		return reflect.DeepEqual(v.Interface(), zero.Interface())
	}
	return false
}

func Reverse(slice interface{}) interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		fmt.Println("Input is not a slice")
	}

	length := s.Len()
	result := reflect.MakeSlice(s.Type(), length, length)

	for i, j := 0, length-1; i < length; i, j = i+1, j-1 {
		result.Index(i).Set(s.Index(j))
	}

	return result.Interface()
}

func GetCpuInfo() (float64, error) {
	err := InstallCommandIfMissing(extras.FreeCmd)
	if err != nil {
		return 0, err
	}

	cmd := exec.Command("top", "-bn", "1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Error:", err)
		return 0, err
	}

	cpuUtilization := 0.0
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "%Cpu(s):") {
			fields := strings.Fields(line)
			idleTime, err := strconv.ParseFloat(fields[7], 64)
			if err != nil {
				log.Println("Error parsing CPU idle time:", err)
				return 0, err
			}
			cpuUtilization = 100 - idleTime
			break
		}
	}

	return cpuUtilization, nil
}

func GetRamInfo() (float64, error) {
	err := InstallCommandIfMissing(extras.FreeCmd)
	if err != nil {
		return 0, err
	}

	cmd := exec.Command("free", "-m")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Error:", err)
		return 0, err
	}

	ramUtilization := 0.0
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Mem:") {
			fields := strings.Fields(line)
			totalRam, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				log.Println("Error parsing total RAM:", err)
				return 0, err
			}
			availableRam, err := strconv.ParseFloat(fields[6], 64)
			if err != nil {
				log.Println("Error parsing available RAM:", err)
				return 0, err
			}
			ramUtilization = (totalRam - availableRam) / totalRam
			ramUtilization = ramUtilization * 100
			ramUtilization = math.Round(ramUtilization*100) / 100
			break
		}
	}

	return ramUtilization, nil
}

func GetSpaceInfo() (float64, error) {
	err := InstallCommandIfMissing(extras.DmiDecodeCmd)
	if err != nil {
		return 0, err
	}

	statfs := &syscall.Statfs_t{}
	err = syscall.Statfs("/", statfs)
	if err != nil {
		return 0, err
	}

	totalSpace := statfs.Blocks * uint64(statfs.Bsize)
	availableSpace := statfs.Bavail * uint64(statfs.Bsize)

	percentage := (float64(totalSpace-availableSpace) / float64(totalSpace))

	percentage = percentage * 100
	percentage = math.Round(percentage*100) / 100

	return percentage, nil
}

func InstallCommandIfMissing(command string) error {
	_, err := exec.LookPath(command)
	if err != nil {
		switch command {
		case extras.DmiDecodeCmd:
			return installDmiDecode()
		case extras.FreeCmd:
			return installFree()
		default:
			return errors.New("unknown command: " + command)
		}
	}
	return nil
}

func installFree() error {
	log.Println("Installing free...")
	cmd := exec.Command("apt-get", "install", "-y", "procps")
	err := cmd.Run()
	if err != nil {
		return errors.New("failed to install free: " + err.Error())
	}
	return nil
}

func installDmiDecode() error {
	log.Println("Installing dmidecode...")
	cmd := exec.Command("apt-get", "install", "-y", "dmidecode")
	err := cmd.Run()
	if err != nil {
		return errors.New("failed to install dmidecode: " + err.Error())
	}
	return nil
}

func QEscape(filename string) string {
	return url.QueryEscape(filename)
}

func QUnescape(filename string) string {
	name, err := url.QueryUnescape(filename)
	if err != nil {
		name = filename
	}
	return name
}

func CompareIPs(ip1, ip2 string) int {
	var parseIPPart = func(part string) int {
		num := 0
		for _, digit := range part {
			num = num*10 + int(digit-'0')
		}
		return num
	}

	ip1Parts := strings.Split(ip1, ".")
	ip2Parts := strings.Split(ip2, ".")

	for i := 0; i < 4; i++ {
		part1 := parseIPPart(ip1Parts[i])
		part2 := parseIPPart(ip2Parts[i])

		if part1 < part2 {
			return -1
		} else if part1 > part2 {
			return 1
		}
	}

	return 0
}

// func SortMap(data any, isFod bool) any {
// 	arr := []any{}

// 	if isFod {
// 		value := data.(map[string]model.FileOnDemand)
// 		for job, fod := range value {
// 			arr = append(arr, map[string]any{
// 				"jobid": job,
// 				"value": fod,
// 			})
// 		}
// 	} else {
// 		value := data.(map[string]model.UrlOnDemand)
// 		for job, uod := range value {
// 			arr = append(arr, map[string]any{
// 				"jobid": job,
// 				"value": uod,
// 			})
// 		}
// 	}

// 	sort.Slice(arr, func(i, j int) bool {
// 		var t1, t2 time.Time
// 		if val, ok := arr[i].(map[string]any)["value"].(model.FileOnDemand); ok {
// 			t1, _ = time.Parse(extras.TIME_FORMAT, val.SubmittedTime)
// 			t2, _ = time.Parse(extras.TIME_FORMAT, arr[i].(map[string]any)["value"].(model.FileOnDemand).SubmittedTime)
// 		} else if val, ok := arr[i].(map[string]any)["value"].(model.UrlOnDemand); ok {
// 			t1, _ = time.Parse(extras.TIME_FORMAT, val.SubmittedTime)
// 			t2, _ = time.Parse(extras.TIME_FORMAT, arr[i].(map[string]any)["value"].(model.UrlOnDemand).SubmittedTime)

// 		}

// 		return t2.After(t1)
// 	})

// 	if isFod {
// 		value := []model.FileOnDemand{}
// 		for _, val := range arr {
// 			v := val.(map[string]any)
// 			if fod, ok := v["value"].(model.FileOnDemand); ok {
// 				// t1, _ := time.Parse(extras.TIME_FORMAT, fod.SubmittedTime)
// 				// fod.SubmittedTime = t1.Format("2006-01-02 15:04:05")
// 				value = append(value, fod)
// 			}
// 		}
// 		return value

// 	} else {
// 		value := []model.UrlOnDemand{}
// 		for _, val := range arr {
// 			v := val.(map[string]any)
// 			if uod, ok := v["value"].(model.UrlOnDemand); ok {
// 				// t1, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", uod.SubmittedTime)
// 				// uod.SubmittedTime = t1.Format("2006-01-02 15:04:05")
// 				value = append(value, uod)
// 			}
// 		}
// 		return value
// 	}
// }

func Encrypt(plaintext []byte) (string, error) {
	// Define the key
	key, err := hex.DecodeString("33b04b7e103a0cd8b54763051cef082155abe02ffdebae5e1d417e2fdb1a11a3")
	if err != nil {
		return "", err
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Generate a random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// pkcs7Pad pads the given input byte slice according to the PKCS7 standard
	var pkcs7Pad = func(input []byte, blockSize int) []byte {
		padding := blockSize - len(input)%blockSize
		padText := bytes.Repeat([]byte{byte(padding)}, padding)
		return append(input, padText...)
	}
	// Apply PKCS7 padding to plaintext
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)

	// Create a cipher block mode
	mode := cipher.NewCBCEncrypter(block, iv)

	// Encrypt the plaintext
	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)

	// Combine IV and ciphertext
	encrypted := append(iv, ciphertext...)

	// Base64 encode the ciphertext
	ciphertextBase64 := base64.StdEncoding.EncodeToString(encrypted)

	return ciphertextBase64, nil
}

func Decrypt(ciphertextBase64 string) (string, error) {
	// Decode the base64 encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}

	// Define the key
	key, err := hex.DecodeString("33b04b7e103a0cd8b54763051cef082155abe02ffdebae5e1d417e2fdb1a11a3")
	if err != nil {
		return "", err
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Extract the IV from the ciphertext
	iv := ciphertext[:aes.BlockSize]

	// Decrypt the ciphertext (excluding the IV)
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(ciphertext)-aes.BlockSize)
	mode.CryptBlocks(decrypted, ciphertext[aes.BlockSize:])

	// pkcs7Unpad removes the padding from the given input byte slice according to the PKCS7 standard
	var pkcs7Unpad = func(input []byte) ([]byte, error) {
		if len(input) == 0 {
			return nil, errors.New("input is empty")
		}
		padding := int(input[len(input)-1])
		if padding == 0 || padding > len(input) {
			return nil, errors.New("invalid padding")
		}
		return input[:len(input)-padding], nil
	}

	// Remove padding from the decrypted plaintext
	decrypted, err = pkcs7Unpad(decrypted)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func IsInPermanentInterfaces(macAddr string) bool {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	for _, iface := range interfaces {
		// Skip loopback and virtual interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp != 0 {
			continue
		}

		// Convert MAC address to lowercase and remove delimiters
		ifaceMAC := strings.ToLower(strings.ReplaceAll(iface.HardwareAddr.String(), ":", ""))
		macAddr = strings.ToLower(strings.ReplaceAll(macAddr, ":", ""))

		if ifaceMAC == macAddr {
			return true
		}
	}

	return false
}

func FormatWithOrdinal(t time.Time) string {
	day := t.Day()
	suffix := "th"
	switch day {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	}

	return fmt.Sprintf("%d%s %s %d", day, suffix, t.Month().String(), t.Year())
}

func CalculateHash(multipartForm *multipart.Form, hashType string) (string, error) {

	file, err := multipartForm.File["filename"][0].Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	var hash hash.Hash

	switch hashType {
	case "md5":
		hash = md5.New()
	case "sha1":
		hash = sha1.New()
	case "sha256":
		hash = sha256.New()
	default:
		return "", fmt.Errorf("unsupported hash type")
	}

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func EmbedTextInHTML(textFilePath string) {
	content, _ := os.ReadFile(textFilePath)

	os.Remove(textFilePath)

	// Create a new template and parse the HTML template into it
	tmpl, _ := template.New("page").Parse(extras.HTMLTemplate)

	// Populate the template with the text file content
	type PageData struct {
		Content string
	}
	data := PageData{Content: string(content)}

	// Write the output HTML to a file
	outputFile, _ := os.Create(textFilePath)
	defer outputFile.Close()

	tmpl.Execute(outputFile, data)
}

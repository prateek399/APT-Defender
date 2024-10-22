package hash

import (
	"anti-apt-backend/extras"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"slices"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type BlockedHashes struct {
	Hashes map[string]bool `yaml:"hashes"`
}

type UrlCache struct {
	Verdict int
	Time    time.Time
}
type Urls struct {
	Cache map[string]UrlCache `yaml:"url_cache"`
}

type AllowedHashes struct {
	Hashes map[string]bool `yaml:"hashes"`
}

var urlCaches Urls
var fileQueue []string
var blockedHashes BlockedHashes

// var writeMutex sync.RWMutex

var allowedHashes AllowedHashes

var fileQueueReadLock sync.RWMutex

func IsFileBeingProcessed(filePath string) bool {
	fileQueueReadLock.Lock()
	defer fileQueueReadLock.Unlock()
	hashValue, _ := CalculateHash(filePath, "md5")
	return slices.Contains(fileQueue, hashValue)
}

var fileLock sync.RWMutex

func AddFileToQueue(filePath string) {
	fileLock.Lock()
	defer fileLock.Unlock()
	hashValue, _ := CalculateHash(filePath, "md5")
	fileQueue = append(fileQueue, hashValue)
}

func RemoveFileFromQueue(filePath string) {
	fileLock.Lock()
	defer fileLock.Unlock()
	hashValue, _ := CalculateHash(filePath, "md5")
	for i := range fileQueue {
		if fileQueue[i] == hashValue {
			if i+1 < len(fileQueue) {
				fileQueue = append(fileQueue[:i], fileQueue[i+1:]...)
			} else {
				fileQueue = fileQueue[:i]
			}
			break
		}
	}
}

var urlLock sync.RWMutex

func GetUrlCaches(url string) (bool, UrlCache) {
	urlLock.Lock()
	defer urlLock.Unlock()

	if _, ok := urlCaches.Cache[url]; !ok {
		return false, UrlCache{}
	}

	if !urlCaches.Cache[url].Time.Add(24*time.Hour).Before(time.Now()) && urlCaches.Cache[url].Verdict == extras.ANALYSING {
		return false, UrlCache{}
	}

	return true, urlCaches.Cache[url]
}

var cleanhashLock sync.RWMutex
var blockhashLock sync.RWMutex

func IsMalwareHash(hash string) (bool, error) {
	blockhashLock.Lock()
	defer blockhashLock.Unlock()

	value, present := blockedHashes.Hashes[hash]
	return present && value, nil
}

func IsCleanHash(hash string) (bool, error) {
	cleanhashLock.Lock()
	defer cleanhashLock.Unlock()

	value, present := allowedHashes.Hashes[hash]
	return present && value, nil
}

const (
	URL_CACHE_FILE    = extras.DATABASE_PATH + "url_cache.yaml"
	BLOCKED_HASH_FILE = extras.DATABASE_PATH + "blocked_hashes.yaml"
	ALLOWED_HASH_FILE = extras.DATABASE_PATH + "allowed_hashes.yaml"
)

func InitHashes() {

	if _, err := os.Stat(BLOCKED_HASH_FILE); os.IsNotExist(err) {
		if _, err := os.Create(BLOCKED_HASH_FILE); err != nil {
			log.Println("Error creating blocked hashes file: " + err.Error())
		}
	}
	if _, err := os.Stat(ALLOWED_HASH_FILE); os.IsNotExist(err) {
		if _, err := os.Create(ALLOWED_HASH_FILE); err != nil {
			log.Println("Error creating allowed hashes file: " + err.Error())
		}
	}

	yamlFile, err := os.ReadFile(BLOCKED_HASH_FILE)
	if err != nil {
		log.Println("Error reading blocked hashes file: " + err.Error())
	}

	blockedHashes = BlockedHashes{Hashes: make(map[string]bool)}
	err = yaml.Unmarshal(yamlFile, &blockedHashes)
	if err != nil {
		log.Println("Error reading blocked hashes file: " + err.Error())
	}

	yamlFile, err = os.ReadFile(ALLOWED_HASH_FILE)
	if err != nil {
		log.Println("Error reading allowed hashes file: " + err.Error())
	}

	allowedHashes = AllowedHashes{Hashes: make(map[string]bool)}
	err = yaml.Unmarshal(yamlFile, &allowedHashes)
	if err != nil {
		log.Println("Error reading allowed hashes file: " + err.Error())
	}
}

func InitUrlCache() {
	if _, err := os.Stat(URL_CACHE_FILE); os.IsNotExist(err) {
		if _, err := os.Create(URL_CACHE_FILE); err != nil {
			log.Println("Error creating url cache file: " + err.Error())
		}
	}
	yamlFile, err := os.ReadFile(URL_CACHE_FILE)
	if err != nil {
		log.Println("Error reading url cache file: " + err.Error())
	}

	urlCaches = Urls{Cache: make(map[string]UrlCache)}

	err = yaml.Unmarshal(yamlFile, &urlCaches)
	if err != nil {
		log.Println("Error reading url cache file: " + err.Error())
	}

	// fmt.Println("URL Caches: ", urlCaches)
}

func CalculateHash(filepath string, hashType string) (string, error) {
	file, err := os.Open(filepath)
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

func IsAllowedHash(filepath string) bool {

	if filepath == extras.EMPTY_STRING {
		// slog.Println("filepath is required")
	}

	hashValue, _ := CalculateHash(filepath, "md5")
	isAllowed, err := IsCleanHash(hashValue)
	if err != nil {
		return false
	}

	if isAllowed {
		return true
	} else {
		// slog.Println("No allowed hash detected")
	}
	return false

}

func IsBlockedHash(filepath string) bool {

	if filepath == extras.EMPTY_STRING {
		// slog.Println("filepath is required for checking in blocked hashes")
	}

	var hashValue string
	if filepath != "" {
		hashValue, _ = CalculateHash(filepath, "md5")
	}

	var matchedHashType string

	hashTypes := []string{"md5"}

	for _, hashType := range hashTypes {
		isMalware, err := IsMalwareHash(hashValue)
		if err != nil {
			continue
		}

		if isMalware {
			matchedHashType = hashType
		}
	}

	if matchedHashType != "" {
		return true
	} else {
		// slog.Println("No malware hash detected")
	}
	return false
}

// var hashWriteLock sync.RWMutex

func SaveVerdict(hash string, verdict string) error {

	if hash == extras.EMPTY_STRING {
		return nil
	}

	if verdict != extras.BLOCK && verdict != extras.ALLOW {
		return fmt.Errorf("invalid verdict")
	}

	finalVerdict := verdict == extras.ALLOW

	if blockedHashes.Hashes == nil {
		blockedHashes.Hashes = make(map[string]bool)
	}
	if allowedHashes.Hashes == nil {
		allowedHashes.Hashes = make(map[string]bool)
	}

	blockhashLock.Lock()
	blockedHashes.Hashes[hash] = !finalVerdict
	blockhashLock.Unlock()

	cleanhashLock.Lock()
	allowedHashes.Hashes[hash] = finalVerdict
	cleanhashLock.Unlock()

	WriteHashesFile(BLOCKED_HASH_FILE, blockedHashes)
	WriteHashesFile(ALLOWED_HASH_FILE, allowedHashes)

	return nil
}

// var urlCacheMutex sync.RWMutex

func ReplaceUrlCache(url string, status int) {
	urlLock.Lock()
	defer urlLock.Unlock()

	if urlCaches.Cache == nil {
		urlCaches.Cache = make(map[string]UrlCache)
	}

	urlCaches.Cache[url] = UrlCache{
		Verdict: status,
		Time:    time.Now(),
	}

	// go func() {
	// 	retryCount := 3
	// 	for retryCount > 0 {
	WriteHashesFile(URL_CACHE_FILE, urlCaches)
	// if err != nil {
	// 	fmt.Println("hash write file error: ", err)
	// }
	// 		retryCount--
	// 	}
	// }()

}

func WriteHashesFile(filePath string, data interface{}) error {
	// writeMutex.Lock()
	// defer writeMutex.Unlock()

	yamlData, err := yaml.Marshal(&data)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_, err = os.Create(filePath)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(filePath, yamlData, 0644)
	if err != nil {
		return err
	}

	return nil
}

package logger

import (
	"anti-apt-backend/dao"
	"anti-apt-backend/extras"
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/gookit/slog"
	"github.com/sirupsen/logrus"
)

var logOnce sync.Once
var logger *logrus.Logger

func LoggerFunc(LevelType string, v ...interface{}) {
	viewLog, err := dao.FetchViewLogs()
	if err != nil {
		fmt.Println(err)
		return
	}
	if !viewLog.Enabled {
		return
	}
	if len(v) == 2 && (v[1] == nil || !v[1].(bool)) {
		return
	}

	if len(v) == 2 {
		if v[1].(bool) {
			logger = setupLogger()
			for _, value := range v {
				switch LevelType {
				case "debug":
					logger.Debug(value)

				case "info":
					logger.Info(value)

				case "warn":
					logger.Warn(value)

				case "error":
					logger.Error(value)

				case "fatal":
					logger.Fatal(value)

				case "panic":
					logger.Panic(value)

				}
			}
		}
	} else {
		logger = setupLogger()
		for _, value := range v {
			switch LevelType {
			case "debug":
				logger.Debug(value)

			case "info":
				logger.Info(value)

			case "warn":
				logger.Warn(value)

			case "error":
				logger.Error(value)

			case "fatal":
				logger.Fatal(value)

			case "panic":
				logger.Panic(value)

			}
		}
	}
}

func LoggerMessage(message interface{}) error {
	// _, file, line, _ := runtime.Caller(0)
	// file = file + ":" + fmt.Sprintf("%d", line) + ": " + message
	// return file
	// pc, file, line, _ := runtime.Caller(1)

	return fmt.Errorf("%v", message)
}

func setupLogger() *logrus.Logger {
	logger_Logrus := logrus.New()
	logger_Logrus.SetLevel(logrus.DebugLevel)
	logger_Logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger_Logrus.SetOutput(getFile())

	return logger_Logrus
}

var LoggingEnabled bool

func IsLoggingEnabled() bool {
	file, err := os.Create(extras.LOGGING_ENABLED_FLAG_PATH)
	if err != nil {
		// error reading the file, assume logging is disabled
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "1" {
			return true
		}
	}

	return false
}

var mu sync.Mutex

func LogAccToTaskId(taskId int, message string) {

	// if !IsLoggingEnabled() {
	// 	return
	// }

	mu.Lock()
	defer mu.Unlock()

	_, filename, line, _ := runtime.Caller(1)

	fileName := fmt.Sprintf(extras.TASK_LOGS_PATH+"task_%d.log", taskId)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("ERROR WHILE OPENING FILE FOR TASK %d: %v", taskId, err)
		return
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	logger.Println(filename + ":" + fmt.Sprintf("%d", line) + ": " + message)
}

func UpdateLoggingEnabledFlag() error {

	file, err := os.Create(extras.LOGGING_ENABLED_FLAG_PATH)
	if err != nil {
		return err
	}
	defer file.Close()

	LoggingEnabled = !LoggingEnabled

	var flag string
	if LoggingEnabled {
		flag = "1"
	} else {
		flag = "0"
	}
	_, err = file.WriteString(flag)
	if err != nil {
		return err
	}

	return nil
}

func Print(format string, args ...interface{}) {
	if LoggingEnabled {
		slog.Printf(format, args...)
	}
}

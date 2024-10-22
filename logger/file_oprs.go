package logger

import (
	"anti-apt-backend/extras"
	"os"
	"path/filepath"
)

var logFile *os.File

func CloseLogFile() {
	logger.Out.(*os.File).Close()
}

func getFile() *os.File {
	logOnce.Do(openLogFile)
	return logFile
}

func openLogFile() {
	var err error
	logFile, err = os.OpenFile(filepath.Join(extras.DATABASE_PATH, "logFile.txt"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("Incorrect file path or file cannot be created!")
	}
}

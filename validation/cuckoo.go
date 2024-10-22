package validation

import (
	"anti-apt-backend/extras"
	"fmt"
	"os"
	"strings"
)

func ValidateFile(fps []string) error {

	const maxFileSize = 200 << 20
	for _, fp := range fps {

		fp = strings.TrimSpace(fp)

		if fp == extras.EMPTY_STRING {
			return fmt.Errorf("file path is required")
		}

		if _, err := os.Stat(fp); os.IsNotExist(err) {
			return extras.ErrFileNotFound
		}

		fileInfo, err := os.Stat(fp)
		if err != nil {
			return fmt.Errorf("error getting file info for the file: %s, %w", fp, err)
		}

		if fileInfo.Size() > maxFileSize {
			return fmt.Errorf("size of file %s exceeds maximum allowed size(200 MB)", fp)
		}
	}
	return nil
}

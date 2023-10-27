package packages

import (
	"fmt"
	"os"
)

// GetFileSize returns the size of a file in KB, or MB if > 1024 KB.
func GetFileSize(filename string) (string, error) {
	file, err := os.Stat(filename)
	if err != nil {
		return "", err
	}

	return SizeToString(file.Size()), nil
}

func SizeToString(size int64) string {
	// convert file size to KB
	fileSize := float64(size) / 1024
	sizeUnit := "KB"

	if fileSize > 1024 {
		// convert file size to MB if it's more than 1024 KB
		fileSize = fileSize / 1024
		sizeUnit = "MB"
	}

	return fmt.Sprintf("%.1f %s", fileSize, sizeUnit)
}

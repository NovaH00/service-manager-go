package manager

import (
	"fmt"
	"os"
	"path/filepath"
)

func openAppend(filePath string) (*os.File, error) {
	// Flags to open file in append mode, if file is not exist then create and open
	// in append mode
	parentDir := filepath.Dir(filePath)

	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("create parent dir: %w", err)
	}

	flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY

	return os.OpenFile(filePath, flags, 0644)
}

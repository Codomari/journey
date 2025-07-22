package helpers

import (
	"os"
	"path/filepath"
)

// GetFilenameWithoutExtension returns the filename from path without its extension.
func GetFilenameWithoutExtension(path string) string {
	return filepath.Base(path)[0 : len(filepath.Base(path))-len(filepath.Ext(path))]
}

// IsDirectory returns true if path is a directory, false otherwise.
func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// FileExists returns true if the file at path exists and can be accessed.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		// We could check with os.IsNotExist(err) here, but since os.Stat threw an error, we likely can't use the file anyway.
		return false
	}
	return true
}

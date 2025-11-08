package main

import (
	"fmt"
	"os"
)

// checkStorage verifies that the storage path is mounted and accessible
func checkStorage(path string) error {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("storage path does not exist: %s", path)
		}
		return fmt.Errorf("cannot access storage path: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("storage path is not a directory: %s", path)
	}

	// Check if it's a mountpoint (platform-specific, defined in storage_*.go)
	if !isMountpoint(path) {
		return fmt.Errorf("storage path is not mounted: %s", path)
	}

	// Check if writable
	testFile := path + "/.ctrlsrv-test"
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("storage path is not writable: %w", err)
	}
	os.Remove(testFile)

	return nil
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Platform-specific functions are defined in:
// - storage_linux.go (for Linux/Debian)
// These functions are implemented differently per platform:
// - getStorageUsage(path string) (used, free, total uint64, err error)
// - isMountpoint(path string) bool

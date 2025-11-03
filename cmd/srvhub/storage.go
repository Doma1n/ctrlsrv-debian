package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
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

	// Check if it's a mountpoint
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

// isMountpoint checks if a path is a mountpoint
func isMountpoint(path string) bool {
	// Method 1: Use mountpoint command if available
	cmd := exec.Command("mountpoint", "-q", path)
	if err := cmd.Run(); err == nil {
		return true
	}

	// Method 2: Check if parent has different device ID
	pathInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	parentPath := path + "/.."
	parentInfo, err := os.Stat(parentPath)
	if err != nil {
		return false
	}

	pathStat := pathInfo.Sys().(*syscall.Stat_t)
	parentStat := parentInfo.Sys().(*syscall.Stat_t)

	// If device IDs differ, it's a mountpoint
	return pathStat.Dev != parentStat.Dev
}

// getStorageUsage returns storage usage statistics
func getStorageUsage(path string) (used, free, total uint64, err error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, 0, err
	}

	total = stat.Blocks * uint64(stat.Bsize)
	free = stat.Bfree * uint64(stat.Bsize)
	used = total - free

	return used, free, total, nil
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

// checkStorageHealth performs comprehensive storage health check
func checkStorageHealth(path string) (healthy bool, issues []string) {
	healthy = true
	issues = []string{}

	// Check if mounted
	if err := checkStorage(path); err != nil {
		healthy = false
		issues = append(issues, fmt.Sprintf("Storage check failed: %v", err))
		return
	}

	// Check disk usage
	used, free, total, err := getStorageUsage(path)
	if err != nil {
		issues = append(issues, fmt.Sprintf("Cannot get disk usage: %v", err))
	} else {
		usedPct := float64(used) / float64(total) * 100
		if usedPct > 90 {
			healthy = false
			issues = append(issues, fmt.Sprintf("Disk usage critical: %.1f%% (%s / %s)",
				usedPct, formatBytes(used), formatBytes(total)))
		} else if usedPct > 80 {
			issues = append(issues, fmt.Sprintf("Disk usage warning: %.1f%% (%s / %s)",
				usedPct, formatBytes(used), formatBytes(total)))
		}
	}

	// Check SMART status if available
	if smartStatus := checkSMARTStatus(path); smartStatus != "" {
		if strings.Contains(smartStatus, "FAILED") || strings.Contains(smartStatus, "ERROR") {
			healthy = false
			issues = append(issues, fmt.Sprintf("SMART status: %s", smartStatus))
		}
	}

	return
}

// checkSMARTStatus checks disk SMART status
func checkSMARTStatus(path string) string {
	// Try to find the device for this mountpoint
	cmd := exec.Command("findmnt", "-n", "-o", "SOURCE", path)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	device := strings.TrimSpace(string(output))
	if device == "" {
		return ""
	}

	// Get SMART status
	cmd = exec.Command("smartctl", "-H", device)
	output, err = cmd.Output()
	if err != nil {
		// smartctl not available or device doesn't support SMART
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "SMART overall-health") {
			return strings.TrimSpace(line)
		}
	}

	return ""
}

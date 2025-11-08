//go:build linux

package main

import (
	"os"
	"os/exec"
	"syscall"
)

// getStorageUsage returns storage usage statistics (Linux)
func getStorageUsage(path string) (used, free, total uint64, err error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, 0, err
	}

	// Available blocks * size = bytes
	total = stat.Blocks * uint64(stat.Bsize)
	free = stat.Bfree * uint64(stat.Bsize)
	used = total - free

	return used, free, total, nil
}

// isMountpoint checks if a path is a mountpoint (Linux)
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

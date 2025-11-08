package main

import (
	"os/exec"
	"strings"
)

// ServiceManager handles systemd service operations
type ServiceManager struct{}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{}
}

// IsActive checks if a service is currently active (running)
func (sm *ServiceManager) IsActive(name string) bool {
	cmd := exec.Command("systemctl", "is-active", name)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "active"
}

// IsEnabled checks if a service is enabled (starts on boot)
func (sm *ServiceManager) IsEnabled(name string) bool {
	cmd := exec.Command("systemctl", "is-enabled", name)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	status := strings.TrimSpace(string(output))
	return status == "enabled" || status == "static"
}

// GetStatus returns full status information for a service
func (sm *ServiceManager) GetStatus(name string) ServiceStatus {
	return ServiceStatus{
		Name:    name,
		Active:  sm.IsActive(name),
		Enabled: sm.IsEnabled(name),
	}
}

// GetMultipleStatuses returns status for multiple services
func (sm *ServiceManager) GetMultipleStatuses(services []string) []ServiceStatus {
	statuses := make([]ServiceStatus, 0, len(services))
	for _, name := range services {
		statuses = append(statuses, sm.GetStatus(name))
	}
	return statuses
}

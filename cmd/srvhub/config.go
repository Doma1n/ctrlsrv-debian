package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Storage   StorageConfig   `yaml:"storage"`
	CUPS      CUPSConfig      `yaml:"cups"`
	Edge      EdgeConfig      `yaml:"edge"`
	WireGuard WireGuardConfig `yaml:"wireguard"`
}

// ServerConfig contains server settings
type ServerConfig struct {
	ListenAddr string `yaml:"listen_addr"`
	QUICAddr   string `yaml:"quic_addr"`
}

// StorageConfig contains storage settings
type StorageConfig struct {
	Path string `yaml:"path"`
}

// CUPSConfig contains CUPS printer settings
type CUPSConfig struct {
	URL     string `yaml:"url"`
	Printer string `yaml:"printer"`
}

// EdgeConfig contains edge proxy settings
type EdgeConfig struct {
	Endpoint string `yaml:"endpoint"`
	TLSCert  string `yaml:"tls_cert"`
	TLSKey   string `yaml:"tls_key"`
}

// WireGuardConfig contains WireGuard settings
type WireGuardConfig struct {
	Interface       string   `yaml:"interface"`
	AllowedNetworks []string `yaml:"allowed_networks"`
}

// loadConfig loads configuration from file
func loadConfig(path string) (*Config, error) {
	// Determine config file path
	if path == "" {
		// Try config.yaml first, then config.example.yaml
		if _, err := os.Stat("config.yaml"); err == nil {
			path = "config.yaml"
		} else if _, err := os.Stat("config.example.yaml"); err == nil {
			path = "config.example.yaml"
		} else {
			return nil, fmt.Errorf("no config file found (tried config.yaml and config.example.yaml)")
		}
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = "0.0.0.0:8080"
	}
	if cfg.Storage.Path == "" {
		cfg.Storage.Path = "/srv/storage1"
	}
	if cfg.CUPS.URL == "" {
		cfg.CUPS.URL = "http://localhost:631"
	}
	if cfg.WireGuard.Interface == "" {
		cfg.WireGuard.Interface = "wg0"
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check storage path exists
	if c.Storage.Path != "" {
		if _, err := os.Stat(c.Storage.Path); os.IsNotExist(err) {
			return fmt.Errorf("storage path does not exist: %s", c.Storage.Path)
		}
	}

	// Check TLS files if edge is configured
	if c.Edge.Endpoint != "" {
		if c.Edge.TLSCert != "" {
			if _, err := os.Stat(c.Edge.TLSCert); os.IsNotExist(err) {
				return fmt.Errorf("TLS cert not found: %s", c.Edge.TLSCert)
			}
		}
		if c.Edge.TLSKey != "" {
			if _, err := os.Stat(c.Edge.TLSKey); os.IsNotExist(err) {
				return fmt.Errorf("TLS key not found: %s", c.Edge.TLSKey)
			}
		}
	}

	return nil
}

// GetPrintDropPath returns the print drop directory path
func (c *Config) GetPrintDropPath() string {
	return filepath.Join(c.Storage.Path, "printdrop")
}

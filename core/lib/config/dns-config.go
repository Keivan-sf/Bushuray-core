package config

import (
	"bushuray-core/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type DNSConfig struct {
	Servers       []string         `json:"servers"`
	QueryStrategy DNSQueryStrategy `json:"query-strategy"`
	Mode          DNSMode          `json:"mode"`
}

type DNSQueryStrategy string

const (
	DNSUseIP     DNSQueryStrategy = "UseIP"
	DNSUseIPv4   DNSQueryStrategy = "UseIPv4"
	DNSUseIPv6   DNSQueryStrategy = "UseIPv6"
	DNSUseSystem DNSQueryStrategy = "UseSystem"
)

type DNSMode string

const (
	DNSModeSystem DNSMode = "system"
	DNSModeProxy  DNSMode = "proxy"
	DNSModeDirect DNSMode = "direct"
)

func defaultDNSConfig() DNSConfig {
	return DNSConfig{
		Servers:       []string{"localhost"},
		QueryStrategy: DNSUseSystem,
		Mode:          DNSModeSystem,
	}
}

func LoadDNSConfig() (DNSConfig, error) {
	currentConfig := defaultDNSConfig()
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return currentConfig, err
	}

	dirPath := filepath.Join(homeDir, ".config", "bushuray")
	configPath := filepath.Join(dirPath, "dns-config.json")
	fileBytes, err := os.ReadFile(configPath)

	if err == nil {
		if err := json.Unmarshal(fileBytes, &currentConfig); err != nil {
			return currentConfig, fmt.Errorf("failed to parse config file: %w", err)
		}
		return currentConfig, nil
	}

	if !os.IsNotExist(err) {
		return currentConfig, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return currentConfig, fmt.Errorf("create config directory %q: %w", dirPath, err)
	}

	defaultConfigJSON, err := json.MarshalIndent(currentConfig, "", " ")
	if err != nil {
		return currentConfig, fmt.Errorf("marshal default config: %w", err)
	}

	if err := os.WriteFile(configPath, defaultConfigJSON, 0o644); err != nil {
		return currentConfig, fmt.Errorf("write default config %q: %w", configPath, err)
	}

	return currentConfig, nil
}

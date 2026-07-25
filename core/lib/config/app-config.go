package config

import (
	"bushuray-core/utils"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type AppConfig struct {
	SocksPort          int       `json:"socks-port"`
	HttpPort           int       `json:"http-port"`
	CoreTCPPort        int       `json:"core-tcp-port"`
	TestPortRange      PortRange `json:"test-port-range"`
	TestURL            string    `json:"test-url"`
	NoBackground       bool      `json:"no-background,omitzero"`
	AutoConnectOnStart bool      `json:"auto-connect-on-start,omitzero"`
}

type PortRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

func defaultAppConfig() AppConfig {
	return AppConfig{
		SocksPort:   3090,
		HttpPort:    3091,
		CoreTCPPort: 4897,
		TestPortRange: PortRange{
			Start: 3095,
			End:   30120,
		},
		NoBackground:       false,
		AutoConnectOnStart: false,
		TestURL:            "https://cp.cloudflare.com",
	}
}

func LoadAppConfig() (AppConfig, error) {
	currentConfig := defaultAppConfig()
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return currentConfig, err
	}
	dirPath := filepath.Join(homeDir, ".config", "bushuray")
	configPath := filepath.Join(dirPath, "config.json")
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

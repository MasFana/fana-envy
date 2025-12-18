package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	AppName       = "fana-envy"
	Version       = "1.0.0"
	EnvFolderName = "envs"
	ConfigName    = ".fana_config"
	HistoryFile   = ".fana_history"
)

type AppConfig struct {
	LastProfile string `json:"last_profile"`
}

func LoadConfig(envDir string) AppConfig {
	var config AppConfig
	data, _ := os.ReadFile(filepath.Join(envDir, ConfigName))
	json.Unmarshal(data, &config)
	return config
}

func SaveConfig(envDir string, lastProfile string) {
	config := AppConfig{LastProfile: lastProfile}
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(envDir, ConfigName), data, 0644)
}

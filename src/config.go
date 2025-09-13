package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	General struct {
		MinDuration     string `yaml:"min_duration"`
		MinDurationTime time.Duration
		EnableNotify    bool `yaml:"enable_notify"`
	} `yaml:"general"`
	
	Docker struct {
		Monitor bool `yaml:"monitor"`
		Filters []string `yaml:"filters"`
	} `yaml:"docker"`
	
	HTTP struct {
		Port    int  `yaml:"port"`
		Enabled bool `yaml:"enabled"`
	} `yaml:"http"`
	
	Notification struct {
		Method   string `yaml:"method"`
		Sound    bool   `yaml:"sound"`
		Position string `yaml:"position"`
	} `yaml:"notification"`
}

const (
	DefaultConfigDir  = ".cmdbell"
	DefaultConfigFile = "config.yaml"
)

func getDefaultConfig() Config {
	config := Config{}
	config.General.MinDuration = "15s"
	config.General.MinDurationTime = 15 * time.Second
	config.General.EnableNotify = true
	
	config.Docker.Monitor = true
	config.Docker.Filters = []string{}
	
	config.HTTP.Port = 59721
	config.HTTP.Enabled = true
	
	config.Notification.Method = "auto"
	config.Notification.Sound = true
	config.Notification.Position = "top-right"
	
	return config
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, DefaultConfigDir)
	configPath := filepath.Join(configDir, DefaultConfigFile)
	
	return configPath, nil
}

func ensureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, DefaultConfigDir)
	
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}
	
	return nil
}

func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config := getDefaultConfig()
		if err := SaveConfig(&config); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return &config, nil
	}
	
	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Parse duration string to time.Duration
	if config.General.MinDuration != "" {
		duration, err := time.ParseDuration(config.General.MinDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid min_duration format: %w", err)
		}
		config.General.MinDurationTime = duration
	} else {
		config.General.MinDurationTime = 15 * time.Second
	}
	
	return &config, nil
}

func SaveConfig(config *Config) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}
	
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}
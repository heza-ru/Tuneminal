package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	// Audio settings
	DefaultVolume float64 `json:"default_volume"`
	ShuffleMode   bool    `json:"shuffle_mode"`
	RepeatMode    bool    `json:"repeat_mode"`

	// UI settings
	Theme string `json:"theme"`

	// Library settings
	MusicDirectory string `json:"music_directory"`
	AutoLoadLast   bool   `json:"auto_load_last"`

	// Performance settings
	BufferSize     int    `json:"buffer_size"`
	SeekStep       int    `json:"seek_step"` // seconds
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		DefaultVolume:  1.0,     // 100%
		ShuffleMode:    false,
		RepeatMode:     false,
		Theme:          "default",
		MusicDirectory: filepath.Join(homeDir, "Music"),
		AutoLoadLast:   true,
		BufferSize:     1024,
		SeekStep:       10, // 10 seconds
	}
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := DefaultConfig()
		// Save default config
		if err := defaultConfig.SaveConfig(configPath); err != nil {
			return defaultConfig, err
		}
		return defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		// If config is corrupted, return default
		defaultConfig := DefaultConfig()
		return defaultConfig, nil
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func (c *Config) SaveConfig(configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".tuneminal", "config.json")
}

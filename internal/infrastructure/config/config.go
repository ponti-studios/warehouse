package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Database DatabaseConfig `json:"database"`
}

// DatabaseConfig holds database-related settings
type DatabaseConfig struct {
	Path string `json:"path"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Database: DatabaseConfig{
			Path: filepath.Join(homeDir, "Documents", "finance", "transactions", "hominem.sqlite3"),
		},
	}
}

// Load reads the configuration from the config file
func Load() (*Config, error) {
	configPath := getConfigPath()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	if dbPath := os.Getenv("HOMINEM_DB_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	}

	return &config, nil
}

// Save writes the configuration to the config file
func (c *Config) Save() error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return ".config/hominem/settings.json"
	}

	// Check for XDG_CONFIG_HOME
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "hominem", "settings.json")
	}

	return filepath.Join(homeDir, ".config", "hominem", "settings.json")
}

// GetDatabasePath returns the database path from config or environment
func GetDatabasePath() string {
	config, err := Load()
	if err != nil {
		// Return default if loading fails
		return DefaultConfig().Database.Path
	}
	return config.Database.Path
}

package config

import (
	"os"
	"strings"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	MQTT     MQTTConfig
	App      AppConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Path     string
	SQLMode  string // aggressive|balanced|safe
}

// MQTTConfig holds MQTT-related configuration
type MQTTConfig struct {
	Broker   string
	Prefix   string
	Username string
	Password string
	ClientID string
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	ConsoleCount  int
	DefaultPrice  int
	DefaultAdmin  AdminConfig
}

// AdminConfig holds default admin configuration
type AdminConfig struct {
	Username string
	Password string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnvOrDefault("PORT", "8080"),
		},
		Database: DatabaseConfig{
			Path:    getEnvOrDefault("DB_PATH", "heheswitch.db"),
			SQLMode: strings.ToLower(getEnvOrDefault("SQLITE_MODE", "balanced")),
		},
		MQTT: MQTTConfig{
			Broker:   os.Getenv("MQTT_BROKER"),
			Prefix:   getEnvOrDefault("MQTT_PREFIX", "ps"),
			Username: os.Getenv("MQTT_USERNAME"),
			Password: os.Getenv("MQTT_PASSWORD"),
			ClientID: os.Getenv("MQTT_CLIENT_ID"),
		},
		App: AppConfig{
			ConsoleCount: 5,
			DefaultPrice: 40000,
			DefaultAdmin: AdminConfig{
				Username: "admin",
				Password: "admin123",
			},
		},
	}
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NormalizePort removes leading colon from port if present
func (c *Config) NormalizePort() {
	c.Server.Port = strings.TrimPrefix(c.Server.Port, ":")
}

// GetDatabaseConnectionString returns the database connection string
func (c *Config) GetDatabaseConnectionString() string {
	return "file:" + c.Database.Path + "?_pragma=busy_timeout(5000)"
}
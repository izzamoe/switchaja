package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear any environment variables that might affect the test
	os.Clearenv()

	config := LoadConfig()

	// Test server config defaults
	assert.Equal(t, "8080", config.Server.Port)

	// Test database config defaults
	assert.Equal(t, "heheswitch.db", config.Database.Path)
	assert.Equal(t, "balanced", config.Database.SQLMode)

	// Test MQTT config defaults
	assert.Equal(t, "", config.MQTT.Broker)
	assert.Equal(t, "ps", config.MQTT.Prefix)
	assert.Equal(t, "", config.MQTT.Username)
	assert.Equal(t, "", config.MQTT.Password)
	assert.Equal(t, "", config.MQTT.ClientID)

	// Test app config defaults
	assert.Equal(t, 5, config.App.ConsoleCount)
	assert.Equal(t, 40000, config.App.DefaultPrice)
	assert.Equal(t, "admin", config.App.DefaultAdmin.Username)
	assert.Equal(t, "admin123", config.App.DefaultAdmin.Password)
}

func TestLoadConfig_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("DB_PATH", "/custom/path/database.db")
	os.Setenv("SQLITE_MODE", "AGGRESSIVE")
	os.Setenv("MQTT_BROKER", "tcp://mqtt.example.com:1883")
	os.Setenv("MQTT_PREFIX", "custom_ps")
	os.Setenv("MQTT_USERNAME", "mqttuser")
	os.Setenv("MQTT_PASSWORD", "mqttpass")
	os.Setenv("MQTT_CLIENT_ID", "client123")

	// Cleanup after test
	defer func() {
		os.Clearenv()
	}()

	config := LoadConfig()

	// Test server config with env vars
	assert.Equal(t, "9090", config.Server.Port)

	// Test database config with env vars
	assert.Equal(t, "/custom/path/database.db", config.Database.Path)
	assert.Equal(t, "aggressive", config.Database.SQLMode) // Should be lowercase

	// Test MQTT config with env vars
	assert.Equal(t, "tcp://mqtt.example.com:1883", config.MQTT.Broker)
	assert.Equal(t, "custom_ps", config.MQTT.Prefix)
	assert.Equal(t, "mqttuser", config.MQTT.Username)
	assert.Equal(t, "mqttpass", config.MQTT.Password)
	assert.Equal(t, "client123", config.MQTT.ClientID)

	// App config should remain unchanged
	assert.Equal(t, 5, config.App.ConsoleCount)
	assert.Equal(t, 40000, config.App.DefaultPrice)
}

func TestGetEnvOrDefault(t *testing.T) {
	// Clear environment
	os.Clearenv()

	t.Run("environment variable exists", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		result := getEnvOrDefault("TEST_VAR", "default_value")
		assert.Equal(t, "test_value", result)
	})

	t.Run("environment variable does not exist", func(t *testing.T) {
		result := getEnvOrDefault("NON_EXISTENT_VAR", "default_value")
		assert.Equal(t, "default_value", result)
	})

	t.Run("environment variable is empty", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_VAR")

		result := getEnvOrDefault("EMPTY_VAR", "default_value")
		assert.Equal(t, "default_value", result)
	})
}

func TestConfig_NormalizePort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "port with colon",
			input:    ":8080",
			expected: "8080",
		},
		{
			name:     "port without colon",
			input:    "8080",
			expected: "8080",
		},
		{
			name:     "empty port",
			input:    "",
			expected: "",
		},
		{
			name:     "port with multiple colons",
			input:    "::8080",
			expected: ":8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Server: ServerConfig{
					Port: tt.input,
				},
			}

			config.NormalizePort()
			assert.Equal(t, tt.expected, config.Server.Port)
		})
	}
}

func TestConfig_GetDatabaseConnectionString(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			Path: "/path/to/database.db",
		},
	}

	connectionString := config.GetDatabaseConnectionString()
	expected := "file:/path/to/database.db?_pragma=busy_timeout(5000)"
	assert.Equal(t, expected, connectionString)
}

func TestConfig_GetDatabaseConnectionString_WithDefaultPath(t *testing.T) {
	config := LoadConfig()

	connectionString := config.GetDatabaseConnectionString()
	expected := "file:heheswitch.db?_pragma=busy_timeout(5000)"
	assert.Equal(t, expected, connectionString)
}
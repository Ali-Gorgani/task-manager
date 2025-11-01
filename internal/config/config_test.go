package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Default values", func(t *testing.T) {
		// Reset viper to use only defaults
		viper.Reset()

		cfg := LoadConfig()
		assert.Equal(t, "3000", cfg.ServerPort)
		assert.Contains(t, cfg.DatabaseURL, "postgres://")
		assert.Equal(t, "localhost:6379", cfg.RedisURL)
		assert.Equal(t, "development", cfg.Environment)
		assert.Equal(t, 0, cfg.RedisDB)
	})

	t.Run("Custom values via Viper", func(t *testing.T) {
		// Reset viper and set custom values
		viper.Reset()
		viper.Set("SERVER_PORT", "9000")
		viper.Set("DATABASE_URL", "postgres://custom:custom@localhost:5432/custom")
		viper.Set("REDIS_URL", "redis:6379")
		viper.Set("REDIS_PASSWORD", "secret")
		viper.Set("REDIS_DB", 5)
		viper.Set("ENVIRONMENT", "production")

		cfg := LoadConfig()
		assert.Equal(t, "9000", cfg.ServerPort)
		assert.Equal(t, "postgres://custom:custom@localhost:5432/custom", cfg.DatabaseURL)
		assert.Equal(t, "redis:6379", cfg.RedisURL)
		assert.Equal(t, "secret", cfg.RedisPassword)
		assert.Equal(t, 5, cfg.RedisDB)
		assert.Equal(t, "production", cfg.Environment)

		// Clean up
		viper.Reset()
	})
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"Development", "development", true},
		{"Production", "production", false},
		{"Staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}
			assert.Equal(t, tt.expected, cfg.IsDevelopment())
		})
	}
}

func TestConfig_GetServerAddress(t *testing.T) {
	cfg := &Config{ServerPort: "3000"}
	assert.Equal(t, ":3000", cfg.GetServerAddress())

	cfg.ServerPort = "9000"
	assert.Equal(t, ":9000", cfg.GetServerAddress())
}

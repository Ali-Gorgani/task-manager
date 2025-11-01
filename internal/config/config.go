package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	ServerPort    string
	DatabaseURL   string
	RedisURL      string
	RedisPassword string
	RedisDB       int
	Environment   string
}

// LoadConfig loads configuration from .env file or environment variables
func LoadConfig() *Config {
	// Set config name and type
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// Add paths to search for .env file
	viper.AddConfigPath(".")
	viper.AddConfigPath("./")

	// Read environment variables (they take precedence over .env file)
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("SERVER_PORT", "3000")
	viper.SetDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/taskmanager?sslmode=disable")
	viper.SetDefault("REDIS_URL", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("ENVIRONMENT", "development")

	// Try to read .env file (not required, just optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No .env file found, using environment variables and defaults")
		} else {
			log.Printf("Error reading .env file: %v", err)
		}
	} else {
		log.Printf("Using .env file: %s", viper.ConfigFileUsed())
	}

	return &Config{
		ServerPort:    viper.GetString("SERVER_PORT"),
		DatabaseURL:   viper.GetString("DATABASE_URL"),
		RedisURL:      viper.GetString("REDIS_URL"),
		RedisPassword: viper.GetString("REDIS_PASSWORD"),
		RedisDB:       viper.GetInt("REDIS_DB"),
		Environment:   viper.GetString("ENVIRONMENT"),
	}
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf(":%s", c.ServerPort)
}

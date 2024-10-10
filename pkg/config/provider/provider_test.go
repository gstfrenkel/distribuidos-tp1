package provider

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig("config.toml")
	assert.NoError(t, err, "LoadConfig should not return an errors")
	assert.NotNil(t, cfg, "config should be loaded and not nil")
}

func TestBool(t *testing.T) {
	cfg, err := LoadConfig("config.toml")
	assert.NoError(t, err)

	debug := cfg.Bool("debug", false)
	assert.Equal(t, true, debug, "debug should be true from the config")

	nonExistentBool := cfg.Bool("non_existent_bool", false)
	assert.Equal(t, false, nonExistentBool, "non_existent_bool should return the default value false")
}

func TestInt(t *testing.T) {
	cfg, err := LoadConfig("config.toml")
	assert.NoError(t, err)

	port := cfg.Int("port", 80)
	assert.Equal(t, 8080, port, "port should be 8080 from the config")

	nonExistentInt := cfg.Int("non_existent_int", 999)
	assert.Equal(t, 999, nonExistentInt, "non_existent_int should return the default value 999")
}

func TestString(t *testing.T) {
	cfg, err := LoadConfig("config.toml")
	assert.NoError(t, err)

	appName := cfg.String("app_name", "defaultApp")
	assert.Equal(t, "TestApp", appName, "app_name should be 'TestApp' from the config")

	nonExistentStr := cfg.String("non_existent_string", "defaultStr")
	assert.Equal(t, "defaultStr", nonExistentStr, "non_existent_string should return the default value 'defaultStr'")
}

func TestDuration(t *testing.T) {
	cfg, err := LoadConfig("config.toml")
	assert.NoError(t, err)

	nonExistentDuration := cfg.Duration("non_existent_duration", 5*time.Second)
	assert.Equal(t, 5*time.Second, nonExistentDuration, "non_existent_duration should return the default value of 5 seconds")
}

func TestUint(t *testing.T) {
	cfg, err := LoadConfig("config.toml")
	assert.NoError(t, err)

	nonExistentUint := cfg.Uint("non_existent_uint", 10)
	assert.Equal(t, uint(10), nonExistentUint, "non_existent_uint should return the default value of 10")
}

func TestSlice(t *testing.T) {
	cfg, err := LoadConfig("config.toml")
	assert.NoError(t, err)

	nonExistentIntSlice := cfg.IntSlice("non_existent_int_slice", []int{1, 2, 3})
	assert.Equal(t, []int{1, 2, 3}, nonExistentIntSlice, "non_existent_int_slice should return the default value [1, 2, 3]")

	nonExistentStringSlice := cfg.StringSlice("non_existent_string_slice", []string{"a", "b"})
	assert.Equal(t, []string{"a", "b"}, nonExistentStringSlice, "non_existent_string_slice should return the default value [a, b]")
}

func TestUnmarshal(t *testing.T) {
	cfg, err := LoadConfig("config.toml")
	assert.NoError(t, err)

	type GatewayConfig struct {
		Host     string
		Port     int
		Username string
		Password string
	}

	var dbConfig GatewayConfig

	err = cfg.Unmarshal("gateway", &dbConfig)
	assert.NoError(t, err, "Unmarshal should not return an errors for existing 'gateway' key")
	assert.Equal(t, "127.0.0.1", dbConfig.Host, "gateway.host should be '127.0.0.1'")
	assert.Equal(t, 9876, dbConfig.Port, "gateway.port should be 9876")
}

func TestContains(t *testing.T) {
	cfg, err := LoadConfig("config.toml")
	assert.NoError(t, err)

	exists := cfg.Contains("port")
	assert.True(t, exists, "port key should exist in the config")

	notExists := cfg.Contains("non_existent_key")
	assert.False(t, notExists, "non_existent_key should not exist in the config")
}

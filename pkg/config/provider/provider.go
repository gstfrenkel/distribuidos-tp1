package provider

import (
	"fmt"
	"os"
	"time"

	"tp1/pkg/config"

	"github.com/spf13/viper"
)

type cfg struct {
	v *viper.Viper
}

// LoadConfig loads a provider with the specified files and local configuration.
func LoadConfig(filepath string) (config.Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	v := viper.New()

	v.SetConfigFile(fmt.Sprintf("%s/%s", wd, filepath))
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return &cfg{v: v}, nil
}

// Bool returns the value associated with the key as a boolean. If the key is missing, def is returned.
func (cfg cfg) Bool(key string, def bool) bool {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetBool(key)
}

// Duration returns the value associated with the key as a duration. If the key is missing, def is returned.
func (cfg cfg) Duration(key string, def time.Duration) time.Duration {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetDuration(key)
}

// String returns the value associated with the key as a string. If the key is missing, def is returned.
func (cfg cfg) String(key string, def string) string {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetString(key)
}

// Int returns the value associated with the key as a signed integer. If the key is missing, def is returned.
func (cfg cfg) Int(key string, def int) int {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetInt(key)
}

// Int32 returns the value associated with the key as a signed integer. If the key is missing, def is returned.
func (cfg cfg) Int32(key string, def int32) int32 {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetInt32(key)
}

// Int64 returns the value associated with the key as a signed integer. If the key is missing, def is returned.
func (cfg cfg) Int64(key string, def int64) int64 {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetInt64(key)
}

// Uint returns the value associated with the key as an unsigned integer. If the key is missing, def is returned.
func (cfg cfg) Uint(key string, def uint) uint {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetUint(key)
}

// Uint16 returns the value associated with the key as an unsigned integer. If the key is missing, def is returned.
func (cfg cfg) Uint16(key string, def uint16) uint16 {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetUint16(key)
}

// Uint32 returns the value associated with the key as an unsigned integer. If the key is missing, def is returned.
func (cfg cfg) Uint32(key string, def uint32) uint32 {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetUint32(key)
}

// Uint64 returns the value associated with the key as an unsigned integer. If the key is missing, def is returned.
func (cfg cfg) Uint64(key string, def uint64) uint64 {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetUint64(key)
}

// Float64 returns the value associated with the key as a float. If the key is missing, def is returned.
func (cfg cfg) Float64(key string, def float64) float64 {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetFloat64(key)
}

// IntSlice returns the value associated with the key as a slice of integers. If the key is missing, def is returned.
func (cfg cfg) IntSlice(key string, def []int) []int {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetIntSlice(key)
}

// StringSlice returns the value associated with the key as a slice of strings. If the key is missing, def is returned.
func (cfg cfg) StringSlice(key string, def []string) []string {
	if !cfg.Contains(key) {
		return def
	}
	return cfg.v.GetStringSlice(key)
}

// Unmarshal returns the value associated with the key as the type of any.
func (cfg cfg) Unmarshal(key string, val any) error {
	if !cfg.Contains(key) {
		return nil
	}
	return cfg.v.UnmarshalKey(key, val)
}

// Contains checks if the key has been set. Contains is case-insensitve for a given key.
func (cfg cfg) Contains(key string) bool {
	return cfg.v.IsSet(key)
}

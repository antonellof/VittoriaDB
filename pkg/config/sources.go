package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ConfigSource represents a source of configuration data
type ConfigSource interface {
	Load(config *VittoriaConfig) error
	Name() string
}

// FileSource loads configuration from a YAML file
type FileSource struct {
	filepath string
}

// NewFileSource creates a new file-based configuration source
func NewFileSource(filepath string) *FileSource {
	return &FileSource{filepath: filepath}
}

func (f *FileSource) Name() string {
	return fmt.Sprintf("file:%s", f.filepath)
}

func (f *FileSource) Load(config *VittoriaConfig) error {
	data, err := os.ReadFile(f.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, skip silently
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	config.Source = f.Name()
	return nil
}

// EnvSource loads configuration from environment variables
type EnvSource struct {
	prefix string
}

// NewEnvSource creates a new environment variable configuration source
func NewEnvSource(prefix string) *EnvSource {
	return &EnvSource{prefix: prefix}
}

func (e *EnvSource) Name() string {
	return fmt.Sprintf("env:%s", e.prefix)
}

func (e *EnvSource) Load(config *VittoriaConfig) error {
	return e.loadFromEnv(reflect.ValueOf(config).Elem(), "")
}

func (e *EnvSource) loadFromEnv(v reflect.Value, prefix string) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get env tag
		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			// If no env tag, try to recurse into struct fields
			if field.Kind() == reflect.Struct {
				newPrefix := prefix
				if prefix != "" {
					newPrefix += "_"
				}
				newPrefix += strings.ToUpper(fieldType.Name)
				if err := e.loadFromEnv(field, newPrefix); err != nil {
					return err
				}
			}
			continue
		}

		// Build full environment variable name
		envName := e.prefix + envTag
		if prefix != "" {
			envName = e.prefix + prefix + "_" + envTag
		}

		// Get environment variable value
		envValue := os.Getenv(envName)
		if envValue == "" {
			continue
		}

		// Set field value based on type
		if err := e.setFieldValue(field, envValue, envName); err != nil {
			return err
		}
	}

	return nil
}

func (e *EnvSource) setFieldValue(field reflect.Value, value, envName string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			// Handle time.Duration
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration for %s: %w", envName, err)
			}
			field.SetInt(int64(duration))
		} else {
			// Handle regular integers
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer for %s: %w", envName, err)
			}
			field.SetInt(intValue)
		}

	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float for %s: %w", envName, err)
		}
		field.SetFloat(floatValue)

	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean for %s: %w", envName, err)
		}
		field.SetBool(boolValue)

	default:
		return fmt.Errorf("unsupported field type %s for %s", field.Kind(), envName)
	}

	return nil
}

// FlagSource loads configuration from command-line flags
type FlagSource struct {
	flags map[string]string
}

// NewFlagSource creates a new flag-based configuration source
func NewFlagSource(flags map[string]string) *FlagSource {
	return &FlagSource{flags: flags}
}

func (f *FlagSource) Name() string {
	return "flags"
}

func (f *FlagSource) Load(config *VittoriaConfig) error {
	// Map common flags to configuration fields
	flagMappings := map[string]func(string) error{
		"port": func(value string) error {
			port, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			config.Server.Port = port
			return nil
		},
		"host": func(value string) error {
			config.Server.Host = value
			return nil
		},
		"data-dir": func(value string) error {
			config.DataDir = value
			return nil
		},
		"log-level": func(value string) error {
			config.Logging.Level = value
			return nil
		},
		"cache-size": func(value string) error {
			size, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			config.Storage.CacheSize = size
			return nil
		},
		"max-workers": func(value string) error {
			workers, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			config.Search.Parallel.MaxWorkers = workers
			return nil
		},
	}

	for flag, value := range f.flags {
		if mapper, exists := flagMappings[flag]; exists {
			if err := mapper(value); err != nil {
				return fmt.Errorf("invalid value for flag --%s: %w", flag, err)
			}
		}
	}

	return nil
}

// DefaultSource provides default configuration values
type DefaultSource struct{}

// NewDefaultSource creates a new default configuration source
func NewDefaultSource() *DefaultSource {
	return &DefaultSource{}
}

func (d *DefaultSource) Name() string {
	return "defaults"
}

func (d *DefaultSource) Load(config *VittoriaConfig) error {
	// Defaults are already set in DefaultConfig()
	// This source is mainly for explicit ordering
	return nil
}

// Convenience functions for common configuration loading patterns

// FromFile loads configuration from a YAML file
func FromFile(filepath string) ConfigSource {
	return NewFileSource(filepath)
}

// FromEnv loads configuration from environment variables with the given prefix
func FromEnv(prefix string) ConfigSource {
	return NewEnvSource(prefix)
}

// FromFlags loads configuration from command-line flags
func FromFlags(flags map[string]string) ConfigSource {
	return NewFlagSource(flags)
}

// FromDefaults provides default configuration values
func FromDefaults() ConfigSource {
	return NewDefaultSource()
}

// LoadConfigFromFile is a convenience function to load config from a single file
func LoadConfigFromFile(filepath string) (*VittoriaConfig, error) {
	return LoadConfig(
		FromDefaults(),
		FromFile(filepath),
		FromEnv("VITTORIA_"),
	)
}

// LoadConfigFromEnv is a convenience function to load config from environment variables
func LoadConfigFromEnv(prefix string) (*VittoriaConfig, error) {
	return LoadConfig(
		FromDefaults(),
		FromEnv(prefix),
	)
}

// LoadConfigWithOverrides loads config with multiple sources and precedence
func LoadConfigWithOverrides(configFile string, envPrefix string, flags map[string]string) (*VittoriaConfig, error) {
	sources := []ConfigSource{
		FromDefaults(),
	}

	if configFile != "" {
		sources = append(sources, FromFile(configFile))
	}

	if envPrefix != "" {
		sources = append(sources, FromEnv(envPrefix))
	}

	if len(flags) > 0 {
		sources = append(sources, FromFlags(flags))
	}

	return LoadConfig(sources...)
}

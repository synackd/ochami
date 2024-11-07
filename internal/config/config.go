package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/OpenCHAMI/ochami/internal/log"
)

// RemoveFromSlice removes an element from a slice and returns the resulting
// slice. The element to be removed is identified by its index in the slice.
func RemoveFromSlice[T any](slice []T, index int) []T {
	slice[len(slice)-1], slice[index] = slice[index], slice[len(slice)-1]
	return slice[:len(slice)-1]
}

// LoadConfig takes a path and config file format and reads in the file pointed
// to by path, loading it as a configuration file using viper. If path is empty,
// an error is returned. LoadConfig accepts any config file types that viper
// accepts. If format is specified (not empty), its value is used as the
// configuration format. If format is empty, the format is guessed by the config
// file's file extension. If there is no file extension or format is empty, YAML
// format is used.
func LoadConfig(path, format string) error {
	if path == "" {
		return fmt.Errorf("no configuration file path passed")
	}
	base := filepath.Dir(path)
	name := filepath.Base(path)
	ext := strings.Trim(filepath.Ext(path), ".")
	fullPath := filepath.Join(base, name)
	log.Logger.Debug().Msgf("Configuration file is: %s", fullPath)

	// Determine format of config file to tell viper to use.
	var viperFormat string
	if format != "" {
		if ext != "" {
			log.Logger.Debug().Msgf("Using passed value of %q as config file format even though file extension is %q", format, ext)
		} else {
			log.Logger.Debug().Msgf("Using passed value of %q as config file format", format)
		}
		viperFormat = strings.ToLower(format)
	} else {
		if ext != "" {
			log.Logger.Debug().Msgf("No config file format passed, inferring from file extension: %s", ext)
			viperFormat = strings.ToLower(ext)
		} else {
			log.Logger.Debug().Msg("No config file format passed and file has no extension, defaulting to YAML")
			viperFormat = "yaml"
		}
	}

	// Tell viper about config file
	viper.SetConfigName(name)
	viper.SetConfigType(viperFormat)
	viper.AddConfigPath(base)

	// Load configuration from file
	err := viper.ReadInConfig()

	return err
}

// WriteConfig takes a path and config file format and writes the current viper
// configuration to the file pointed to by path in the format specified. If path
// is empty, an error is returned. WriteConfig accepts any config file types
// that viper accepts. If format is empty, the format is guessed by the config
// file's file extension. If there is no file extension and format is empty,
// YAML is used.
func WriteConfig(path, format string) error {
	if path == "" {
		return fmt.Errorf("no configuration file path passed")
	}
	base := filepath.Dir(path)
	name := filepath.Base(path)
	ext := strings.Trim(filepath.Ext(path), ".")
	fullPath := filepath.Join(base, name)
	log.Logger.Debug().Msgf("Configuration file is: %s", fullPath)

	// Determine format of config file to tell viper to use.
	var viperFormat string
	if format != "" {
		if ext != "" {
			log.Logger.Debug().Msgf("Using passed value of %q as config file format even though file extension is %q", format, ext)
		} else {
			log.Logger.Debug().Msgf("Using passed value of %q as config file format", format)
		}
		viperFormat = strings.ToLower(format)
	} else {
		if ext != "" {
			log.Logger.Debug().Msgf("No config file format passed, inferring from file extension: %s", ext)
			viperFormat = strings.ToLower(ext)
		} else {
			log.Logger.Debug().Msg("No config file format passed and file has no extension, defaulting to YAML")
			viperFormat = "yaml"
		}
	}

	// Tell viper about config file
	viper.SetConfigName(name)
	viper.SetConfigType(viperFormat)
	viper.AddConfigPath(base)

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}
	log.Logger.Info().Msgf("wrote config to %s", path)

	return nil
}

func SetDefaults() {
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.level", "warning")
}

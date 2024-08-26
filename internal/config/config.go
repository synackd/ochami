package config

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/synackd/ochami/internal/log"
)

// LoadConfig()
func LoadConfig(path, format string) error {
	base := filepath.Dir(path)
	name := filepath.Base(path)
	ext := filepath.Ext(path)
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

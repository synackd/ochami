package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/synackd/ochami/internal/log"
)

// LoadConfig() takes a path and log format and reads in the file pointed to by
// path, loading it as a configuration file using viper. If path is empty, an
// error is returned. LoadConfig() accepts any config file types that viper
// accepts. If format is specified (not empty), its value is used as the
// configuration format. If format is empty, the format is guessed by the config
// file's file extension. If both of these are empty, YAML format is used.
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

package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var (
	Logger zerolog.Logger
)

// Init() initializes the global logging object so it can be used for logging by
// any package that imports this internal log package.
func Init(ll, lf string) error {
	var loggerLevel zerolog.Level
	switch ll {
	case "warning":
		loggerLevel = zerolog.WarnLevel
	case "info":
		loggerLevel = zerolog.InfoLevel
	case "debug":
		loggerLevel = zerolog.DebugLevel
	default:
		return fmt.Errorf("unknown log level: %s", ll)
	}

	switch lf {
	case "rfc3339":
		Logger = zerolog.New(
			zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		).Level(loggerLevel).With().Timestamp().Caller().Logger()
	case "basic":
		Logger = zerolog.New(
			zerolog.ConsoleWriter{
				Out:             os.Stderr,
				FormatTimestamp: func(i interface{}) string { return "" },
				FormatLevel:     func(i interface{}) string { return strings.ToUpper(fmt.Sprintf("%-6s|", i)) },
			},
		).Level(loggerLevel).With().Caller().Logger()
	case "json":
		Logger = zerolog.New(os.Stderr).Level(loggerLevel).With().Timestamp().Logger()
	default:
		return fmt.Errorf("unknown log format: %s", lf)
	}

	return nil
}

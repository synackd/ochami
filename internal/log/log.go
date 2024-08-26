package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type LogLevel int
type LogFormat int

const (
	LogLevelWarning LogLevel = iota
	LogLevelInfo
	LogLevelDebug
)

const (
	LogFormatRFC3339 = iota
	LogFormatBasic
	LogFormatJSON
)

var (
	Logger zerolog.Logger
)

// Init() initializes the global logging object so it can be used for logging by
// any package that imports this internal log package.
func Init(ll LogLevel, lf LogFormat) error {
	var loggerLevel zerolog.Level
	switch ll {
	case LogLevelWarning:
		loggerLevel = zerolog.WarnLevel
	case LogLevelInfo:
		loggerLevel = zerolog.InfoLevel
	case LogLevelDebug:
		loggerLevel = zerolog.DebugLevel
	default:
		return fmt.Errorf("unknown log level: %d", int(ll))
	}

	switch lf {
	case LogFormatRFC3339:
		Logger = zerolog.New(
			zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		).Level(loggerLevel).With().Timestamp().Caller().Logger()
	case LogFormatBasic:
		Logger = zerolog.New(
			zerolog.ConsoleWriter{
				Out:             os.Stderr,
				FormatTimestamp: func(i interface{}) string { return "" },
				FormatLevel:     func(i interface{}) string { return strings.ToUpper(fmt.Sprintf("%-6s|", i)) },
			},
		).Level(loggerLevel).With().Caller().Logger()
	case LogFormatJSON:
		Logger = zerolog.New(os.Stderr).Level(loggerLevel).With().Timestamp().Logger()
	default:
		return fmt.Errorf("unknown log format: %d", int(lf))
	}

	return nil
}

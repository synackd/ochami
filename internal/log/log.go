package log

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type LogLevel int

const (
	LogLevelWarning LogLevel = iota
	LogLevelInfo
	LogLevelDebug
)

var (
	Logger zerolog.Logger
)

// Init() initializes the global logging object so it can be used for logging by
// any package that imports this internal log package.
func Init(l LogLevel) error {
	var loggerLevel zerolog.Level
	switch l {
	case LogLevelWarning:
		loggerLevel = zerolog.WarnLevel
	case LogLevelInfo:
		loggerLevel = zerolog.InfoLevel
	case LogLevelDebug:
		loggerLevel = zerolog.DebugLevel
	default:
		return fmt.Errorf("unknown log level: %d", int(l))
	}

	Logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).Level(loggerLevel).With().Timestamp().Caller().Logger()

	return nil
}

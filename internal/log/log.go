package log

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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

	cw := zerolog.ConsoleWriter{Out: os.Stderr}
	switch lf {
	case "rfc3339":
		cw.TimeFormat = time.RFC3339
		cw.FormatCaller = getFormatCaller(cw.NoColor)
		Logger = zerolog.New(cw).Level(loggerLevel).With().Timestamp().Caller().Logger()
	case "basic":
		cw.FormatTimestamp = func(i interface{}) string { return "" }
		cw.FormatLevel = func(i interface{}) string { return strings.ToUpper(fmt.Sprintf("%-6s|", i)) }
		cw.FormatCaller = getFormatCaller(cw.NoColor)
		Logger = zerolog.New(cw).Level(loggerLevel).With().Caller().Logger()
	case "json":
		Logger = zerolog.New(cw).Level(loggerLevel).With().Timestamp().Logger()
	default:
		return fmt.Errorf("unknown log format: %s", lf)
	}

	return nil
}

// getFormatCaller is a wrapper that generates a Formatter for the
// ConsoleWriter.FormatCaller field. The Formatter generated uses the base name
// of the source file where the log message originated from and ensures that it
// is still colorized, if enabled.
func getFormatCaller(noColor bool) zerolog.Formatter {
	return func(i interface{}) string {
		re := regexp.MustCompile(`(?P<path>.*):(?P<line>\d+)`)
		path := re.ReplaceAllString(i.(string), "${path}")
		line := re.ReplaceAllString(i.(string), "${line}")

		var out string
		_, f, l, ok := runtime.Caller(7)
		if ok {
			out = fmt.Sprintf("%s:%d", filepath.Base(f), l)
		} else {
			out = fmt.Sprintf("%s:%d", path, line)
		}

		return colorize(out, colorBold, noColor) + colorize(" >", colorCyan, noColor)
	}
}

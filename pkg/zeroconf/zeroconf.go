package zeroconf

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogLevel defines what levels of log we support
type LogLevel string

// Pre-defined base log levels
const (
	LogLevelPanic LogLevel = "panic"
	LogLevelError LogLevel = "error"
	LogLevelWarn  LogLevel = "warn"
	LogLevelInfo  LogLevel = "info"
	LogLevelDebug LogLevel = "debug"
)

const (
	googleLevelFieldName = "severity"
)

// Log defines the configurable paremeters supported.
type Log struct {
	Level   LogLevel `json:"level"`
	Console bool     `json:"console"`
}

// Init is used to prepare the standard global logger
func (l *Log) Init(role string) {
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(l.Level.zlevel())
	zerolog.LevelFieldName = googleLevelFieldName
	zl := zerolog.New(os.Stderr)
	if l.Console {
		zl = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		})
	}
	b := zl.With().Timestamp()
	if role != "" {
		b = b.Str("role", role)
	}
	log.Logger = b.Logger()
}

func (ll LogLevel) zlevel() zerolog.Level {
	switch ll {
	case LogLevelPanic:
		return zerolog.PanicLevel
	case LogLevelError:
		return zerolog.ErrorLevel
	case LogLevelWarn:
		return zerolog.WarnLevel
	case LogLevelInfo:
		return zerolog.InfoLevel
	case LogLevelDebug:
		return zerolog.DebugLevel
	default:
		return zerolog.InfoLevel
	}
}

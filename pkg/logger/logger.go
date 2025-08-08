package logger

import (
	"io"
	"os"
	"time"

	"github.com/johnmikee/cuebert/pkg/version"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

const (
	DEBUG = "debug"
	INFO  = "info"
	TRACE = "trace"
	WARN  = "warn"
)

// Logger is our custom logger struct wrapping zerolog.
type Logger struct {
	zerolog.Logger
}

// Env represents the environment the service is running in.
type Env string

const (
	Prod Env = "prod"
	Dev  Env = "dev"
)

// Config is a struct that holds the configuration for the logger.
type Config struct {
	ToFile  bool
	Level   string
	Service string
	Env     string
	Writer  io.Writer // supply a custom writer. only used for testing at the moment.
}

var log zerolog.Logger

// NewLogger returns a new Logger struct with the log to file
// and log level arguments.
func New(l *Config) Logger {
	return Logger{initLogger(l)}
}

// Default returns a logger set to debug.
func Default() Logger {
	return Logger{
		initLogger(&Config{
			ToFile:  false,
			Level:   "debug",
			Service: "logger",
			Env:     "prod",
		}),
	}
}

func WithLogger(zerolog.Logger) Logger {
	return Logger{log}
}

func SetLevel(l string) zerolog.Level {
	return leveler(l)
}

func leveler(l string) zerolog.Level {
	switch l {
	case INFO:
		return zerolog.InfoLevel
	case DEBUG:
		return zerolog.DebugLevel
	case TRACE:
		return zerolog.TraceLevel
	case WARN:
		return zerolog.WarnLevel
	default:
		return zerolog.DebugLevel
	}
}

// Setup initializes the logger with the provided configuration.
func Setup(l *Config) {
	log = initLogger(l)
}

// initLogger initializes a zerolog logger.
func initLogger(l *Config) zerolog.Logger {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339

	var output io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	if l.Env == "prod" {
		output = zerolog.MultiLevelWriter(os.Stderr)
	}

	if l.ToFile {
		fileLogger := &lumberjack.Logger{
			Filename:   l.Service + ".log",
			MaxSize:    10,
			MaxBackups: 10,
			MaxAge:     14,
			Compress:   true,
		}
		output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)
	}

	if l.Writer != nil {
		output = l.Writer
	}

	version := version.Version().Version
	log = zerolog.New(output).
		Level(leveler(l.Level)).
		With().
		Str("service", l.Service).
		Timestamp().
		Caller().
		Str("version", version).
		Logger()

	return log
}

// ChildLogger returns the default logger with a child service appended
// for additional context.
func ChildLogger(service string, z *Logger) Logger {
	return Logger{z.With().Str("child", service).Logger()}
}

func GetLogger() zerolog.Logger {
	return log
}

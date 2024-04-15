package logger_test

import (
	"testing"

	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/rzajac/zltest"
)

func TestNewLogger(t *testing.T) {
	tst := zltest.New(t)
	config := &logger.Config{
		ToFile:  true,
		Level:   logger.DEBUG,
		Service: "test-service",
		Env:     "prod",
		Writer:  tst,
	}
	// Configure zerolog and pass tester as a writer.
	log := logger.New(config)
	// Inject log to tested service or package.
	srv := (log)

	if log.GetLevel().String() != config.Level {
		t.Errorf("NewLogger failed: expected level %s, got %s", config.Level, log.GetLevel().String())
	}

	srv.Debug().Msg("new logger test")

	// Test if log messages were generated properly.
	ent := tst.LastEntry()

	ent.ExpMsg("new logger test")
	ent.ExpLevel(zerolog.DebugLevel)
	ent.ExpKey("service")
}

func TestDefaultLevelSet(t *testing.T) {
	logger := logger.Default()

	if logger.GetLevel() != zerolog.DebugLevel {
		t.Error("Default failed: expected DebugLevel")
	}
}

func TestLeveler(t *testing.T) {
	level := logger.SetLevel(logger.INFO)
	if level != zerolog.InfoLevel {
		t.Errorf("leveler failed: expected InfoLevel, got %s", level.String())
	}

	level = logger.SetLevel(logger.TRACE)
	if level != zerolog.TraceLevel {
		t.Errorf("leveler failed: expected TraceLevel, got %s", level.String())
	}

	level = logger.SetLevel(logger.WARN)
	if level != zerolog.WarnLevel {
		t.Errorf("leveler failed: expected WarnLevel, got %s", level.String())
	}

	level = logger.SetLevel("unknown")
	if level != zerolog.DebugLevel {
		t.Errorf("leveler failed: expected DebugLevel, got %s", level.String())
	}
}

func TestInitLogger(t *testing.T) {
	config := &logger.Config{
		ToFile:  true,
		Level:   logger.DEBUG,
		Service: "test-service",
		Env:     "prod",
	}

	logger.Setup(config)
	log := logger.GetLogger()

	if log.GetLevel().String() != config.Level {
		t.Errorf("initLogger failed: expected level %s, got %s", config.Level, log.GetLevel().String())
	}
}

func TestChildLogger(t *testing.T) {
	tst := zltest.New(t)

	config := &logger.Config{
		ToFile:  true,
		Level:   logger.DEBUG,
		Service: "test-child-logger-service",
		Env:     "prod",
		Writer:  tst,
	}
	// Configure zerolog and pass tester as a writer.
	log := logger.New(config)
	// Inject log to tested service or package.
	srv := (log)

	srv.Debug().Msg("message")

	// Test if log messages were generated properly.
	ent := tst.LastEntry()
	ent.ExpMsg("message")
	ent.ExpLevel(zerolog.DebugLevel)
	ent.ExpKey("service")
	service, _ := ent.Str("service")
	if service != "test-child-logger-service" {
		t.Errorf("Expected service to be 'test-child-logger-service', got '%s'", service)
	}
}

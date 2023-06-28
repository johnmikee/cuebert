package logger

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/rzajac/zltest"
)

func TestNewLogger(t *testing.T) {
	tst := zltest.New(t)
	config := &Config{
		ToFile:  true,
		Level:   DEBUG,
		Service: "test-service",
		Env:     "prod",
		writer:  tst,
	}
	// Configure zerolog and pass tester as a writer.
	log := NewLogger(config)
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
	logger := Default()

	if logger.GetLevel() != zerolog.DebugLevel {
		t.Error("Default failed: expected DebugLevel")
	}

}

func TestLeveler(t *testing.T) {
	level := leveler(INFO)
	if level != zerolog.InfoLevel {
		t.Errorf("leveler failed: expected InfoLevel, got %s", level.String())
	}

	level = leveler(TRACE)
	if level != zerolog.TraceLevel {
		t.Errorf("leveler failed: expected TraceLevel, got %s", level.String())
	}

	level = leveler(WARN)
	if level != zerolog.WarnLevel {
		t.Errorf("leveler failed: expected WarnLevel, got %s", level.String())
	}

	level = leveler("unknown")
	if level != zerolog.DebugLevel {
		t.Errorf("leveler failed: expected DebugLevel, got %s", level.String())
	}
}

func TestInitLogger(t *testing.T) {
	config := &Config{
		ToFile:  true,
		Level:   DEBUG,
		Service: "test-service",
		Env:     "prod",
	}

	logger := initLogger(config)

	if logger.GetLevel().String() != config.Level {
		t.Errorf("initLogger failed: expected level %s, got %s", config.Level, logger.GetLevel().String())
	}
}

func TestChildLogger(t *testing.T) {
	tst := zltest.New(t)

	config := &Config{
		ToFile:  true,
		Level:   DEBUG,
		Service: "test-child-logger-service",
		Env:     "prod",
		writer:  tst,
	}
	// Configure zerolog and pass tester as a writer.
	log := NewLogger(config)
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

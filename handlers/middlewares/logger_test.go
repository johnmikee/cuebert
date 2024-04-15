package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	// Buffer to capture log output
	var logBuf bytes.Buffer

	// Create a logger that writes to the buffer
	log := logger.New(&logger.Config{
		ToFile:  false,
		Level:   logger.DEBUG,
		Service: "test-service",
		Env:     string(logger.Prod),
		Writer:  &logBuf,
	})
	// Define a simple test handler that writes a response
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a request to test the middleware
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Record the response using httptest
	recorder := httptest.NewRecorder()

	// Create the logger middleware and wrap the test handler
	middleware := Logger(&log)
	wrappedHandler := middleware(testHandler)

	// Serve the request
	wrappedHandler.ServeHTTP(recorder, req)

	// Check the response code
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Check the log output
	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "method", "GET")
	assert.Contains(t, logOutput, "url", "/test")
	assert.Contains(t, logOutput, "status", http.StatusOK)

	// Test Panic Recovery
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
	wrappedPanicHandler := middleware(panicHandler)

	// Reset log buffer
	logBuf.Reset()

	// Serve the request that causes panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code did not recover from the panic")
		}
	}()
	wrappedPanicHandler.ServeHTTP(httptest.NewRecorder(), req)

	// Check log for panic recovery
	assert.Contains(t, logBuf.String(), "recover_info", "test panic")
	assert.Contains(t, logBuf.String(), "error")

	// Optionally, you can also verify latency logging, error logging, etc.
}

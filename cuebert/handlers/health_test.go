package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusHandler_SetStatus(t *testing.T) {
	sh := &StatusHandler{}

	status := StatusMessage{
		Message: "Test message",
		Code:    200,
		DB: &DBStatus{
			Connected: true,
		},
		Poll: &RoutineStatus{
			Name:    "Poll",
			Message: "Poll routine is running",
		},
	}

	sh.SetStatus(status)

	got := sh.GetStatus()
	if got.Message != status.Message {
		t.Errorf("SetStatus() failed, got message: %s, want: %s", got.Message, status.Message)
	}
}

func TestStatusHandler_GetStatus(t *testing.T) {
	sh := &StatusHandler{}

	// Set a status
	status := StatusMessage{
		Message: "Test message",
		Code:    200,
		DB: &DBStatus{
			Connected: true,
		},
		Poll: &RoutineStatus{
			Name:    "Poll",
			Message: "Poll routine is running",
		},
	}
	sh.SetStatus(status)

	// Get the status
	got := sh.GetStatus()

	// Compare the fields
	if got.Message != status.Message {
		t.Errorf("GetStatus() failed, got message: %s, want: %s", got.Message, status.Message)
	}

}

func TestStartHealthHandler(t *testing.T) {
	sh := &StatusHandler{}

	// Set a status
	status := StatusMessage{
		Message: "Test message",
		Code:    200,
		DB: &DBStatus{
			Connected: true,
		},
		Poll: &RoutineStatus{
			Name:    "Poll",
			Message: "Poll routine is running",
		},
	}
	sh.SetStatus(status)

	// Create a test request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Create a test HTTP handler for the health endpoint
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the status from the handler
		status := sh.GetStatus()

		// Convert the status message to JSON
		jsonData, err := json.Marshal(status)
		if err != nil {
			t.Fatal(err)
		}

		// Set the appropriate content type
		w.Header().Set("Content-Type", "application/json")

		// Write the JSON data to the response writer
		_, err = w.Write(jsonData)
		if err != nil {
			t.Fatal(err)
		}
	})

	// Serve the request using the test handler
	healthHandler.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("StartHealthHandler() failed, got status code: %d, want: %d", rr.Code, http.StatusOK)
	}

	// Check the response body
	var got StatusMessage
	err = json.Unmarshal(rr.Body.Bytes(), &got)
	if err != nil {
		t.Errorf("StartHealthHandler() failed to parse response body: %v", err)
	}

	// Compare the fields in the response body
	if got.Message != status.Message {
		t.Errorf("StartHealthHandler() failed, got message: %s, want: %s", got.Message, status.Message)
	}
}

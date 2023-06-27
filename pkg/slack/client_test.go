package slack

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestSlackClient_Do_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseBody := `{"status": "ok"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	logger := &logger.Logger{}
	client := NewClient("token", server.URL, nil, logger)

	request, err := client.newRequest("GET", "/some-endpoint", nil)
	if err != nil {
		t.Fatal(err)
	}

	var response interface{}
	err = client.do(request, &response)

	assert.NoError(t, err)
	assert.Equal(t, "ok", response.(map[string]interface{})["status"])
}

func TestSlackClient_DoBuild_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseBody := `{"error": "some_error"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	client := NewClient("token", server.URL, nil, &logger.Logger{})

	request, err := client.newRequest("GET", "/some-endpoint", nil)
	assert.NoError(t, err)

	request.URL = &url.URL{
		Scheme:      "",
		Opaque:      "",
		User:        &url.Userinfo{},
		Host:        "",
		Path:        "",
		RawPath:     "",
		OmitHost:    false,
		ForceQuery:  false,
		RawQuery:    "",
		Fragment:    "",
		RawFragment: "",
	}

	var response interface{}
	err = client.do(request, &response)

	assert.EqualError(t, err, "Get \"//@\": unsupported protocol scheme \"\"")
	assert.Empty(t, response)
}

func TestSlackClient_Do_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseBody := `{"error": "some_error"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	client := NewClient("token", server.URL, nil, &logger.Logger{})

	request, err := client.newRequest("GET", "/some-endpoint", nil)
	assert.NoError(t, err)

	request.Body = nil

	var response interface{}
	err = client.do(request, &response)

	assert.EqualError(t, err, "unexpected response. status=400 api error: {\"error\": \"some_error\"}")
	assert.Empty(t, response)
}

func TestSlackClient_Body_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseBody := `{"error": "invalid_request"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	logger := &logger.Logger{}
	client := NewClient("token", server.URL, nil, logger)

	invalidPayload := make(chan int)
	_, err := client.newRequest("POST", "/some-endpoint", &invalidPayload)

	if err != nil {
		if err.Error() != "json: unsupported type: chan int" {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestSlackClient_DoBody_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseBody := `{"status": "ok"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	logger := &logger.Logger{}
	client := NewClient("token", server.URL, nil, logger)

	payload := struct {
		Message string `json:"message"`
	}{
		Message: "Hello, Slack!",
	}
	request, _ := client.newRequest("POST", "/some-endpoint", payload)

	var response interface{}
	err := client.do(request, &response)

	assert.NoError(t, err)
	assert.Equal(t, "ok", response.(map[string]interface{})["status"])
}

func TestSlackClient_DoBody_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseBody := `{"invalid_json": }`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	logger := &logger.Logger{}
	client := NewClient("token", server.URL, nil, logger)

	payload := struct {
		Message string `json:"message"`
	}{
		Message: "Hello, Slack!",
	}
	request, _ := client.newRequest("POST", "/some-endpoint", payload)

	var response interface{}
	err := client.do(request, &response)

	assert.Error(t, err)
	assert.Empty(t, response)
}

func TestSlackClient_WithCustomClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseBody := `{"status": "ok"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	logger := &logger.Logger{}
	client := NewClient(
		"token",
		server.URL,
		&http.Client{
			Timeout: 10 * time.Second,
		},
		logger,
	)

	request, _ := client.newRequest("GET", "/some-endpoint", nil)

	var response interface{}
	err := client.do(request, &response)

	assert.NoError(t, err)
	assert.Equal(t, "ok", response.(map[string]interface{})["status"])
}

func TestSlackClient_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient("token", server.URL, nil, &logger.Logger{})

	request, _ := client.newRequest("GET", "/some-endpoint", nil)

	var response interface{}
	err := client.do(request, &response)

	assert.EqualError(t, err, "rate-limited")
	assert.Empty(t, response)
}

func TestSlackClient_brokenRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseBody := `{"status": "ok"}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	logger := &logger.Logger{}
	client := NewClient(
		"token",
		server.URL,
		&http.Client{
			Timeout: 10 * time.Second,
		},
		logger,
	)

	request, err := client.newRequest("*?", "/some-endpoint", nil)

	assert.EqualError(t, err, "net/http: invalid method \"*?\"")
	assert.Nil(t, request)
}

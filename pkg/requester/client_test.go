package requester_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/johnmikee/cuebert/pkg/requester"
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (c *MockClient) Do(req *http.Request) (*http.Response, error) {
	return c.DoFunc(req)
}

func TestNew(t *testing.T) {
	body := map[string]interface{}{
		"key": "value",
	}

	t.Run("NoOverrideURL", func(t *testing.T) {
		req, err := requester.New(http.MethodPost, "http://example.com", "/api/endpoint", false, body)
		if err != nil {
			t.Fatalf("Error creating request: %s", err)
		}

		if req.Method != http.MethodPost {
			t.Errorf("Expected method %s, got %s", http.MethodPost, req.Method)
		}

		expectedURL := "http://example.com/api/endpoint"
		if req.URL.String() != expectedURL {
			t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
		}

		var actualBody map[string]interface{}
		decoder := json.NewDecoder(req.Body)
		err = decoder.Decode(&actualBody)
		if err != nil {
			t.Fatalf("Error decoding body: %s", err)
		}

		if !reflect.DeepEqual(body, actualBody) {
			t.Errorf("Expected body %v, got %v", body, actualBody)
		}
	})

	t.Run("WithOverrideURL", func(t *testing.T) {
		overrideURL := "http://custom-url.com/other/endpoint"
		req, err := requester.New(http.MethodPost, "http://example.com", overrideURL, true, body)
		if err != nil {
			t.Fatalf("Error creating request: %s", err)
		}

		if req.Method != http.MethodPost {
			t.Errorf("Expected method %s, got %s", http.MethodPost, req.Method)
		}

		expectedURL := "http://custom-url.com/other/endpoint"
		if req.URL.String() != expectedURL {
			t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
		}

		var actualBody map[string]interface{}
		decoder := json.NewDecoder(req.Body)
		err = decoder.Decode(&actualBody)
		if err != nil {
			t.Fatalf("Error decoding body: %s", err)
		}

		if !reflect.DeepEqual(body, actualBody) {
			t.Errorf("Expected body %v, got %v", body, actualBody)
		}
	})
}

func TestNewWithOverrideURL(t *testing.T) {
	t.Run("OverrideURLProvided", func(t *testing.T) {
		overrideURL := "http://custom-url.com/some/endpoint"
		req, err := requester.New(http.MethodGet, "http://example.com", overrideURL, true, nil)
		if err != nil {
			t.Fatalf("Error creating request: %s", err)
		}

		expectedURL := "http://custom-url.com/some/endpoint"
		if req.URL.String() != expectedURL {
			t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
		}
	})

	t.Run("OverrideURLNotProvided", func(t *testing.T) {
		req, err := requester.New(http.MethodGet, "http://example.com", "/api/endpoint", false, nil)
		if err != nil {
			t.Fatalf("Error creating request: %s", err)
		}

		expectedURL := "http://example.com/api/endpoint"
		if req.URL.String() != expectedURL {
			t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
		}
	})
}

func TestDo_Success(t *testing.T) {
	mockClient := &MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"result": "success"}`)),
			}, nil
		},
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api/endpoint", nil)
	resp, err := requester.Do(mockClient, req, nil)
	if err != nil {
		t.Fatalf("Error performing request: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestDo_ClientError(t *testing.T) {
	mockClient := &MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("client closed request")
		},
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api/endpoint", nil)
	_, err := requester.Do(mockClient, req, nil)

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	expectedError := "client closed request"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestDo_ServerError(t *testing.T) {
	mockClient := &MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error": "internal server error"}`)),
			}, nil
		},
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api/endpoint", nil)
	resp, err := requester.Do(mockClient, req, nil)
	if err != nil {
		t.Fatalf("Error performing request: %s", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

func TestDo_ResponseParsing(t *testing.T) {
	mockClient := &MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"key": "value"}`)),
			}, nil
		},
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api/endpoint", nil)
	var result map[string]interface{}
	_, err := requester.Do(mockClient, req, &result)
	if err != nil {
		t.Fatalf("Error performing request: %s", err)
	}

	expectedResult := map[string]interface{}{
		"key": "value",
	}

	if len(result) != len(expectedResult) {
		t.Errorf("Expected result length %d, got %d", len(expectedResult), len(result))
	}
}

func TestDo_CopyResponseBody(t *testing.T) {
	mockClient := &MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"result": "success"}`)),
			}, nil
		},
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api/endpoint", nil)

	var buf bytes.Buffer
	resp, err := requester.Do(mockClient, req, &buf)
	if err != nil {
		t.Fatalf("Error performing request: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	expectedBody := `{"result": "success"}`
	actualBody := buf.String()

	if actualBody != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, actualBody)
	}
}

func TestHTTPClient(t *testing.T) {
	t.Run("NonNilClient", func(t *testing.T) {
		client := http.DefaultClient
		result := requester.Default(client)

		if result != client {
			t.Errorf("Expected non-nil client, got %+v", result)
		}
	})

	t.Run("NilClient", func(t *testing.T) {
		var client *http.Client = nil
		result := requester.Default(client)

		// We can't directly use IsNil on the reflect.Value of a zero value interface.
		// Instead, we can check if the result is not nil and is an instance of *http.Client.
		if result == nil || reflect.TypeOf(result) != reflect.TypeOf(&http.Client{}) {
			t.Errorf("Expected nil client to be replaced with *http.Client")
		}
	})
}

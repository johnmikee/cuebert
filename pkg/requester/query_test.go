package requester_test

import (
	"reflect"
	"testing"

	"github.com/johnmikee/cuebert/pkg/requester"
)

func TestBuildQuery(t *testing.T) {
	t.Run("NoParams", func(t *testing.T) {
		query := &requester.Query{
			Endpoint: "http://example.com/api/endpoint",
			Params:   nil,
		}

		result, err := requester.BuildQuery(query)
		if err != nil {
			t.Fatalf("Error building query: %s", err)
		}

		expected := "http://example.com/api/endpoint"
		if result != expected {
			t.Errorf("Expected query %s, got %s", expected, result)
		}
	})

	t.Run("WithParams", func(t *testing.T) {
		params := requester.Params{
			"key1": "value1",
			"key2": "value2",
		}

		query := &requester.Query{
			Endpoint: "http://example.com/api/endpoint",
			Params:   params,
		}

		result, err := requester.BuildQuery(query)
		if err != nil {
			t.Fatalf("Error building query: %s", err)
		}

		expected := "http://example.com/api/endpoint?key1=value1&key2=value2"
		if result != expected {
			t.Errorf("Expected query %s, got %s", expected, result)
		}
	})

	t.Run("WithEmptyParams", func(t *testing.T) {
		params := requester.Params{
			"key1": "value1",
			"key2": "value2",
			"key3": "",
			"key4": "value4",
			"key5": "",
			"key6": "",
			"key7": "value7",
		}

		query := &requester.Query{
			Endpoint: "http://example.com/api/endpoint",
			Params:   params,
		}

		result, err := requester.BuildQuery(query)
		if err != nil {
			t.Fatalf("Error building query: %s", err)
		}

		expected := "http://example.com/api/endpoint?key1=value1&key2=value2&key4=value4&key7=value7"
		if result != expected {
			t.Errorf("Expected query %s, got %s", expected, result)
		}
	})
}

func TestStructToQuery(t *testing.T) {
	type Person struct {
		Name    string
		Age     int
		Country string
	}

	person := Person{
		Name:    "John Doe",
		Age:     30,
		Country: "USA",
	}

	expected := map[string]string{
		"Name":    "John Doe",
		"Age":     "30",
		"Country": "USA",
	}

	result := requester.StructToQuery(person)
	if result == nil {
		t.Errorf("Expected result to not be nil")
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

func TestStructToQueryWithTags(t *testing.T) {
	type QueryOpts struct {
		AssetTag     string `json:"asset_tag"`
		DeviceID     string `json:"device_id"`
		DeviceName   string `json:"device_name"`
		Model        string `json:"model"`
		OSVersion    string `json:"os_version"`
		SerialNumber string `json:"serial_number"`
		Platform     string `json:"platform"`
		UserEmail    string `json:"user_email"`
		UserID       string `json:"user_id"`
		UserName     string `json:"user_name"`
	}

	opts := QueryOpts{
		AssetTag:     "1234",
		DeviceID:     "1234",
		DeviceName:   "1234",
		Model:        "1234",
		OSVersion:    "1234",
		SerialNumber: "1234",
		Platform:     "1234",
		UserEmail:    "1234",
		UserID:       "1234",
		UserName:     "1234",
	}

	expected := map[string]string{
		"asset_tag":     "1234",
		"device_id":     "1234",
		"device_name":   "1234",
		"model":         "1234",
		"os_version":    "1234",
		"serial_number": "1234",
		"platform":      "1234",
		"user_email":    "1234",
		"user_id":       "1234",
		"user_name":     "1234",
	}

	result := requester.StructToQuery(opts)
	if result == nil {
		t.Errorf("Expected result to not be nil")
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected:\n%v\nGot:\n%v", expected, result)
	}
}

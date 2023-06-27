package helpers

import (
	"testing"
)

func TestURLShaper(t *testing.T) {
	baseURL := "https://example.com"
	suffix := "/api/v1/"
	expected := "https://example.com/api/v1/"
	result := URLShaper(baseURL, suffix)

	if result != expected {
		t.Errorf("URLShaper failed: expected '%s', got '%s'", expected, result)
	}
}
func TestTokenValidator(t *testing.T) {
	token := "SSWS ABC-123"
	prefix := "SSWS"
	expected := "SSWS ABC-123"
	result := TokenValidator(token, prefix)

	if result != expected {
		t.Errorf("TokenValidator failed: expected '%s', got '%s'", expected, result)
	}
}
func TestTokenValidator_WithoutPrefix(t *testing.T) {
	token := "ABC-123"
	prefix := "SSWS"
	expected := "SSWS ABC-123"
	result := TokenValidator(token, prefix)

	if result != expected {
		t.Errorf("TokenValidator failed: expected '%s', got '%s'", expected, result)
	}
}

func TestURLShaperWithSuffix(t *testing.T) {
	baseURL := "https://example.com"
	suffix := "api/v1"
	expected := "https://example.com/api/v1"
	result := URLShaper(baseURL, suffix)

	if result != expected {
		t.Errorf("URLShaper failed: expected '%s', got '%s'", expected, result)
	}
}

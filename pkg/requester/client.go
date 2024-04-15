package requester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/johnmikee/cuebert/pkg/helpers"
)

// HTTPClient is an interface representing an HTTP client's Do method.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Default returns the default HTTP client.
func Default(c HTTPClient) HTTPClient {
	return httpClient(c)
}

// httpClient creates an HTTP client if the input is nil.
func httpClient(c HTTPClient) HTTPClient {
	if !reflect.ValueOf(c).IsNil() {
		return c
	}

	return &http.Client{}
}

// New creates a new HTTP request with the specified method, base URL, endpoint, and optional body.
func New(method, baseUrl, endpoint string, override bool, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer

	var url string
	if override {
		url = endpoint
	} else {
		url = fmt.Sprintf("%s%s", helpers.URLShaper(baseUrl, "/"), strings.TrimPrefix(endpoint, "/"))
	}

	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// Do performs an HTTP request using the provided client and request, optionally decoding the response body into v.
func Do(client HTTPClient, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := httpClient(client).Do(req)
	if err != nil {
		return nil, err
	}

	// If StatusCode is not in the 200 range, something went wrong. Return the
	// response but do not process its body.
	if o := resp.StatusCode; 200 > o || o > 299 {
		return resp, nil
	}

	defer resp.Body.Close()
	if v != nil {
		// If v implements io.Writer, copy the response body to it.
		if w, ok := v.(io.Writer); ok {
			_, err := io.Copy(w, resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to copy response body: %s", err.Error())
			}
		} else {
			// Otherwise, decode the response body into v.
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // Ignore EOF errors caused by empty response body.
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return resp, err
}

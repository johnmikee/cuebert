package slack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestSlackClient_GetSlackMemberEmail(t *testing.T) {
	// Create a mock server to simulate the Slack API response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request URL and method
		assert.Equal(t, "/api/users.info", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Write a sample response body
		responseBody := `{"ok": true, "user": {"id": "user123", "name": "john.doe", "profile": {"real_name": "John Doe"}}}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	// Create a SlackClient instance with the mock server URL
	client := NewClient("token", server.URL, nil, &logger.Logger{})

	// Call the GetSlackMemberEmail method
	resp, err := client.GetSlackMemberEmail("user123")

	// Verify the response and error
	assert.NoError(t, err)
	assert.True(t, resp.Ok)
	assert.Equal(t, "user123", resp.User.ID)
	assert.Equal(t, "john.doe", resp.User.Name)
	assert.Equal(t, "John Doe", resp.User.Profile.RealName)
}

func TestSlackClient_GetSlackID(t *testing.T) {
	// Create a mock server to simulate the Slack API response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request URL and method
		assert.Equal(t, "/api/users.lookupByEmail", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Write a sample response body
		responseBody := `{"ok": true, "user": {"id": "user123", "name": "john.doe", "profile": {"real_name": "John Doe"}}}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	// Create a SlackClient instance with the mock server URL
	client := NewClient("token", server.URL, nil, &logger.Logger{})

	// Call the GetSlackID method
	resp, err := client.GetSlackID("john.doe@example.com")

	// Verify the response and error
	assert.NoError(t, err)
	assert.True(t, resp.Ok)
	assert.Equal(t, "user123", resp.User.ID)
	assert.Equal(t, "john.doe", resp.User.Name)
	assert.Equal(t, "John Doe", resp.User.Profile.RealName)
}

func TestUserPagination_list(t *testing.T) {
	// Create a mock server to simulate the Slack API response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request URL and method
		assert.Equal(t, "/api/users.list", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Write a sample response body
		responseBody := `{"ok": true, "members": [{"id": "user1", "name": "john.doe1"}, {"id": "user2", "name": "john.doe2"}], "response_metadata": {"next_cursor": ""}}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	// Create a SlackClient instance with the mock server URL
	client := NewClient("token", server.URL, nil, &logger.Logger{})

	// Create a UserPagination instance with the mock server and client
	p := &UserPagination{
		users:  []Member{},
		limit:  200,
		cursor: "123",
		c:      client,
	}

	// Call the list method
	resp, err := p.list()

	// Verify the response and error
	assert.NoError(t, err)
	assert.True(t, resp.Ok)
	assert.Len(t, resp.Members, 2)
	assert.Equal(t, "user1", resp.Members[0].ID)
	assert.Equal(t, "john.doe1", resp.Members[0].Name)
	assert.Equal(t, "user2", resp.Members[1].ID)
	assert.Equal(t, "john.doe2", resp.Members[1].Name)
}

func TestSlackClient_ListUsers(t *testing.T) {
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseBody := ListResponse{
			Ok: true,
			Members: []Member{
				{ID: "user" + fmt.Sprint(count), Name: "john.doe" + fmt.Sprint(count)},
				{ID: "user" + fmt.Sprint(count+1), Name: "john.doe" + fmt.Sprint(count+1)},
				{ID: "user" + fmt.Sprint(count+2), Name: "john.doe" + fmt.Sprint(count+2)},
			},
			ResponseMetadata: ResponseMetadata{
				NextCursor: func() string {
					if count >= 5 {
						return ""
					} else {
						return fmt.Sprint(count)
					}
				}(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		resp, err := json.Marshal(responseBody)
		count++
		assert.NoError(t, err)
		_, _ = w.Write(resp)

	}))
	defer server.Close()

	// Create a SlackClient instance with the mock server URL
	client := NewClient("token", server.URL, nil, &logger.Logger{})

	// Call the ListUsers method
	users := client.ListUsers()

	// Verify the retrieved users
	assert.Len(t, users, 15)
	assert.Equal(t, "user0", users[0].ID)
	assert.Equal(t, "john.doe1", users[1].Name)
}

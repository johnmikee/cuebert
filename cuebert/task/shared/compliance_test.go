package shared

import (
	"testing"

	"github.com/johnmikee/cuebert/db/shared/overdue"
	"github.com/stretchr/testify/assert"
)

func TestCheckCompliance(t *testing.T) {
	ou := []overdue.Info{
		{UserID: 1, UserServiceID: "svc1", UserEmail: "email1@example.com"},
		{UserID: 2, UserServiceID: "svc2", UserEmail: "email2@example.com"},
		{UserID: 3, UserServiceID: "svc3", UserEmail: "email3@example.com"},
	}

	emailsToCheck := []string{"email1@example.com", "email3@example.com"}

	// Create a map for efficient lookup
	emailMap := make(map[string]struct{})
	for _, email := range emailsToCheck {
		emailMap[email] = struct{}{}
	}

	// Filter ou slice
	filteredOU := []overdue.Info{}
	for _, item := range ou {
		if _, exists := emailMap[item.UserEmail]; exists {
			filteredOU = append(filteredOU, item)
		}
	}

	assert.Contains(t, filteredOU, overdue.Info{UserID: 1, UserServiceID: "svc1", UserEmail: "email1@example.com"})
	assert.Contains(t, filteredOU, overdue.Info{UserID: 3, UserServiceID: "svc3", UserEmail: "email3@example.com"})
	assert.NotContains(t, filteredOU, overdue.Info{UserID: 2, UserServiceID: "svc2", UserEmail: "email2@example.com"})
}

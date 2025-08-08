package users

import (
	"context"

	"github.com/johnmikee/cuebert/db/repository"
	"github.com/johnmikee/cuebert/pkg/errors"
)

// QueryRefactored demonstrates how to use the repository pattern instead of the old query builder.
// This replaces the old Query struct and its methods.
type QueryRefactored struct {
	userRepo repository.Repository[repository.User]
}

// NewQueryRefactored creates a new query instance using the repository pattern.
func NewQueryRefactored(config *Config) *QueryRefactored {
	userRepo := repository.NewUserRepository(config.db, config.log)
	return &QueryRefactored{
		userRepo: userRepo,
	}
}

// All fetches all users using the repository pattern.
func (q *QueryRefactored) All(ctx context.Context) ([]repository.User, error) {
	users, err := q.userRepo.FindAll(ctx, nil)
	if err != nil {
		return nil, errors.DBError("fetch all users", err)
	}
	return users, nil
}

// BySlackID finds a user by their Slack ID.
func (q *QueryRefactored) BySlackID(ctx context.Context, slackID string) (*repository.User, error) {
	user, err := q.userRepo.FindByID(ctx, slackID)
	if err != nil {
		return nil, errors.DBError("find user by slack ID", err)
	}
	return user, nil
}

// ByEmail finds users by email.
func (q *QueryRefactored) ByEmail(ctx context.Context, email string) ([]repository.User, error) {
	users, err := q.userRepo.FindAll(ctx, repository.Conditions{"user_email": email})
	if err != nil {
		return nil, errors.DBError("find users by email", err)
	}
	return users, nil
}

// ByServiceID finds users by service ID.
func (q *QueryRefactored) ByServiceID(ctx context.Context, serviceID string) ([]repository.User, error) {
	users, err := q.userRepo.FindAll(ctx, repository.Conditions{"user_service_id": serviceID})
	if err != nil {
		return nil, errors.DBError("find users by service ID", err)
	}
	return users, nil
}

// Usage example showing how to migrate old code:
//
// Old way:
//   query := config.By().All().Query()
//   users, err := query.Query()
//
// New way:
//   query := NewQueryRefactored(config)
//   users, err := query.All(context.Background())
//
// Old way with filter:
//   query := config.By().Opts(UserSlackID, "U12345").Query()
//   users, err := query.Query()
//
// New way with filter:
//   query := NewQueryRefactored(config)
//   user, err := query.BySlackID(context.Background(), "U12345")
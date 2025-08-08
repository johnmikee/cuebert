package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// User represents a user entity.
type User struct {
	UserServiceID string    `db:"user_service_id" json:"user_service_id"`
	UserLongName  string    `db:"user_long_name" json:"user_long_name"`
	UserEmail     string    `db:"user_email" json:"user_email"`
	UserSlackID   string    `db:"user_slack_id" json:"user_slack_id"`
	TZOffset      int       `db:"tz_offset" json:"tz_offset"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// UserRepository extends the base repository with user-specific methods.
type UserRepository interface {
	Repository[User]
	FindBySlackID(ctx context.Context, slackID string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindActiveUsers(ctx context.Context) ([]User, error)
	UpdateTimezone(ctx context.Context, slackID string, offset int) error
}

type userRepository struct {
	Repository[User]
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *pgxpool.Pool, log logger.Logger) UserRepository {
	base := New[User](Config{
		DB:         db,
		Log:        log,
		Table:      "users",
		PrimaryKey: "user_slack_id",
	})

	return &userRepository{
		Repository: base,
	}
}

// FindBySlackID finds a user by Slack ID.
func (r *userRepository) FindBySlackID(ctx context.Context, slackID string) (*User, error) {
	return r.FindOne(ctx, Conditions{"user_slack_id": slackID})
}

// FindByEmail finds a user by email.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	return r.FindOne(ctx, Conditions{"user_email": email})
}

// FindActiveUsers returns all active users.
func (r *userRepository) FindActiveUsers(ctx context.Context) ([]User, error) {
	// This would include additional logic for determining active users
	// For now, return all users
	return r.FindAll(ctx, Conditions{})
}

// UpdateTimezone updates a user's timezone offset.
func (r *userRepository) UpdateTimezone(ctx context.Context, slackID string, offset int) error {
	return r.Update(ctx, slackID, Updates{
		"tz_offset": offset,
	})
}

package exclusions

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Info represents the columns in the exclusions table
type Info struct {
	Approved     bool      `json:"approved"`
	CreatedAt    time.Time `json:"created_at"`
	SerialNumber string    `json:"serial_number"`
	Reason       string    `json:"reason"`
	UpdatedAt    time.Time `json:"updated_at"`
	UserEmail    string    `json:"user_email"`
	Until        time.Time `json:"until"`
}

type EI []Info

func (e EI) Empty() bool {
	return len(e) == 0
}

type Config struct {
	db  *pgxpool.Conn
	ctx context.Context
	log logger.Logger
	st  sq.StatementBuilderType
}

const table = "exclusions"

var columns = []string{
	"approved",
	"serial_number",
	"user_email",
	"reason",
	"until",
	"created_at",
	"updated_at",
}

// Exclusion returns a new client used to interact with the exclusions table
func Exclusion(d *pgxpool.Pool, l *logger.Logger) *Config {
	conn, err := d.Acquire(context.Background())
	if err != nil {
		l.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	return &Config{
		ctx: context.Background(),
		db:  conn,
		log: logger.ChildLogger("db/exclusions", l),
		st:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

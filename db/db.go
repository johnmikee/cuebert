package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"

	"github.com/johnmikee/cuebert/pkg/logger"
)

// DBConfig is used to configure the connection to the DB.
type DBConfig struct {
	Host     string
	Name     string
	Password string
	Port     string
	User     string

	ctx context.Context
	db  *pgxpool.Pool
	log logger.Logger
}

type config struct {
	host     string
	name     string
	password string
	port     string
	user     string

	ctx context.Context
	db  *pgxpool.Pool
}

// Close is used to close the connection to the DB.
type Close func(context.Context, *pgxpool.Pool)

// PGDB is returned after an established connection has been made.
//
// The values returned willbe passed around as different functions
// interact with the DB.
type PGDB struct {
	DB     *pgxpool.Pool
	Ctx    context.Context
	Logger logger.Logger
	Close  Close
}

type DB = pgxpool.Pool

var CueTables = []string{"bot_results", "devices", "exclusions", "users"}

// New returns a new empty dbconfig. The methods of dbconfig below help to configure
// the required items to form a connection with the DB.
func New() *config {
	return &config{}
}

// NewConn initializes a new config set by passing the values.
func NewConn(d *DBConfig) *DBConfig {
	return &DBConfig{
		Host:     d.Host,
		Name:     d.Name,
		Password: d.Password,
		Port:     d.Port,
		User:     d.User,
		ctx:      d.ctx,
		db:       d.db,
		log:      d.log,
	}
}

// Connect establishes the connection with the DB.
//
// note: this does not close the connection. please use the Close
// function to end the connection when done.
func (d *DBConfig) Connect() (*PGDB, error) {
	ok, err := d.validate()

	if !ok {
		return nil, err
	}

	connArgs := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", d.User, d.Password, d.Host, d.Port, d.Name)
	conn, err := pgxpool.New(context.Background(), connArgs)

	if err != nil {
		return nil, err
	}

	return &PGDB{
		DB:     conn,
		Ctx:    d.ctx,
		Logger: d.log,
		Close:  d.closeConnection,
	}, err
}

// closeConnection will terminate the connection with the DB.
func (d *DBConfig) closeConnection(ctx context.Context, db *pgxpool.Pool) {
	d.log.Debug().Msg("closing the connection")
	db.Close()
}

// validate will check the required arguments to establish a connection
func (d *DBConfig) validate() (bool, error) {
	if d.User == "" {
		return false, errors.New("must include user value")
	}

	if d.Password == "" {
		return false, errors.New("must include password value")
	}

	if d.Host == "" {
		return false, errors.New("must include host value")
	}

	if d.Name == "" {
		return false, errors.New("must include dbname value")
	}

	if d.Port == "" {
		return false, errors.New("must include port value")
	}

	return true, nil
}

// User is the user used to authenticate to the DB
func (c *config) User(user string) *config {
	c.user = user
	return c
}

// Host is the address used to reach the DB
func (c *config) Host(host string) *config {
	c.host = host
	return c
}

// DBName is the name of the DB
func (c *config) DBName(dbname string) *config {
	c.name = dbname
	return c
}

// Password is the password used to authenticate to the DB
func (c *config) Password(password string) *config {
	c.password = password
	return c
}

// Port is the port the DB is listening on
func (c *config) Port(port string) *config {
	c.port = port
	return c
}

// WithContext allows you to optionally pass in context
func (c *config) WithContext(ctx context.Context) *config {
	c.ctx = ctx
	return c
}

// WithDB will allow you to optionally pass in a preconfigured DB
func (c *config) WithDB(db *pgxpool.Pool) *config {
	c.db = db
	return c
}

// Connect connects to the DB.
func Connect(cfg *DBConfig) (*PGDB, error) {
	db, err := cfg.Connect()

	return db, err
}

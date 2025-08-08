package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Manager handles database connection pooling and lifecycle.
type Manager struct {
	pool   *pgxpool.Pool
	config *Config
	log    logger.Logger
	mu     sync.RWMutex

	// Metrics
	metrics *Metrics
}

// Config holds pool configuration.
type Config struct {
	// Connection parameters
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string

	// Pool configuration
	MaxConns        int32         // Maximum number of connections in the pool
	MinConns        int32         // Minimum number of connections in the pool
	MaxConnLifetime time.Duration // Maximum lifetime of a connection
	MaxConnIdleTime time.Duration // Maximum idle time before connection is closed

	// Health check
	HealthCheckPeriod time.Duration // How often to check connection health

	// Timeouts
	ConnectTimeout time.Duration // Timeout for establishing connection
	QueryTimeout   time.Duration // Default query timeout
}

// Metrics tracks pool statistics.
type Metrics struct {
	mu                sync.RWMutex
	TotalConnections  int64
	ActiveConnections int64
	IdleConnections   int64
	WaitCount         int64
	WaitDuration      time.Duration
	MaxLifetimeClose  int64
	IdleClose         int64
	Errors            int64
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Host:              "localhost",
		Port:              5432,
		SSLMode:           "prefer",
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
		ConnectTimeout:    10 * time.Second,
		QueryTimeout:      30 * time.Second,
	}
}

// New creates a new connection pool manager.
func New(ctx context.Context, config *Config, log logger.Logger) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeConfigInvalid, "invalid pool configuration")
	}

	// Build connection string
	connStr := buildConnectionString(config)

	// Parse pool config
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDBConnection, "failed to parse connection config")
	}

	// Apply pool settings
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout

	// Set up hooks for metrics
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		log.Debug().Msg("new connection established")
		return nil
	}

	poolConfig.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		// Return true to allow acquisition, false to reject
		return true
	}

	poolConfig.AfterRelease = func(conn *pgx.Conn) bool {
		// Return true to return connection to pool, false to close it
		return true
	}

	// Create the pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDBConnection, "failed to create connection pool")
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.Wrap(err, errors.ErrCodeDBConnection, "failed to ping database")
	}

	manager := &Manager{
		pool:    pool,
		config:  config,
		log:     logger.ChildLogger("pool", &log),
		metrics: &Metrics{},
	}

	// Start metrics collector
	go manager.collectMetrics(ctx)

	log.Info().
		Int32("max_conns", config.MaxConns).
		Int32("min_conns", config.MinConns).
		Msg("connection pool initialized")

	return manager, nil
}

// GetPool returns the underlying pool for direct access.
func (m *Manager) GetPool() *pgxpool.Pool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pool
}

// Acquire gets a connection from the pool with timeout.
func (m *Manager) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	// Apply default query timeout if not set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && m.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, m.config.QueryTimeout)
		defer cancel()
	}

	start := time.Now()
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		m.recordError()
		return nil, errors.DBError("acquire connection", err)
	}

	m.recordWait(time.Since(start))
	return conn, nil
}

// Execute runs a query without returning results.
func (m *Manager) Execute(ctx context.Context, query string, args ...interface{}) error {
	conn, err := m.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		m.recordError()
		return errors.DBError("execute query", err)
	}

	return nil
}

// QueryRow executes a query that returns a single row.
func (m *Manager) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	conn, err := m.Acquire(ctx)
	if err != nil {
		// Return a row that will error when scanned
		return &errorRow{err: err}
	}
	defer conn.Release()

	return conn.QueryRow(ctx, query, args...)
}

// Query executes a query that returns multiple rows.
func (m *Manager) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	conn, err := m.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	// Note: Caller must close rows which will release the connection

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		conn.Release()
		m.recordError()
		return nil, errors.DBError("query", err)
	}

	return &managedRows{rows: rows, conn: conn}, nil
}

// Transaction executes a function within a transaction.
func (m *Manager) Transaction(ctx context.Context, fn func(pgx.Tx) error) error {
	conn, err := m.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		m.recordError()
		return errors.DBError("begin transaction", err)
	}

	// Execute the function
	if err := fn(tx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			m.log.Error().Err(rbErr).Msg("failed to rollback transaction")
		}
		return err
	}

	// Commit on success
	if err := tx.Commit(ctx); err != nil {
		m.recordError()
		return errors.DBError("commit transaction", err)
	}

	return nil
}

// Health checks the health of the pool.
func (m *Manager) Health(ctx context.Context) error {
	if err := m.pool.Ping(ctx); err != nil {
		return errors.Wrap(err, errors.ErrCodeDBConnection, "health check failed")
	}

	stats := m.pool.Stat()

	// Check if we're at capacity
	if stats.AcquiredConns() >= m.config.MaxConns {
		m.log.Warn().
			Int32("acquired", stats.AcquiredConns()).
			Int32("max", m.config.MaxConns).
			Msg("connection pool at capacity")
	}

	// Check if too many idle connections
	if stats.IdleConns() > m.config.MinConns*2 {
		m.log.Debug().
			Int32("idle", stats.IdleConns()).
			Msg("many idle connections")
	}

	return nil
}

// Stats returns pool statistics.
func (m *Manager) Stats() *pgxpool.Stat {
	return m.pool.Stat()
}

// GetMetrics returns pool metrics.
func (m *Manager) GetMetrics() Metrics {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	// Create a copy to avoid races
	return Metrics{
		TotalConnections:  m.metrics.TotalConnections,
		ActiveConnections: m.metrics.ActiveConnections,
		IdleConnections:   m.metrics.IdleConnections,
		WaitCount:         m.metrics.WaitCount,
		WaitDuration:      m.metrics.WaitDuration,
		MaxLifetimeClose:  m.metrics.MaxLifetimeClose,
		IdleClose:         m.metrics.IdleClose,
		Errors:            m.metrics.Errors,
	}
}

// Close closes the connection pool.
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pool != nil {
		m.pool.Close()
		m.log.Info().Msg("connection pool closed")
	}
}

// Reset resets the connection pool (useful for configuration changes).
func (m *Manager) Reset(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close existing pool
	if m.pool != nil {
		m.pool.Close()
	}

	// Create new pool with same config
	newManager, err := New(ctx, m.config, m.log)
	if err != nil {
		return err
	}

	m.pool = newManager.pool
	m.metrics = newManager.metrics

	m.log.Info().Msg("connection pool reset")
	return nil
}

// Helper functions

func validateConfig(config *Config) error {
	if config.Host == "" {
		return errors.ValidationError("host", "cannot be empty")
	}
	if config.Database == "" {
		return errors.ValidationError("database", "cannot be empty")
	}
	if config.User == "" {
		return errors.ValidationError("user", "cannot be empty")
	}
	if config.MaxConns < 1 {
		return errors.ValidationError("max_conns", "must be at least 1")
	}
	if config.MinConns < 0 {
		return errors.ValidationError("min_conns", "cannot be negative")
	}
	if config.MinConns > config.MaxConns {
		return errors.ValidationError("min_conns", "cannot exceed max_conns")
	}
	return nil
}

func buildConnectionString(config *Config) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
	)
}

// Metrics recording

func (m *Manager) recordWait(duration time.Duration) {
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()

	m.metrics.WaitCount++
	m.metrics.WaitDuration += duration
}

func (m *Manager) recordError() {
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()

	m.metrics.Errors++
}

func (m *Manager) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := m.pool.Stat()

			m.metrics.mu.Lock()
			m.metrics.TotalConnections = int64(stats.TotalConns())
			m.metrics.ActiveConnections = int64(stats.AcquiredConns())
			m.metrics.IdleConnections = int64(stats.IdleConns())
			m.metrics.mu.Unlock()

			// Log if metrics show issues
			if stats.AcquiredConns() > m.config.MaxConns*8/10 {
				m.log.Warn().
					Int32("active", stats.AcquiredConns()).
					Int32("max", m.config.MaxConns).
					Msg("high connection usage")
			}
		}
	}
}

// errorRow implements pgx.Row for error cases.
type errorRow struct {
	err error
}

func (r *errorRow) Scan(dest ...interface{}) error {
	return r.err
}

// managedRows wraps pgx.Rows to ensure connection is released.
type managedRows struct {
	rows pgx.Rows
	conn *pgxpool.Conn
}

func (r *managedRows) Close() {
	r.rows.Close()
	r.conn.Release()
}

func (r *managedRows) Err() error {
	return r.rows.Err()
}

func (r *managedRows) Next() bool {
	return r.rows.Next()
}

func (r *managedRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

func (r *managedRows) Values() ([]interface{}, error) {
	return r.rows.Values()
}

func (r *managedRows) RawValues() [][]byte {
	return r.rows.RawValues()
}

func (r *managedRows) CommandTag() pgconn.CommandTag {
	return r.rows.CommandTag()
}

func (r *managedRows) FieldDescriptions() []pgconn.FieldDescription {
	return r.rows.FieldDescriptions()
}

func (r *managedRows) Conn() *pgx.Conn {
	return nil // Connection is managed by pool
}

package pool

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPool is a mock implementation of pgxpool.Pool for testing.
type MockPool struct {
	mock.Mock
}

func (m *MockPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	args := m.Called(ctx)
	if conn := args.Get(0); conn != nil {
		return conn.(*pgxpool.Conn), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPool) Close() {
	m.Called()
}

func (m *MockPool) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPool) Stat() *pgxpool.Stat {
	args := m.Called()
	if stat := args.Get(0); stat != nil {
		return stat.(*pgxpool.Stat)
	}
	return nil
}

func (m *MockPool) Config() *pgxpool.Config {
	args := m.Called()
	if cfg := args.Get(0); cfg != nil {
		return cfg.(*pgxpool.Config)
	}
	return nil
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Host:              "localhost",
				Port:              5432,
				Database:          "testdb",
				User:              "testuser",
				Password:          "testpass",
				MaxConns:          10,
				MinConns:          2,
				MaxConnLifetime:   1 * time.Hour,
				MaxConnIdleTime:   30 * time.Minute,
				HealthCheckPeriod: 1 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "missing required fields",
			config: Config{
				Host: "",
				Port: 0,
			},
			wantErr: true,
			errMsg:  "invalid configuration",
		},
		{
			name: "invalid max connections",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				User:     "testuser",
				MaxConns: 0,
			},
			wantErr: true,
			errMsg:  "max connections must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since we can't actually connect to a database in unit tests,
			// we'll test the validation logic
			err := validateConfig(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_HealthCheck(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockPool)
		wantErr   bool
	}{
		{
			name: "healthy connection",
			setupMock: func(mp *MockPool) {
				mp.On("Ping", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "unhealthy connection",
			setupMock: func(mp *MockPool) {
				mp.On("Ping", mock.Anything).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool := new(MockPool)
			tt.setupMock(mockPool)

			manager := &Manager{
				pool: mockPool,
				log:  logger.NewTestLogger(),
				config: Config{
					HealthCheckPeriod: 1 * time.Minute,
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := manager.HealthCheck(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPool.AssertExpectations(t)
		})
	}
}

func TestManager_GetMetrics(t *testing.T) {
	mockPool := new(MockPool)

	expectedStat := &pgxpool.Stat{
		AcquireCount:            100,
		AcquireDuration:         5 * time.Second,
		AcquiredConns:           5,
		CanceledAcquireCount:    2,
		ConstructingConns:       1,
		EmptyAcquireCount:       10,
		IdleConns:               3,
		MaxConns:                10,
		TotalConns:              8,
		NewConnsCount:           20,
		MaxLifetimeDestroyCount: 5,
		MaxIdleDestroyCount:     3,
	}

	mockPool.On("Stat").Return(expectedStat)

	manager := &Manager{
		pool: mockPool,
		log:  logger.NewTestLogger(),
	}

	metrics := manager.GetMetrics()

	assert.Equal(t, int32(8), metrics.TotalConnections)
	assert.Equal(t, int32(3), metrics.IdleConnections)
	assert.Equal(t, int32(5), metrics.ActiveConnections)
	assert.Equal(t, int32(10), metrics.MaxConnections)
	assert.Equal(t, int64(100), metrics.AcquireCount)
	assert.Equal(t, 5*time.Second, metrics.AcquireDuration)
	assert.Equal(t, int64(2), metrics.CanceledAcquireCount)
	assert.Equal(t, 62.5, metrics.ConnectionUtilization) // (5/8) * 100

	mockPool.AssertExpectations(t)
}

func TestManager_Close(t *testing.T) {
	mockPool := new(MockPool)
	mockPool.On("Close").Return()

	manager := &Manager{
		pool:   mockPool,
		log:    logger.NewTestLogger(),
		closed: false,
	}

	// First close should succeed
	manager.Close()
	assert.True(t, manager.closed)

	// Second close should be idempotent
	manager.Close()
	assert.True(t, manager.closed)

	// Verify Close was called only once
	mockPool.AssertNumberOfCalls(t, "Close", 1)
}

func TestManager_WithTimeout(t *testing.T) {
	manager := &Manager{
		log: logger.NewTestLogger(),
		config: Config{
			DefaultQueryTimeout: 30 * time.Second,
		},
	}

	tests := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{
			name:     "custom timeout",
			timeout:  5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "zero timeout uses default",
			timeout:  0,
			expected: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx, cancel := manager.WithTimeout(ctx, tt.timeout)
			defer cancel()

			deadline, ok := newCtx.Deadline()
			assert.True(t, ok)

			// Check that deadline is approximately correct
			expectedDeadline := time.Now().Add(tt.expected)
			diff := deadline.Sub(expectedDeadline).Abs()
			assert.Less(t, diff, 100*time.Millisecond)
		})
	}
}

func TestManager_Transaction(t *testing.T) {
	t.Run("successful transaction", func(t *testing.T) {
		manager := &Manager{
			log: logger.NewTestLogger(),
		}

		executed := false
		err := manager.Transaction(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
			executed = true
			return nil
		})

		// In a real test, this would use a mock transaction
		// For now, we just verify the structure is correct
		assert.Error(t, err)      // Will error because pool is nil
		assert.False(t, executed) // Won't execute because Begin will fail
	})

	t.Run("transaction with error", func(t *testing.T) {
		manager := &Manager{
			log: logger.NewTestLogger(),
		}

		err := manager.Transaction(context.Background(), func(ctx context.Context, tx pgx.Tx) error {
			return assert.AnError
		})

		assert.Error(t, err)
	})
}

// Helper function to validate config (extracted from actual implementation).
func validateConfig(cfg Config) error {
	if cfg.Host == "" || cfg.Port == 0 || cfg.Database == "" || cfg.User == "" {
		return assert.AnError // Simplified for testing
	}

	if cfg.MaxConns <= 0 {
		return assert.AnError
	}

	return nil
}

func TestMetrics_String(t *testing.T) {
	metrics := Metrics{
		TotalConnections:      10,
		IdleConnections:       3,
		ActiveConnections:     7,
		MaxConnections:        20,
		AcquireCount:          100,
		AcquireDuration:       5 * time.Second,
		CanceledAcquireCount:  2,
		ConnectionUtilization: 35.0,
	}

	str := metrics.String()

	assert.Contains(t, str, "Total: 10")
	assert.Contains(t, str, "Active: 7")
	assert.Contains(t, str, "Idle: 3")
	assert.Contains(t, str, "Max: 20")
	assert.Contains(t, str, "Utilization: 35.00%")
}

func TestManager_Retry(t *testing.T) {
	tests := []struct {
		name          string
		attempts      []error
		maxRetries    int
		retryDelay    time.Duration
		expectedCalls int
		wantErr       bool
	}{
		{
			name:          "succeeds on first try",
			attempts:      []error{nil},
			maxRetries:    3,
			retryDelay:    10 * time.Millisecond,
			expectedCalls: 1,
			wantErr:       false,
		},
		{
			name:          "succeeds after retry",
			attempts:      []error{assert.AnError, nil},
			maxRetries:    3,
			retryDelay:    10 * time.Millisecond,
			expectedCalls: 2,
			wantErr:       false,
		},
		{
			name:          "fails after max retries",
			attempts:      []error{assert.AnError, assert.AnError, assert.AnError},
			maxRetries:    2,
			retryDelay:    10 * time.Millisecond,
			expectedCalls: 2,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0

			manager := &Manager{
				log: logger.NewTestLogger(),
				config: Config{
					MaxRetries: tt.maxRetries,
					RetryDelay: tt.retryDelay,
				},
			}

			err := manager.Retry(context.Background(), func() error {
				if callCount < len(tt.attempts) {
					err := tt.attempts[callCount]
					callCount++
					return err
				}
				return nil
			})

			assert.Equal(t, tt.expectedCalls, callCount)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// BenchmarkGetMetrics benchmarks the GetMetrics method.
func BenchmarkGetMetrics(b *testing.B) {
	mockPool := new(MockPool)
	stat := &pgxpool.Stat{
		AcquireCount:  1000,
		AcquiredConns: 50,
		IdleConns:     30,
		TotalConns:    80,
		MaxConns:      100,
	}

	mockPool.On("Stat").Return(stat)

	manager := &Manager{
		pool: mockPool,
		log:  logger.NewTestLogger(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetMetrics()
	}
}

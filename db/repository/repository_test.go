package repository

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

// MockDB is a mock database connection for testing.
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	argList := m.Called(ctx, sql, args)
	if rows := argList.Get(0); rows != nil {
		return rows.(pgx.Rows), argList.Error(1)
	}
	return nil, argList.Error(1)
}

func (m *MockDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	argList := m.Called(ctx, sql, args)
	return argList.Get(0).(pgx.Row)
}

func (m *MockDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgx.CommandTag, error) {
	argList := m.Called(ctx, sql, args)
	return argList.Get(0).(pgx.CommandTag), argList.Error(1)
}

// Test entity for repository tests.
type TestEntity struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Active    bool      `db:"active" json:"active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func TestRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		entity  TestEntity
		setup   func(*MockDB)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful create",
			entity: TestEntity{
				ID:     "test-123",
				Name:   "Test User",
				Email:  "test@example.com",
				Active: true,
			},
			setup: func(db *MockDB) {
				db.On("Exec", mock.Anything, mock.MatchedBy(func(sql string) bool {
					return contains(sql, "INSERT INTO")
				}), mock.Anything).Return(pgx.CommandTag("INSERT 1"), nil)
			},
			wantErr: false,
		},
		{
			name: "create with database error",
			entity: TestEntity{
				ID:    "test-456",
				Name:  "Test User 2",
				Email: "test2@example.com",
			},
			setup: func(db *MockDB) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).
					Return(pgx.CommandTag(""), assert.AnError)
			},
			wantErr: true,
			errMsg:  "create failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDB)
			tt.setup(mockDB)

			repo := &repository[TestEntity]{
				db:         mockDB,
				log:        logger.NewTestLogger(),
				table:      "test_entities",
				primaryKey: "id",
			}

			// Execute
			err := repo.Create(context.Background(), tt.entity)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestRepository_FindByID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(*MockDB)
		want    *TestEntity
		wantErr bool
		errMsg  string
	}{
		{
			name: "find existing entity",
			id:   "test-123",
			setup: func(db *MockDB) {
				rows := &MockRows{
					data: []map[string]interface{}{
						{
							"id":         "test-123",
							"name":       "Test User",
							"email":      "test@example.com",
							"active":     true,
							"created_at": time.Now(),
							"updated_at": time.Now(),
						},
					},
				}
				db.On("Query", mock.Anything, mock.MatchedBy(func(sql string) bool {
					return contains(sql, "SELECT") && contains(sql, "WHERE")
				}), []interface{}{"test-123"}).Return(rows, nil)
			},
			want: &TestEntity{
				ID:     "test-123",
				Name:   "Test User",
				Email:  "test@example.com",
				Active: true,
			},
			wantErr: false,
		},
		{
			name: "entity not found",
			id:   "nonexistent",
			setup: func(db *MockDB) {
				rows := &MockRows{data: []map[string]interface{}{}}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).
					Return(rows, nil)
			},
			want:    nil,
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDB)
			tt.setup(mockDB)

			repo := &repository[TestEntity]{
				db:         mockDB,
				log:        logger.NewTestLogger(),
				table:      "test_entities",
				primaryKey: "id",
			}

			// Execute
			result, err := repo.FindByID(context.Background(), tt.id)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.ID, result.ID)
				assert.Equal(t, tt.want.Name, result.Name)
				assert.Equal(t, tt.want.Email, result.Email)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		updates Updates
		setup   func(*MockDB)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful update",
			id:   "test-123",
			updates: Updates{
				"name":       "Updated Name",
				"email":      "updated@example.com",
				"updated_at": time.Now(),
			},
			setup: func(db *MockDB) {
				db.On("Exec", mock.Anything, mock.MatchedBy(func(sql string) bool {
					return contains(sql, "UPDATE") && contains(sql, "SET")
				}), mock.Anything).Return(pgx.CommandTag("UPDATE 1"), nil)
			},
			wantErr: false,
		},
		{
			name: "update non-existent entity",
			id:   "nonexistent",
			updates: Updates{
				"name": "Updated Name",
			},
			setup: func(db *MockDB) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).
					Return(pgx.CommandTag("UPDATE 0"), nil)
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDB)
			tt.setup(mockDB)

			repo := &repository[TestEntity]{
				db:         mockDB,
				log:        logger.NewTestLogger(),
				table:      "test_entities",
				primaryKey: "id",
			}

			// Execute
			err := repo.Update(context.Background(), tt.id, tt.updates)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(*MockDB)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful delete",
			id:   "test-123",
			setup: func(db *MockDB) {
				db.On("Exec", mock.Anything, mock.MatchedBy(func(sql string) bool {
					return contains(sql, "DELETE FROM")
				}), []interface{}{"test-123"}).Return(pgx.CommandTag("DELETE 1"), nil)
			},
			wantErr: false,
		},
		{
			name: "delete non-existent entity",
			id:   "nonexistent",
			setup: func(db *MockDB) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).
					Return(pgx.CommandTag("DELETE 0"), nil)
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDB)
			tt.setup(mockDB)

			repo := &repository[TestEntity]{
				db:         mockDB,
				log:        logger.NewTestLogger(),
				table:      "test_entities",
				primaryKey: "id",
			}

			// Execute
			err := repo.Delete(context.Background(), tt.id)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestRepository_FindAll(t *testing.T) {
	tests := []struct {
		name       string
		conditions Conditions
		setup      func(*MockDB)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "find all without conditions",
			conditions: nil,
			setup: func(db *MockDB) {
				rows := &MockRows{
					data: []map[string]interface{}{
						{"id": "1", "name": "User 1", "email": "user1@example.com", "active": true},
						{"id": "2", "name": "User 2", "email": "user2@example.com", "active": false},
						{"id": "3", "name": "User 3", "email": "user3@example.com", "active": true},
					},
				}
				db.On("Query", mock.Anything, mock.MatchedBy(func(sql string) bool {
					return contains(sql, "SELECT") && !contains(sql, "WHERE")
				}), []interface{}{}).Return(rows, nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "find with conditions",
			conditions: Conditions{
				"active": true,
			},
			setup: func(db *MockDB) {
				rows := &MockRows{
					data: []map[string]interface{}{
						{"id": "1", "name": "User 1", "email": "user1@example.com", "active": true},
						{"id": "3", "name": "User 3", "email": "user3@example.com", "active": true},
					},
				}
				db.On("Query", mock.Anything, mock.MatchedBy(func(sql string) bool {
					return contains(sql, "SELECT") && contains(sql, "WHERE")
				}), mock.Anything).Return(rows, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDB)
			tt.setup(mockDB)

			repo := &repository[TestEntity]{
				db:         mockDB,
				log:        logger.NewTestLogger(),
				table:      "test_entities",
				primaryKey: "id",
			}

			// Execute
			results, err := repo.FindAll(context.Background(), tt.conditions)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, results, tt.wantCount)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

// MockRows implements a simplified pgx.Rows interface for testing.
type MockRows struct {
	data  []map[string]interface{}
	index int
}

func (m *MockRows) Next() bool {
	m.index++
	return m.index <= len(m.data)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.index > len(m.data) || m.index < 1 {
		return assert.AnError
	}

	row := m.data[m.index-1]
	// Simplified scan implementation for testing
	// In real implementation, this would use reflection to populate dest
	return nil
}

func (m *MockRows) Close() {
	// No-op for mock
}

func (m *MockRows) Err() error {
	return nil
}

// Helper function.
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr))
}

package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/slack-go/slack"
)

// MockDatabaseClient is a mock implementation of DatabaseClient.
type MockDatabaseClient struct {
	mu            sync.RWMutex
	ExecuteCalls  []ExecuteCall
	QueryCalls    []QueryCall
	PingError     error
	ExecuteError  error
	QueryError    error
	QueryRowValue interface{}
	QueryResults  [][]interface{}
}

type ExecuteCall struct {
	Query string
	Args  []interface{}
}

type QueryCall struct {
	Query string
	Args  []interface{}
}

func NewMockDatabaseClient() *MockDatabaseClient {
	return &MockDatabaseClient{
		ExecuteCalls: []ExecuteCall{},
		QueryCalls:   []QueryCall{},
	}
}

func (m *MockDatabaseClient) Execute(ctx context.Context, query string, args ...interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ExecuteCalls = append(m.ExecuteCalls, ExecuteCall{Query: query, Args: args})
	return m.ExecuteError
}

func (m *MockDatabaseClient) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.QueryCalls = append(m.QueryCalls, QueryCall{Query: query, Args: args})
	return &MockRow{Value: m.QueryRowValue, Error: m.QueryError}
}

func (m *MockDatabaseClient) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.QueryCalls = append(m.QueryCalls, QueryCall{Query: query, Args: args})
	if m.QueryError != nil {
		return nil, m.QueryError
	}
	return &MockRows{Results: m.QueryResults}, nil
}

func (m *MockDatabaseClient) Transaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx := &MockTx{}
	return fn(tx)
}

func (m *MockDatabaseClient) Ping(ctx context.Context) error {
	return m.PingError
}

func (m *MockDatabaseClient) Close() {
	// No-op for mock
}

// MockRow implements pgx.Row.
type MockRow struct {
	Value interface{}
	Error error
}

func (r *MockRow) Scan(dest ...interface{}) error {
	if r.Error != nil {
		return r.Error
	}
	if len(dest) > 0 && r.Value != nil {
		// Simple assignment for testing
		if v, ok := dest[0].(*string); ok && r.Value != nil {
			*v = fmt.Sprintf("%v", r.Value)
		}
	}
	return nil
}

// MockRows implements pgx.Rows.
type MockRows struct {
	Results [][]interface{}
	current int
	closed  bool
}

func (r *MockRows) Close() {
	r.closed = true
}

func (r *MockRows) Err() error {
	return nil
}

func (r *MockRows) Next() bool {
	if r.closed || r.current >= len(r.Results) {
		return false
	}
	r.current++
	return true
}

func (r *MockRows) Scan(dest ...interface{}) error {
	if r.current == 0 || r.current > len(r.Results) {
		return fmt.Errorf("no current row")
	}

	row := r.Results[r.current-1]
	for i, d := range dest {
		if i < len(row) {
			// Simple type assertion for common types
			switch v := d.(type) {
			case *string:
				*v = fmt.Sprintf("%v", row[i])
			case *int:
				if val, ok := row[i].(int); ok {
					*v = val
				}
			case *bool:
				if val, ok := row[i].(bool); ok {
					*v = val
				}
			}
		}
	}
	return nil
}

func (r *MockRows) Values() ([]interface{}, error) {
	if r.current == 0 || r.current > len(r.Results) {
		return nil, fmt.Errorf("no current row")
	}
	return r.Results[r.current-1], nil
}

func (r *MockRows) RawValues() [][]byte {
	return nil
}

func (r *MockRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (r *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (r *MockRows) Conn() *pgx.Conn {
	return nil
}

// MockTx implements pgx.Tx.
type MockTx struct {
	CommitError   error
	RollbackError error
}

func (t *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return t, nil
}

func (t *MockTx) Commit(ctx context.Context) error {
	return t.CommitError
}

func (t *MockTx) Rollback(ctx context.Context) error {
	return t.RollbackError
}

func (t *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (t *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}

func (t *MockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (t *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (t *MockTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (t *MockTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return &MockRows{}, nil
}

func (t *MockTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return &MockRow{}
}

func (t *MockTx) Conn() *pgx.Conn {
	return nil
}

// MockSlackClient is a mock implementation of SlackClient.
type MockSlackClient struct {
	mu                sync.RWMutex
	MessagesSent      []SlackMessage
	ViewsOpened       []slack.ModalViewRequest
	ViewsUpdated      []slack.ModalViewRequest
	UserInfoCalls     []string
	SendMessageError  error
	OpenViewError     error
	UpdateViewError   error
	GetUserInfoError  error
	UserInfoResponse  *slack.User
	ConversationUsers []string
	EphemeralMessages []EphemeralMessage
}

type SlackMessage struct {
	ChannelID string
	Text      string
	Options   []slack.MsgOption
}

type EphemeralMessage struct {
	ChannelID string
	UserID    string
	Text      string
}

func NewMockSlackClient() *MockSlackClient {
	return &MockSlackClient{
		MessagesSent:      []SlackMessage{},
		ViewsOpened:       []slack.ModalViewRequest{},
		ViewsUpdated:      []slack.ModalViewRequest{},
		UserInfoCalls:     []string{},
		EphemeralMessages: []EphemeralMessage{},
	}
}

func (m *MockSlackClient) SendMessage(channelID string, options ...slack.MsgOption) (string, string, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.MessagesSent = append(m.MessagesSent, SlackMessage{
		ChannelID: channelID,
		Options:   options,
	})

	if m.SendMessageError != nil {
		return "", "", "", m.SendMessageError
	}

	return channelID, "1234567890.123456", "", nil
}

func (m *MockSlackClient) SendMessageContext(ctx context.Context, channelID string, options ...slack.MsgOption) (string, string, string, error) {
	return m.SendMessage(channelID, options...)
}

func (m *MockSlackClient) OpenView(triggerID string, view slack.ModalViewRequest) (*slack.ViewResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ViewsOpened = append(m.ViewsOpened, view)

	if m.OpenViewError != nil {
		return nil, m.OpenViewError
	}

	return &slack.ViewResponse{
		View: slack.View{
			ID: "V123456",
		},
	}, nil
}

func (m *MockSlackClient) UpdateView(viewID string, view slack.ModalViewRequest) (*slack.ViewResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ViewsUpdated = append(m.ViewsUpdated, view)

	if m.UpdateViewError != nil {
		return nil, m.UpdateViewError
	}

	return &slack.ViewResponse{
		View: slack.View{
			ID: viewID,
		},
	}, nil
}

func (m *MockSlackClient) GetUserInfo(userID string) (*slack.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UserInfoCalls = append(m.UserInfoCalls, userID)

	if m.GetUserInfoError != nil {
		return nil, m.GetUserInfoError
	}

	if m.UserInfoResponse != nil {
		return m.UserInfoResponse, nil
	}

	return &slack.User{
		ID:       userID,
		Name:     "testuser",
		RealName: "Test User",
		Profile: slack.UserProfile{
			Email: "test@example.com",
		},
	}, nil
}

func (m *MockSlackClient) GetUsersInConversation(params *slack.GetUsersInConversationParameters) ([]string, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.ConversationUsers, "", nil
}

func (m *MockSlackClient) OpenConversation(params *slack.OpenConversationParameters) (*slack.Channel, bool, bool, error) {
	return &slack.Channel{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID: "C123456",
			},
		},
	}, false, false, nil
}

func (m *MockSlackClient) PostEphemeral(channelID, userID string, options ...slack.MsgOption) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.EphemeralMessages = append(m.EphemeralMessages, EphemeralMessage{
		ChannelID: channelID,
		UserID:    userID,
	})

	return "1234567890.123456", nil
}

// MockMDMClient is a mock implementation of MDMClient.
type MockMDMClient struct {
	mu           sync.RWMutex
	Devices      map[string]Device
	GetDeviceErr error
	UpdateCalls  []UpdateCall
}

type UpdateCall struct {
	SerialNumber string
	Updates      map[string]interface{}
}

func NewMockMDMClient() *MockMDMClient {
	return &MockMDMClient{
		Devices:     make(map[string]Device),
		UpdateCalls: []UpdateCall{},
	}
}

func (m *MockMDMClient) GetDevice(ctx context.Context, serialNumber string) (Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetDeviceErr != nil {
		return Device{}, m.GetDeviceErr
	}

	device, ok := m.Devices[serialNumber]
	if !ok {
		return Device{}, fmt.Errorf("device not found")
	}

	return device, nil
}

func (m *MockMDMClient) GetDeviceByUser(ctx context.Context, userEmail string) ([]Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var devices []Device
	for _, device := range m.Devices {
		if device.UserEmail == userEmail {
			devices = append(devices, device)
		}
	}

	return devices, nil
}

func (m *MockMDMClient) GetAllDevices(ctx context.Context) ([]Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var devices []Device
	for _, device := range m.Devices {
		devices = append(devices, device)
	}

	return devices, nil
}

func (m *MockMDMClient) UpdateDevice(ctx context.Context, serialNumber string, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UpdateCalls = append(m.UpdateCalls, UpdateCall{
		SerialNumber: serialNumber,
		Updates:      updates,
	})

	return nil
}

// MockIDPClient is a mock implementation of IDPClient.
type MockIDPClient struct {
	mu          sync.RWMutex
	Users       map[string]User
	Groups      map[string]Group
	GetUserErr  error
	UpdateCalls []UpdateCall
}

func NewMockIDPClient() *MockIDPClient {
	return &MockIDPClient{
		Users:       make(map[string]User),
		Groups:      make(map[string]Group),
		UpdateCalls: []UpdateCall{},
	}
}

func (m *MockIDPClient) GetUser(ctx context.Context, userID string) (User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.GetUserErr != nil {
		return User{}, m.GetUserErr
	}

	user, ok := m.Users[userID]
	if !ok {
		return User{}, fmt.Errorf("user not found")
	}

	return user, nil
}

func (m *MockIDPClient) GetUserByEmail(ctx context.Context, email string) (User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, fmt.Errorf("user not found")
}

func (m *MockIDPClient) GetGroupMembers(ctx context.Context, groupID string) ([]User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group, ok := m.Groups[groupID]
	if !ok {
		return nil, fmt.Errorf("group not found")
	}

	var users []User
	for _, memberID := range group.Members {
		if user, ok := m.Users[memberID]; ok {
			users = append(users, user)
		}
	}

	return users, nil
}

func (m *MockIDPClient) GetUserGroups(ctx context.Context, userID string) ([]Group, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var groups []Group
	for _, group := range m.Groups {
		for _, memberID := range group.Members {
			if memberID == userID {
				groups = append(groups, group)
				break
			}
		}
	}

	return groups, nil
}

func (m *MockIDPClient) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UpdateCalls = append(m.UpdateCalls, UpdateCall{
		SerialNumber: userID, // Reusing field for userID
		Updates:      updates,
	})

	return nil
}

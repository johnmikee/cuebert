package unified

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/repository"
	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Manager handles all task operations using the unified task system
type Manager struct {
	db         *pgxpool.Pool
	log        logger.Logger
	taskRepo   repository.TrackingRepository
	userRepo   repository.UserRepository
	tasks      map[string]Task
	tasksMutex sync.RWMutex
	factories  map[TaskType]TaskFactory
	executors  map[TaskType]TaskExecutor
}

// TaskFactory creates tasks of a specific type
type TaskFactory interface {
	CreateTask(ctx context.Context, params map[string]interface{}) (Task, error)
}

// TaskExecutor handles task-specific execution logic
type TaskExecutor interface {
	Execute(ctx context.Context, task Task) error
	Validate(ctx context.Context, task Task) error
}

// Config holds configuration for the task manager
type Config struct {
	DB       *pgxpool.Pool
	Log      logger.Logger
	UserRepo repository.UserRepository
}

// NewManager creates a new unified task manager
func NewManager(config Config) *Manager {
	return &Manager{
		db:        config.DB,
		log:       config.Log,
		taskRepo:  repository.NewTrackingRepository(config.DB, config.Log, repository.TrackingTypeUpdate),
		userRepo:  config.UserRepo,
		tasks:     make(map[string]Task),
		factories: make(map[TaskType]TaskFactory),
		executors: make(map[TaskType]TaskExecutor),
	}
}

// RegisterTaskType registers a new task type with its factory and executor
func (m *Manager) RegisterTaskType(taskType TaskType, factory TaskFactory, executor TaskExecutor) {
	m.factories[taskType] = factory
	m.executors[taskType] = executor

	m.log.Info().
		Str("task_type", string(taskType)).
		Msg("registered task type")
}

// CreateTask creates a new task of the specified type
func (m *Manager) CreateTask(ctx context.Context, taskType TaskType, params map[string]interface{}) (Task, error) {
	factory, exists := m.factories[taskType]
	if !exists {
		return nil, errors.ValidationError(fmt.Sprintf("unknown task type: %s", taskType))
	}

	task, err := factory.CreateTask(ctx, params)
	if err != nil {
		return nil, errors.InternalError("create task", err)
	}

	// Validate the task
	if err := task.Validate(ctx); err != nil {
		return nil, err
	}

	// Store in memory
	m.tasksMutex.Lock()
	m.tasks[task.GetID()] = task
	m.tasksMutex.Unlock()

	// Persist to database
	if err := m.persistTask(ctx, task); err != nil {
		return nil, errors.DBError("persist task", err)
	}

	m.log.Info().
		Str("task_id", task.GetID()).
		Str("task_type", string(task.GetType())).
		Str("user_id", task.GetUserID()).
		Msg("task created")

	return task, nil
}

// GetTask retrieves a task by ID
func (m *Manager) GetTask(ctx context.Context, taskID string) (Task, error) {
	// Check memory first
	m.tasksMutex.RLock()
	task, exists := m.tasks[taskID]
	m.tasksMutex.RUnlock()

	if exists {
		return task, nil
	}

	// Load from database
	task, err := m.loadTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Cache in memory
	m.tasksMutex.Lock()
	m.tasks[taskID] = task
	m.tasksMutex.Unlock()

	return task, nil
}

// GetTasksForUser retrieves all tasks for a specific user
func (m *Manager) GetTasksForUser(ctx context.Context, userID string) ([]Task, error) {
	// Get tracking records from repository
	trackingRecords, err := m.taskRepo.FindByUserSlackID(ctx, userID)
	if err != nil {
		return nil, errors.DBError("find user tasks", err)
	}

	tasks := make([]Task, 0, len(trackingRecords))
	for _, record := range trackingRecords {
		task, err := m.convertTrackingToTask(record)
		if err != nil {
			m.log.Error().
				Err(err).
				Str("user_id", userID).
				Str("task_name", record.TaskName).
				Msg("failed to convert tracking record to task")
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetPendingTasks retrieves all pending tasks
func (m *Manager) GetPendingTasks(ctx context.Context) ([]Task, error) {
	// Get all incomplete tasks
	trackingRecords, err := m.taskRepo.FindAll(ctx, repository.Conditions{
		"task_complete": false,
	})
	if err != nil {
		return nil, errors.DBError("find pending tasks", err)
	}

	tasks := make([]Task, 0, len(trackingRecords))
	for _, record := range trackingRecords {
		task, err := m.convertTrackingToTask(record)
		if err != nil {
			m.log.Error().
				Err(err).
				Str("task_name", record.TaskName).
				Msg("failed to convert tracking record to task")
			continue
		}

		if task.GetStatus() == TaskStatusPending || task.GetStatus() == TaskStatusInProgress {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}

// ExecuteTask runs a task
func (m *Manager) ExecuteTask(ctx context.Context, taskID string) error {
	task, err := m.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	executor, exists := m.executors[task.GetType()]
	if !exists {
		// Use the task's own Execute method as fallback
		return task.Execute(ctx)
	}

	// Execute with the registered executor
	if err := executor.Execute(ctx, task); err != nil {
		return err
	}

	// Update in database
	return m.updateTask(ctx, task)
}

// ProcessReminders checks and sends reminders for tasks that need them
func (m *Manager) ProcessReminders(ctx context.Context) error {
	tasks, err := m.GetPendingTasks(ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.NeedsReminder() {
			m.log.Info().
				Str("task_id", task.GetID()).
				Str("user_id", task.GetUserID()).
				Msg("task needs reminder")

			// Send reminder (implement your reminder logic here)
			// Update reminder status
			task.SetMetadata("last_reminder_sent", time.Now())
		}
	}

	return nil
}

// ProcessMessages checks and sends first/final messages for tasks
func (m *Manager) ProcessMessages(ctx context.Context) error {
	tasks, err := m.GetPendingTasks(ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.NeedsFirstMessage() {
			m.log.Info().
				Str("task_id", task.GetID()).
				Str("user_id", task.GetUserID()).
				Msg("task needs first message")

			// Mark first message sent
			if err := m.taskRepo.MarkFirstMessageSent(ctx, task.GetUserID(), task.GetName()); err != nil {
				m.log.Error().
					Err(err).
					Str("task_id", task.GetID()).
					Msg("failed to mark first message sent")
			}
		}

		if task.NeedsFinalMessage() {
			m.log.Info().
				Str("task_id", task.GetID()).
				Str("user_id", task.GetUserID()).
				Msg("task needs final message")

			// Mark final message sent
			if err := m.taskRepo.MarkFinalMessageSent(ctx, task.GetUserID(), task.GetName()); err != nil {
				m.log.Error().
					Err(err).
					Str("task_id", task.GetID()).
					Msg("failed to mark final message sent")
			}
		}
	}

	return nil
}

// persistTask saves a task to the database
func (m *Manager) persistTask(ctx context.Context, task Task) error {
	// Convert task to tracking record
	tracking := m.convertTaskToTracking(task)

	// Save to database
	return m.taskRepo.Create(ctx, tracking)
}

// updateTask updates a task in the database
func (m *Manager) updateTask(ctx context.Context, task Task) error {
	// Convert task to tracking record
	tracking := m.convertTaskToTracking(task)

	// Update in database
	return m.taskRepo.Update(ctx, task.GetID(), tracking)
}

// loadTask loads a task from the database
func (m *Manager) loadTask(ctx context.Context, taskID string) (Task, error) {
	// This is a simplified version - you'd need to load from the appropriate table
	// based on task type and reconstruct the proper task object
	tracking, err := m.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, errors.DBError("load task", err)
	}

	return m.convertTrackingToTask(*tracking)
}

// convertTaskToTracking converts a task to a tracking record
func (m *Manager) convertTaskToTracking(task Task) repository.Tracking {
	baseTask := task.(*BaseTask) // Assumes all tasks embed BaseTask

	return repository.Tracking{
		UserServiceID:      task.GetUserID(),
		UserSlackID:        task.GetUserID(),
		UserEmail:          task.GetUserEmail(),
		UserFullName:       baseTask.UserFullName,
		TaskName:           task.GetName(),
		TaskComplete:       task.IsComplete(),
		FirstMessageSent:   baseTask.FirstMessageSent,
		FirstMessageSentAt: baseTask.FirstMessageSentAt,
		FinalMessageSent:   baseTask.FinalMessageSent,
		FinalMessageSentAt: baseTask.FinalMessageSentAt,
		ReminderWaiting:    task.NeedsReminder(),
		UpdatedAt:          baseTask.UpdatedAt,
		CreatedAt:          baseTask.CreatedAt,
	}
}

// convertTrackingToTask converts a tracking record to a task
func (m *Manager) convertTrackingToTask(tracking repository.Tracking) (Task, error) {
	// This is a simplified version - you'd need to determine the task type
	// and create the appropriate task object

	// For now, create a base task
	baseTask := &BaseTask{
		ID:                 fmt.Sprintf("%s-%s", tracking.UserSlackID, tracking.TaskName),
		Type:               TaskTypeUpdate, // Would need to determine from metadata
		Name:               tracking.TaskName,
		UserID:             tracking.UserSlackID,
		UserEmail:          tracking.UserEmail,
		UserFullName:       tracking.UserFullName,
		Status:             m.determineStatus(tracking),
		CreatedAt:          tracking.CreatedAt,
		UpdatedAt:          tracking.UpdatedAt,
		FirstMessageSent:   tracking.FirstMessageSent,
		FirstMessageSentAt: &tracking.FirstMessageSentAt,
		FinalMessageSent:   tracking.FinalMessageSent,
		FinalMessageSentAt: &tracking.FinalMessageSentAt,
		log:                m.log,
	}

	if tracking.TaskComplete {
		baseTask.CompletedAt = &tracking.UpdatedAt
	}

	return baseTask, nil
}

// determineStatus determines task status from tracking record
func (m *Manager) determineStatus(tracking repository.Tracking) TaskStatus {
	if tracking.TaskComplete {
		return TaskStatusCompleted
	}
	if tracking.FirstMessageSent {
		return TaskStatusInProgress
	}
	return TaskStatusPending
}

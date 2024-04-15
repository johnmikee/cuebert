package client

import (
	"github.com/johnmikee/cuebert/cuebert/task"
	"github.com/johnmikee/cuebert/cuebert/task/compliance"
	"github.com/johnmikee/cuebert/cuebert/task/update"
)

// Config represents the configuration for the task client.
type Config struct {
	Rabbit task.Rabbit
}

// Task represents the task client.
type TaskConfig struct {
	Task   task.Task
	Config *task.Config
}

// New creates a new task provider based on the provided task configuration.
func New(t *TaskConfig) task.Rabbit {
	config := Config{
		Rabbit: createTaskProvider(t),
	}

	config.Rabbit.Setup(t.Config)

	return config.Rabbit
}

// createTaskProvider creates and returns a task provider based on the provided task type.
func createTaskProvider(t *TaskConfig) task.Rabbit {
	switch t.Task.Option {
	case task.Update:
		return &update.Client{}
	case task.Compliance:
		return &compliance.Client{}
	}
	return nil
}

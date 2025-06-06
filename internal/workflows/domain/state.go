package domain

import (
	"context"
	"time"

	"getnoti.com/pkg/db"
)

// WorkflowState represents the current state of a workflow execution
type WorkflowState struct {
    ExecutionID      string
    State           []byte
    IsCheckpoint    bool
    CreatedAt       time.Time
    UpdatedAt       time.Time
    CurrentStepIndex int
    StepStates      map[string]interface{}
    Checkpoint      interface{}
    Version         int
    LastUpdated     time.Time
}

// StateManager handles workflow execution state persistence
type StateManager interface {
	SaveState(ctx context.Context, state *WorkflowState, tx db.Transaction) error
	GetState(ctx context.Context, executionID string) (*WorkflowState, error)
	ClearState(ctx context.Context, executionID string, tx db.Transaction) error
}

// RetryPolicy defines how workflow steps should be retried
type RetryPolicy struct {
	MaxAttempts     int           `json:"maxAttempts"`
	InitialInterval time.Duration `json:"initialInterval"`
	MaxInterval     time.Duration `json:"maxInterval"`
	Multiplier      float64       `json:"multiplier"`
	MaxElapsedTime  time.Duration `json:"maxElapsedTime"`
}

// RetryManager handles retry logic for workflow steps
type RetryManager interface {
	ShouldRetry(ctx context.Context, executionID string, stepID string) (bool, time.Time, error)
	RecordAttempt(ctx context.Context, executionID string, stepID string, err error, tx db.Transaction) error
	GetNextRetryTime(ctx context.Context, executionID string, stepID string) (time.Time, error)
}

// RecoveryManager handles workflow recovery after failures
type RecoveryManager interface {
	RecoverExecution(ctx context.Context, executionID string) error
	CheckpointExecution(ctx context.Context, executionID string, checkpoint interface{}, tx db.Transaction) error
	GetFailedExecutions(ctx context.Context, tenantID string) ([]string, error)
}

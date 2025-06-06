package scheduler

import (
	"time"

	"github.com/google/uuid"
)

// ScheduleType represents the type of a scheduled task
type ScheduleType string

const (
	ScheduleTypeNotification ScheduleType = "notification"
	ScheduleTypeWebhook      ScheduleType = "webhook"
	ScheduleTypeCleanup      ScheduleType = "cleanup"
	ScheduleTypeReport       ScheduleType = "report"
)

// ScheduleStatus represents the status of a schedule
type ScheduleStatus string

const (
	ScheduleStatusActive    ScheduleStatus = "active"
	ScheduleStatusPaused    ScheduleStatus = "paused"
	ScheduleStatusCompleted ScheduleStatus = "completed"
	ScheduleStatusFailed    ScheduleStatus = "failed"
)

// Schedule represents a scheduled task
type Schedule struct {
	ID                uuid.UUID      `json:"id" db:"id"`
	TenantID          string         `json:"tenant_id" db:"tenant_id"`
	Type              ScheduleType   `json:"type" db:"type"`
	Name              string         `json:"name" db:"name"`
	Description       string         `json:"description" db:"description"`
	CronExpression    string         `json:"cron_expression" db:"cron_expression"`
	Payload           string         `json:"payload" db:"payload"` // JSON payload
	Config            interface{}    `json:"config" db:"config"` // Configuration for the specific schedule type
	HandlerType       string         `json:"handler_type" db:"handler_type"`
	Status            ScheduleStatus `json:"status" db:"status"`
	IsActive          bool           `json:"is_active" db:"is_active"`
	MaxRetries        int            `json:"max_retries" db:"max_retries"`
	TimeoutSecs       int            `json:"timeout_secs" db:"timeout_secs"`
	NextRunAt         *time.Time     `json:"next_run_at" db:"next_run_at"`
	LastRunAt         *time.Time     `json:"last_run_at" db:"last_run_at"`
	NextExecutionAt   *time.Time     `json:"next_execution_at" db:"next_execution_at"`
	LastExecutionAt   *time.Time     `json:"last_execution_at" db:"last_execution_at"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at" db:"updated_at"`
	CreatedBy         string         `json:"created_by" db:"created_by"`
	UpdatedBy         string         `json:"updated_by" db:"updated_by"`
}

// ScheduleExecutionStatus represents the status of a schedule execution
type ScheduleExecutionStatus string

const (
	ExecutionStatusPending   ScheduleExecutionStatus = "pending"
	ExecutionStatusRunning   ScheduleExecutionStatus = "running"
	ExecutionStatusCompleted ScheduleExecutionStatus = "completed"
	ExecutionStatusFailed    ScheduleExecutionStatus = "failed"
	ExecutionStatusTimeout   ScheduleExecutionStatus = "timeout"
	ExecutionStatusRetrying  ScheduleExecutionStatus = "retrying"
)

// ScheduleExecution represents an execution instance of a schedule
type ScheduleExecution struct {
	ID              uuid.UUID               `json:"id" db:"id"`
	ScheduleID      uuid.UUID               `json:"schedule_id" db:"schedule_id"`
	TenantID        string                  `json:"tenant_id" db:"tenant_id"`
	Status          ScheduleExecutionStatus `json:"status" db:"status"`
	StartedAt       *time.Time              `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time              `json:"completed_at" db:"completed_at"`
	Duration        *int64                  `json:"duration" db:"duration"` // Duration in milliseconds
	RetryCount      int                     `json:"retry_count" db:"retry_count"`
	ErrorMessage    *string                 `json:"error_message" db:"error_message"`
	ErrorCode       *string                 `json:"error_code" db:"error_code"`
	Output          *string                 `json:"output" db:"output"`
	Result          *string                 `json:"result" db:"result"`
	ScheduledFor    time.Time               `json:"scheduled_for" db:"scheduled_for"`
	CreatedAt       time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at" db:"updated_at"`
}

// IsRetryable checks if the execution can be retried
func (se *ScheduleExecution) IsRetryable(maxRetries int) bool {
	return se.Status == ExecutionStatusFailed && se.RetryCount < maxRetries
}

// CanTimeout checks if the execution has exceeded the timeout
func (se *ScheduleExecution) CanTimeout(timeoutSecs int) bool {
	if se.StartedAt == nil || timeoutSecs <= 0 {
		return false
	}
	return time.Since(*se.StartedAt).Seconds() > float64(timeoutSecs)
}

package scheduler

import (
	"context"
	"time"
)

// Repository defines the interface for scheduler persistence operations
type Repository interface {
	// Schedule CRUD operations
	CreateSchedule(ctx context.Context, schedule *Schedule) (*Schedule, error)
	GetScheduleByID(ctx context.Context, scheduleID string) (*Schedule, error)
	GetSchedulesByTenantID(ctx context.Context, tenantID string) ([]*Schedule, error)
	GetActiveSchedules(ctx context.Context) ([]*Schedule, error)
	GetSchedulesDueForExecution(ctx context.Context, beforeTime time.Time, limit int) ([]*Schedule, error)
	UpdateSchedule(ctx context.Context, schedule *Schedule) (*Schedule, error)
	DeleteSchedule(ctx context.Context, scheduleID string) error
	
	// Schedule status operations
	SetScheduleActive(ctx context.Context, scheduleID string, active bool) error
	UpdateScheduleLastExecution(ctx context.Context, scheduleID string, executionTime time.Time) error
	UpdateScheduleNextExecution(ctx context.Context, scheduleID string, nextExecutionTime time.Time) error
	
	// Execution operations
	CreateExecution(ctx context.Context, execution *ScheduleExecution) (*ScheduleExecution, error)
	GetExecutionByID(ctx context.Context, executionID string) (*ScheduleExecution, error)
	GetExecutionsByScheduleID(ctx context.Context, scheduleID string, limit, offset int) ([]*ScheduleExecution, int64, error)
	GetPendingExecutions(ctx context.Context, limit int) ([]*ScheduleExecution, error)
	UpdateExecution(ctx context.Context, execution *ScheduleExecution) (*ScheduleExecution, error)
	
	// Analytics operations
	GetScheduleStats(ctx context.Context) (*ScheduleStats, error)
}

// ScheduleStats contains statistics about scheduler operations
type ScheduleStats struct {
	TotalSchedules     int64 `json:"total_schedules"`
	ActiveSchedules    int64 `json:"active_schedules"`
	InactiveSchedules  int64 `json:"inactive_schedules"`
	TotalExecutions    int64 `json:"total_executions"`
	SuccessfulExecutions int64 `json:"successful_executions"`
	FailedExecutions   int64 `json:"failed_executions"`
	PendingExecutions  int64 `json:"pending_executions"`
}
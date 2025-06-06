package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"getnoti.com/pkg/db"
	"getnoti.com/pkg/logger"
)

type WorkflowScheduler struct {
	db     db.Database
	logger logger.Logger
}

type Schedule struct {
	ID              string          `json:"id"`
	WorkflowID      string          `json:"workflow_id"`
	TenantID        string          `json:"tenant_id"`
	CronExpression  string          `json:"cron_expression"`
	Payload         json.RawMessage `json:"payload"`
	IsActive        bool            `json:"is_active"`
	LastExecutionAt *time.Time      `json:"last_execution_at"`
	NextExecutionAt time.Time       `json:"next_execution_at"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func NewWorkflowScheduler(database db.Database, logger logger.Logger) *WorkflowScheduler {
	return &WorkflowScheduler{
		db:     database,
		logger: logger,
	}
}

// CreateSchedule creates a new workflow schedule
func (s *WorkflowScheduler) CreateSchedule(ctx context.Context, schedule *Schedule) error {
	query := `
		INSERT INTO workflow_schedules (
			id, workflow_id, tenant_id, cron_expression, payload, is_active,
			next_execution_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(ctx, query,
		schedule.ID,
		schedule.WorkflowID,
		schedule.TenantID,
		schedule.CronExpression,
		schedule.Payload,
		schedule.IsActive,
		schedule.NextExecutionAt,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create workflow schedule: %w", err)
	}

	return nil
}

// GetPendingSchedules retrieves schedules that are due for execution
func (s *WorkflowScheduler) GetPendingSchedules(ctx context.Context, limit int) ([]*Schedule, error) {
	query := `
		SELECT id, workflow_id, tenant_id, cron_expression, payload, is_active,
			last_execution_at, next_execution_at, created_at, updated_at
		FROM workflow_schedules
		WHERE is_active = true AND next_execution_at <= ?
		ORDER BY next_execution_at ASC
		LIMIT ?
	`

	rows, err := s.db.Query(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*Schedule
	for rows.Next() {
		var schedule Schedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.WorkflowID,
			&schedule.TenantID,
			&schedule.CronExpression,
			&schedule.Payload,
			&schedule.IsActive,
			&schedule.LastExecutionAt,
			&schedule.NextExecutionAt,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

// UpdateScheduleExecution updates the execution times for a schedule
func (s *WorkflowScheduler) UpdateScheduleExecution(ctx context.Context, scheduleID string, lastExecution time.Time, nextExecution time.Time) error {
	query := `
		UPDATE workflow_schedules
		SET last_execution_at = ?,
			next_execution_at = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(ctx, query, lastExecution, nextExecution, time.Now(), scheduleID)
	if err != nil {
		return fmt.Errorf("failed to update schedule execution: %w", err)
	}

	return nil
}

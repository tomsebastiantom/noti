package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"getnoti.com/pkg/db"
	"github.com/google/uuid"
)

type sqlSchedulerRepository struct {
	db db.Database
}

// NewSchedulerRepository creates a new SQL scheduler repository
func NewSchedulerRepository(database db.Database) Repository {
	return &sqlSchedulerRepository{db: database}
}

// CreateSchedule creates a new schedule in the database
func (r *sqlSchedulerRepository) CreateSchedule(ctx context.Context, schedule *Schedule) (*Schedule, error) {
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
	}
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()

	configJSON, err := json.Marshal(schedule.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO schedules (id, tenant_id, name, type, cron_expression, config, is_active, 
		                      last_execution_at, next_execution_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(ctx, query,
		schedule.ID,
		schedule.TenantID,
		schedule.Name,
		schedule.Type,
		schedule.CronExpression,
		string(configJSON),
		schedule.IsActive,
		schedule.LastExecutionAt,
		schedule.NextExecutionAt,
		schedule.CreatedAt,
		schedule.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	return schedule, nil
}

// GetScheduleByID retrieves a schedule by ID
func (r *sqlSchedulerRepository) GetScheduleByID(ctx context.Context, scheduleID string) (*Schedule, error) {
	query := `
		SELECT id, tenant_id, name, type, cron_expression, config, is_active, 
		       last_execution_at, next_execution_at, created_at, updated_at
		FROM schedules
		WHERE id = ?
	`

	row := r.db.QueryRow(ctx, query, scheduleID)
	return r.scanSchedule(row)
}

// GetSchedulesByTenantID retrieves all schedules for a specific tenant
func (r *sqlSchedulerRepository) GetSchedulesByTenantID(ctx context.Context, tenantID string) ([]*Schedule, error) {
	query := `
		SELECT id, tenant_id, name, type, cron_expression, config, is_active, 
		       last_execution_at, next_execution_at, created_at, updated_at
		FROM schedules
		WHERE tenant_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query schedules for tenant %s: %w", tenantID, err)
	}
	defer rows.Close()

	var schedules []*Schedule
	for rows.Next() {
		schedule, err := r.scanSchedule(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule rows: %w", err)
	}

	return schedules, nil
}

// GetActiveSchedules retrieves all active schedules
func (r *sqlSchedulerRepository) GetActiveSchedules(ctx context.Context) ([]*Schedule, error) {
	query := `
		SELECT id, tenant_id, name, type, cron_expression, config, is_active, 
		       last_execution_at, next_execution_at, created_at, updated_at
		FROM schedules
		WHERE is_active = true
		ORDER BY next_execution_at ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*Schedule
	for rows.Next() {
		schedule, err := r.scanSchedule(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule rows: %w", err)
	}

	return schedules, nil
}

// GetSchedulesDueForExecution retrieves schedules that are due for execution
func (r *sqlSchedulerRepository) GetSchedulesDueForExecution(ctx context.Context, beforeTime time.Time, limit int) ([]*Schedule, error) {
	query := `
		SELECT id, tenant_id, name, type, cron_expression, config, is_active, 
		       last_execution_at, next_execution_at, created_at, updated_at
		FROM schedules
		WHERE is_active = true AND next_execution_at <= ?
		ORDER BY next_execution_at ASC
		LIMIT ?
	`

	rows, err := r.db.Query(ctx, query, beforeTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query schedules due for execution: %w", err)
	}
	defer rows.Close()

	var schedules []*Schedule
	for rows.Next() {
		schedule, err := r.scanSchedule(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule rows: %w", err)
	}

	return schedules, nil
}

// UpdateSchedule updates an existing schedule
func (r *sqlSchedulerRepository) UpdateSchedule(ctx context.Context, schedule *Schedule) (*Schedule, error) {
	schedule.UpdatedAt = time.Now()

	configJSON, err := json.Marshal(schedule.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		UPDATE schedules 
		SET name = ?, type = ?, cron_expression = ?, config = ?, is_active = ?, 
		    last_execution_at = ?, next_execution_at = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = r.db.Exec(ctx, query,
		schedule.Name,
		schedule.Type,
		schedule.CronExpression,
		string(configJSON),
		schedule.IsActive,
		schedule.LastExecutionAt,
		schedule.NextExecutionAt,
		schedule.UpdatedAt,
		schedule.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	return schedule, nil
}

// DeleteSchedule deletes a schedule
func (r *sqlSchedulerRepository) DeleteSchedule(ctx context.Context, scheduleID string) error {
	query := `DELETE FROM schedules WHERE id = ?`
	_, err := r.db.Exec(ctx, query, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}
	return nil
}

// SetScheduleActive sets the active status of a schedule
func (r *sqlSchedulerRepository) SetScheduleActive(ctx context.Context, scheduleID string, active bool) error {
	now := time.Now()
	query := `UPDATE schedules SET is_active = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(ctx, query, active, now, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to set schedule active status: %w", err)
	}
	return nil
}

// UpdateScheduleLastExecution updates the last execution timestamp
func (r *sqlSchedulerRepository) UpdateScheduleLastExecution(ctx context.Context, scheduleID string, executionTime time.Time) error {
	now := time.Now()
	query := `UPDATE schedules SET last_execution_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(ctx, query, executionTime, now, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to update last execution: %w", err)
	}
	return nil
}

// UpdateScheduleNextExecution updates the next execution timestamp
func (r *sqlSchedulerRepository) UpdateScheduleNextExecution(ctx context.Context, scheduleID string, nextExecutionTime time.Time) error {
	now := time.Now()
	query := `UPDATE schedules SET next_execution_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(ctx, query, nextExecutionTime, now, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to update next execution: %w", err)
	}
	return nil
}

// CreateExecution creates a new schedule execution record
func (r *sqlSchedulerRepository) CreateExecution(ctx context.Context, execution *ScheduleExecution) (*ScheduleExecution, error) {
	if execution.ID == uuid.Nil {
		execution.ID = uuid.New()
	}
	execution.CreatedAt = time.Now()
	execution.UpdatedAt = time.Now()

	query := `
		INSERT INTO schedule_executions (id, schedule_id, status, started_at, completed_at, 
		                               error_message, result, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(ctx, query,
		execution.ID,
		execution.ScheduleID,
		execution.Status,
		execution.StartedAt,
		execution.CompletedAt,
		execution.ErrorMessage,
		execution.Result,
		execution.CreatedAt,
		execution.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	return execution, nil
}

// GetExecutionByID retrieves an execution by ID
func (r *sqlSchedulerRepository) GetExecutionByID(ctx context.Context, executionID string) (*ScheduleExecution, error) {
	query := `
		SELECT id, schedule_id, status, started_at, completed_at, error_message, 
		       result, created_at, updated_at
		FROM schedule_executions
		WHERE id = ?
	`

	row := r.db.QueryRow(ctx, query, executionID)
	return r.scanExecution(row)
}

// GetExecutionsByScheduleID retrieves executions for a specific schedule with pagination
func (r *sqlSchedulerRepository) GetExecutionsByScheduleID(ctx context.Context, scheduleID string, limit, offset int) ([]*ScheduleExecution, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM schedule_executions WHERE schedule_id = ?`
	var total int64
	err := r.db.QueryRow(ctx, countQuery, scheduleID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get execution count: %w", err)
	}

	// Get executions with pagination
	query := `
		SELECT id, schedule_id, status, started_at, completed_at, error_message, 
		       result, created_at, updated_at
		FROM schedule_executions
		WHERE schedule_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(ctx, query, scheduleID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query executions: %w", err)
	}
	defer rows.Close()

	var executions []*ScheduleExecution
	for rows.Next() {
		execution, err := r.scanExecution(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan execution: %w", err)
		}
		executions = append(executions, execution)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating execution rows: %w", err)
	}

	return executions, total, nil
}

// GetPendingExecutions retrieves executions that are pending
func (r *sqlSchedulerRepository) GetPendingExecutions(ctx context.Context, limit int) ([]*ScheduleExecution, error) {
	query := `
		SELECT id, schedule_id, status, started_at, completed_at, error_message, 
		       result, created_at, updated_at
		FROM schedule_executions
		WHERE status = ?
		ORDER BY created_at ASC
		LIMIT ?
	`

	rows, err := r.db.Query(ctx, query, ExecutionStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending executions: %w", err)
	}
	defer rows.Close()

	var executions []*ScheduleExecution
	for rows.Next() {
		execution, err := r.scanExecution(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}
		executions = append(executions, execution)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating execution rows: %w", err)
	}

	return executions, nil
}

// UpdateExecution updates an existing execution
func (r *sqlSchedulerRepository) UpdateExecution(ctx context.Context, execution *ScheduleExecution) (*ScheduleExecution, error) {
	execution.UpdatedAt = time.Now()

	query := `
		UPDATE schedule_executions 
		SET status = ?, started_at = ?, completed_at = ?, error_message = ?, 
		    result = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(ctx, query,
		execution.Status,
		execution.StartedAt,
		execution.CompletedAt,
		execution.ErrorMessage,
		execution.Result,
		execution.UpdatedAt,
		execution.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update execution: %w", err)
	}

	return execution, nil
}

// GetScheduleStats retrieves statistics about scheduler operations
func (r *sqlSchedulerRepository) GetScheduleStats(ctx context.Context) (*ScheduleStats, error) {
	stats := &ScheduleStats{}

	// Get schedule counts
	scheduleQuery := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN is_active = true THEN 1 ELSE 0 END) as active,
			SUM(CASE WHEN is_active = false THEN 1 ELSE 0 END) as inactive
		FROM schedules
	`
	
	err := r.db.QueryRow(ctx, scheduleQuery).Scan(
		&stats.TotalSchedules,
		&stats.ActiveSchedules,
		&stats.InactiveSchedules,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule stats: %w", err)
	}

	// Get execution counts
	executionQuery := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as successful,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as failed,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as pending
		FROM schedule_executions
	`
	
	err = r.db.QueryRow(ctx, executionQuery, 
		ExecutionStatusCompleted, 
		ExecutionStatusFailed, 
		ExecutionStatusPending,
	).Scan(
		&stats.TotalExecutions,
		&stats.SuccessfulExecutions,
		&stats.FailedExecutions,
		&stats.PendingExecutions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution stats: %w", err)
	}

	return stats, nil
}

// scanSchedule scans a schedule from a database row
func (r *sqlSchedulerRepository) scanSchedule(scanner interface{}) (*Schedule, error) {
	var schedule Schedule
	var configJSON string
	var lastExecutionAt, nextExecutionAt sql.NullTime

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&schedule.ID,
			&schedule.TenantID,
			&schedule.Name,
			&schedule.Type,
			&schedule.CronExpression,
			&configJSON,
			&schedule.IsActive,
			&lastExecutionAt,
			&nextExecutionAt,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)
	case *sql.Rows:
		err = s.Scan(
			&schedule.ID,
			&schedule.TenantID,
			&schedule.Name,
			&schedule.Type,
			&schedule.CronExpression,
			&configJSON,
			&schedule.IsActive,
			&lastExecutionAt,
			&nextExecutionAt,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan schedule: %w", err)
	}

	// Parse JSON config
	if err := json.Unmarshal([]byte(configJSON), &schedule.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Handle nullable timestamps
	if lastExecutionAt.Valid {
		schedule.LastExecutionAt = &lastExecutionAt.Time
	}
	if nextExecutionAt.Valid {
		schedule.NextExecutionAt = &nextExecutionAt.Time
	}

	return &schedule, nil
}

// scanExecution scans a schedule execution from a database row
func (r *sqlSchedulerRepository) scanExecution(scanner interface{}) (*ScheduleExecution, error) {
	var execution ScheduleExecution
	var startedAt, completedAt sql.NullTime
	var errorMessage, result sql.NullString

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&execution.ID,
			&execution.ScheduleID,
			&execution.Status,
			&startedAt,
			&completedAt,
			&errorMessage,
			&result,
			&execution.CreatedAt,
			&execution.UpdatedAt,
		)
	case *sql.Rows:
		err = s.Scan(
			&execution.ID,
			&execution.ScheduleID,
			&execution.Status,
			&startedAt,
			&completedAt,
			&errorMessage,
			&result,
			&execution.CreatedAt,
			&execution.UpdatedAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan execution: %w", err)
	}

	// Handle nullable fields
	if startedAt.Valid {
		execution.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		execution.CompletedAt = &completedAt.Time
	}
	if errorMessage.Valid {
		execution.ErrorMessage = &errorMessage.String
	}
	if result.Valid {
		execution.Result = &result.String
	}

	return &execution, nil
}
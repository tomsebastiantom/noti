package implementations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"getnoti.com/internal/workflows/domain"
	"getnoti.com/internal/workflows/repos"
	"getnoti.com/pkg/db"
)

type sqlExecutionRepository struct {
	db db.Database
}

// NewExecutionRepository creates a new SQL execution repository
func NewExecutionRepository(database db.Database) repos.ExecutionRepository {
	return &sqlExecutionRepository{db: database}
}

// CreateExecution creates a new workflow execution
func (r *sqlExecutionRepository) CreateExecution(ctx context.Context, execution *domain.WorkflowExecution, tx db.Transaction) error {
	contextJSON, err := json.Marshal(execution.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal execution context: %w", err)
	}

	stepsJSON, err := json.Marshal(execution.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal execution steps: %w", err)
	}

	query := `
		INSERT INTO workflow_executions (id, workflow_id, tenant_id, trigger_id, status, payload, context, steps, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(ctx, query,
		execution.ID,
		execution.WorkflowID,
		execution.TenantID,
		execution.TriggerID,
		execution.Status,
		execution.Payload,
		contextJSON,
		stepsJSON,
		execution.CreatedAt,
		execution.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create workflow execution: %w", err)
	}

	return nil
}

// GetExecutionByID retrieves an execution by ID
func (r *sqlExecutionRepository) GetExecutionByID(ctx context.Context, tenantID, executionID string) (*domain.WorkflowExecution, error) {
	query := `
		SELECT id, workflow_id, tenant_id, trigger_id, status, payload, context, steps, created_at, updated_at, started_at, completed_at, error_message
		FROM workflow_executions
		WHERE id = ? AND tenant_id = ?
	`

	row := r.db.QueryRow(ctx, query, executionID, tenantID)
	execution, err := r.scanExecution(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution by ID: %w", err)
	}

	return execution, nil
}

// UpdateExecution updates an existing execution
func (r *sqlExecutionRepository) UpdateExecution(ctx context.Context, execution *domain.WorkflowExecution, tx db.Transaction) error {
	execution.UpdatedAt = time.Now()

	contextJSON, err := json.Marshal(execution.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal execution context: %w", err)
	}

	stepsJSON, err := json.Marshal(execution.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal execution steps: %w", err)
	}
	
	query := `
		UPDATE workflow_executions
		SET status = ?, payload = ?, context = ?, steps = ?, updated_at = ?, started_at = ?, completed_at = ?, error_message = ?
		WHERE id = ? AND tenant_id = ?
	`
	
	var result sql.Result
	
	if tx != nil {
		result, err = tx.Exec(ctx, query,
			execution.Status,
			execution.Payload,
			contextJSON,
			stepsJSON,
			execution.UpdatedAt,
			execution.StartedAt,
			execution.CompletedAt,
			execution.ErrorMessage,
			execution.ID,
			execution.TenantID)
	} else {
		result, err = r.db.Exec(ctx, query,
			execution.Status,
			execution.Payload,
			contextJSON,
			stepsJSON,
			execution.UpdatedAt,
			execution.StartedAt,
			execution.CompletedAt,
			execution.ErrorMessage,
			execution.ID,
			execution.TenantID)
	}

	if err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("execution not found or no changes made")
	}

	return nil
}

// ListExecutions retrieves executions based on filters
func (r *sqlExecutionRepository) ListExecutions(ctx context.Context, tenantID string, filters repos.ExecutionFilters) ([]*domain.WorkflowExecution, error) {
	var args []interface{}
	var conditions []string

	// Base query
	query := `
		SELECT id, workflow_id, tenant_id, trigger_id, status, payload, context, steps, created_at, updated_at, started_at, completed_at, error_message
		FROM workflow_executions
		WHERE tenant_id = ?
	`
	args = append(args, tenantID)

	// Add filters
	if filters.WorkflowID != "" {
		conditions = append(conditions, "workflow_id = ?")
		args = append(args, filters.WorkflowID)
	}

	if filters.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filters.Status)
	}

	if filters.TriggerID != "" {
		conditions = append(conditions, "trigger_id = ?")
		args = append(args, filters.TriggerID)
	}

	// Add conditions to query
	for _, condition := range conditions {
		query += " AND " + condition
	}

	// Add order by and pagination
	query += " ORDER BY created_at DESC"
	
	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	// Execute query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}
	defer rows.Close()

	var executions []*domain.WorkflowExecution
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

// scanExecution scans a workflow execution from a database row
func (r *sqlExecutionRepository) scanExecution(scanner interface{}) (*domain.WorkflowExecution, error) {
	var execution domain.WorkflowExecution
	var contextJSON, stepsJSON []byte
	var startedAt, completedAt sql.NullTime

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&execution.ID,
			&execution.WorkflowID,
			&execution.TenantID,
			&execution.TriggerID,
			&execution.Status,
			&execution.Payload,
			&contextJSON,
			&stepsJSON,
			&execution.CreatedAt,
			&execution.UpdatedAt,
			&startedAt,
			&completedAt,
			&execution.ErrorMessage,
		)
	case *sql.Rows:
		err = s.Scan(
			&execution.ID,
			&execution.WorkflowID,
			&execution.TenantID,
			&execution.TriggerID,
			&execution.Status,
			&execution.Payload,
			&contextJSON,
			&stepsJSON,
			&execution.CreatedAt,
			&execution.UpdatedAt,
			&startedAt,
			&completedAt,
			&execution.ErrorMessage,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan workflow execution: %w", err)
	}

	// Handle nullable timestamps
	if startedAt.Valid {
		execution.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		execution.CompletedAt = &completedAt.Time
	}

	// Unmarshal JSON fields
	if len(contextJSON) > 0 {
		if err := json.Unmarshal(contextJSON, &execution.Context); err != nil {
			return nil, fmt.Errorf("failed to unmarshal execution context: %w", err)
		}
	}

	if len(stepsJSON) > 0 {
		if err := json.Unmarshal(stepsJSON, &execution.Steps); err != nil {
			return nil, fmt.Errorf("failed to unmarshal execution steps: %w", err)
		}
	}

	return &execution, nil
}

// scanStepExecution scans a step execution from a database row
func (r *sqlExecutionRepository) scanStepExecution(scanner interface{}) (*domain.StepExecution, error) {
	var stepExecution domain.StepExecution
	var startedAt, completedAt, scheduledAt sql.NullTime
	var input, output, errorMsg sql.NullString
	var tempTenantID string // Temporary variable for tenant_id not in struct

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&stepExecution.ID,
			&stepExecution.ExecutionID,
			&tempTenantID, // TenantID is stored in DB but not in the domain model
			&stepExecution.StepID,
			&stepExecution.Status,
			&input,
			&output,
			&errorMsg,
			&stepExecution.CreatedAt,
			&stepExecution.UpdatedAt,
			&startedAt,
			&completedAt,
			&scheduledAt,
		)
	case *sql.Rows:
		err = s.Scan(
			&stepExecution.ID,
			&stepExecution.ExecutionID,
			&tempTenantID, // TenantID is stored in DB but not in the domain model
			&stepExecution.StepID,
			&stepExecution.Status,
			&input,
			&output,
			&errorMsg,
			&stepExecution.CreatedAt,
			&stepExecution.UpdatedAt,
			&startedAt,
			&completedAt,
			&scheduledAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan step execution: %w", err)
	}

	// Handle nullable timestamps
	if startedAt.Valid {
		stepExecution.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		stepExecution.CompletedAt = &completedAt.Time
	}
	if scheduledAt.Valid {
		stepExecution.DelayUntil = &scheduledAt.Time
	}
	
	// Map error message from nullable string
	if errorMsg.Valid {
		stepExecution.ErrorMessage = errorMsg.String
	}
	
	// Handle JSON result if needed
	if output.Valid {
		stepExecution.Result = json.RawMessage(output.String)
	}

	return &stepExecution, nil
}

// GetPendingExecutions retrieves workflow executions that are pending or in progress
func (r *sqlExecutionRepository) GetPendingExecutions(ctx context.Context, limit int) ([]*domain.WorkflowExecution, error) {
	query := `
		SELECT id, workflow_id, tenant_id, trigger_id, status, payload, context, steps, created_at, updated_at, started_at, completed_at, error_message
		FROM workflow_executions
		WHERE status IN (?, ?)
		ORDER BY created_at ASC
		LIMIT ?
	`

	rows, err := r.db.Query(ctx, query, 
		domain.ExecutionStatusPending, 
		domain.ExecutionStatusRunning, 
		limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending executions: %w", err)
	}
	defer rows.Close()

	var executions []*domain.WorkflowExecution
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

// CreateStepExecution creates a new workflow step execution
func (r *sqlExecutionRepository) CreateStepExecution(ctx context.Context, stepExecution *domain.StepExecution, tx db.Transaction) error {
	var dbTx db.Transaction
	var err error

	if tx != nil {
		dbTx = tx
	} else {
		dbTx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				dbTx.Rollback()
			}
		}()
	}

	query := `
		INSERT INTO workflow_step_executions (id, execution_id, tenant_id, step_id, status, input, output, error, created_at, updated_at, started_at, completed_at, scheduled_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = dbTx.Exec(ctx, query,
		stepExecution.ID,
		stepExecution.ExecutionID,
		"", // TenantID is not in the domain model, using empty string
		stepExecution.StepID,
		stepExecution.Status,
		nil,  // input - not used
		stepExecution.Result,
		stepExecution.ErrorMessage,
		stepExecution.CreatedAt,
		stepExecution.UpdatedAt,
		stepExecution.StartedAt,
		stepExecution.CompletedAt,
		stepExecution.DelayUntil,
	)

	if err != nil {
		return fmt.Errorf("failed to create step execution: %w", err)
	}

	if tx == nil {
		if err = dbTx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// UpdateStepExecution updates an existing workflow step execution
func (r *sqlExecutionRepository) UpdateStepExecution(ctx context.Context, stepExecution *domain.StepExecution, tx db.Transaction) error {
	var dbTx db.Transaction
	var err error

	if tx != nil {
		dbTx = tx
	} else {
		dbTx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				dbTx.Rollback()
			}
		}()
	}

	stepExecution.UpdatedAt = time.Now()
		query := `
		UPDATE workflow_step_executions
		SET status = ?, input = ?, output = ?, error = ?, updated_at = ?, started_at = ?, completed_at = ?, scheduled_at = ?
		WHERE id = ? AND execution_id = ?
	`

	_, err = dbTx.Exec(ctx, query,
		stepExecution.Status,
		nil, // Input - not in domain model
		stepExecution.Result, // Output maps to Result
		stepExecution.ErrorMessage, // Error maps to ErrorMessage
		stepExecution.UpdatedAt,
		stepExecution.StartedAt,
		stepExecution.CompletedAt,
		stepExecution.DelayUntil, // ScheduledAt maps to DelayUntil
		stepExecution.ID,
		stepExecution.ExecutionID,
	)

	if err != nil {
		return fmt.Errorf("failed to update step execution: %w", err)
	}

	if tx == nil {
		if err = dbTx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// GetPendingStepExecutions retrieves step executions that are ready to be processed
func (r *sqlExecutionRepository) GetPendingStepExecutions(ctx context.Context, limit int) ([]*domain.StepExecution, error) {
	query := `
		SELECT id, execution_id, tenant_id, step_id, status, input, output, error, created_at, updated_at, started_at, completed_at, scheduled_at
		FROM workflow_step_executions
		WHERE status = ? AND (scheduled_at IS NULL OR scheduled_at <= ?)
		ORDER BY created_at ASC
		LIMIT ?
	`
	rows, err := r.db.Query(ctx, query, 
		domain.ExecutionStatusPending, 
		time.Now(),
		limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending step executions: %w", err)
	}
	defer rows.Close()

	var stepExecutions []*domain.StepExecution
	for rows.Next() {
		stepExecution, err := r.scanStepExecution(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan step execution: %w", err)
		}
		stepExecutions = append(stepExecutions, stepExecution)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating step execution rows: %w", err)
	}

	return stepExecutions, nil
}

// GetDelayedStepExecutions retrieves delayed step executions that are ready to be processed
func (r *sqlExecutionRepository) GetDelayedStepExecutions(ctx context.Context, limit int) ([]*domain.StepExecution, error) {
	query := `
		SELECT id, execution_id, tenant_id, step_id, status, input, output, error, created_at, updated_at, started_at, completed_at, scheduled_at
		FROM workflow_step_executions
		WHERE status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?
		ORDER BY scheduled_at ASC
		LIMIT ?
	`
	rows, err := r.db.Query(ctx, query, 
		"delayed", // Using string directly since no constant exists in domain model
		time.Now(),
		limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query delayed step executions: %w", err)
	}
	defer rows.Close()

	var stepExecutions []*domain.StepExecution
	for rows.Next() {
		stepExecution, err := r.scanStepExecution(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan step execution: %w", err)
		}
		stepExecutions = append(stepExecutions, stepExecution)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating step execution rows: %w", err)
	}

	return stepExecutions, nil
}

// GetStepExecutionsByExecutionID retrieves all step executions for a specific workflow execution
func (r *sqlExecutionRepository) GetStepExecutionsByExecutionID(ctx context.Context, executionID string) ([]*domain.StepExecution, error) {
	query := `
		SELECT id, execution_id, tenant_id, step_id, status, input, output, error, created_at, updated_at, started_at, completed_at, scheduled_at
		FROM workflow_step_executions
		WHERE execution_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query step executions by execution ID: %w", err)
	}
	defer rows.Close()

	var stepExecutions []*domain.StepExecution
	for rows.Next() {
		stepExecution, err := r.scanStepExecution(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan step execution: %w", err)
		}
		stepExecutions = append(stepExecutions, stepExecution)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating step execution rows: %w", err)
	}

	return stepExecutions, nil
}

// SaveExecutionState saves a workflow execution state
func (r *sqlExecutionRepository) SaveExecutionState(ctx context.Context, executionID string, state interface{}, checkpoint bool, tx db.Transaction) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal execution state: %w", err)
	}
	
	var dbTx db.Transaction
	if tx != nil {
		dbTx = tx
	} else {
		dbTx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				dbTx.Rollback()
			}
		}()
	}

	// Check if state already exists
	var exists bool
	err = dbTx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM workflow_states WHERE execution_id = ?)", executionID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if state exists: %w", err)
	}

	now := time.Now()
	if exists {
		// Update existing state
		query := `UPDATE workflow_states SET state = ?, is_checkpoint = ?, updated_at = ? WHERE execution_id = ?`
		_, err = dbTx.Exec(ctx, query, stateJSON, checkpoint, now, executionID)
	} else {
		// Insert new state
		query := `INSERT INTO workflow_states (execution_id, state, is_checkpoint, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
		_, err = dbTx.Exec(ctx, query, executionID, stateJSON, checkpoint, now, now)
	}

	if err != nil {
		return fmt.Errorf("failed to save execution state: %w", err)
	}

	if tx == nil {
		if err = dbTx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// GetExecutionState retrieves a workflow execution state
func (r *sqlExecutionRepository) GetExecutionState(ctx context.Context, executionID string) (*domain.WorkflowState, error) {
	query := `
		SELECT execution_id, state, is_checkpoint, created_at, updated_at
		FROM workflow_states
		WHERE execution_id = ?
	`

	row := r.db.QueryRow(ctx, query, executionID)
	
	var state domain.WorkflowState
	var stateJSON []byte
	
	err := row.Scan(
		&state.ExecutionID,
		&stateJSON,
		&state.IsCheckpoint,
		&state.CreatedAt,
		&state.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No state found
		}
		return nil, fmt.Errorf("failed to get execution state: %w", err)
	}
	type WorkflowState struct {
    ExecutionID      string
    CurrentStepIndex int
    StepStates       map[string]interface{}
    Checkpoint       interface{}
    Version          int
    LastUpdated      time.Time
}

	state.State = stateJSON

	return &state, nil
}

// ClearExecutionState removes a workflow execution state
func (r *sqlExecutionRepository) ClearExecutionState(ctx context.Context, executionID string, tx db.Transaction) error {
	var dbTx db.Transaction
	var err error
	
	if tx != nil {
		dbTx = tx
	} else {
		dbTx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				dbTx.Rollback()
			}
		}()
	}

	query := `DELETE FROM workflow_states WHERE execution_id = ?`
	_, err = dbTx.Exec(ctx, query, executionID)
	if err != nil {
		return fmt.Errorf("failed to clear execution state: %w", err)
	}

	if tx == nil {
		if err = dbTx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// RecordFailedExecution records a failed workflow execution for retry
func (r *sqlExecutionRepository) RecordFailedExecution(ctx context.Context, executionID string, stepID string, reason string, metadata map[string]interface{}, tx db.Transaction) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal retry metadata: %w", err)
	}
	
	var dbTx db.Transaction
	if tx != nil {
		dbTx = tx
	} else {
		dbTx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				dbTx.Rollback()
			}
		}()
	}

	now := time.Now()
	query := `
		INSERT INTO workflow_retries (execution_id, step_id, reason, metadata, created_at, next_retry_at, attempt_count)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`
	
	// Default retry in 5 minutes
	nextRetry := now.Add(5 * time.Minute)
	
	_, err = dbTx.Exec(ctx, query, executionID, stepID, reason, metadataJSON, now, nextRetry)
	if err != nil {
		return fmt.Errorf("failed to record failed execution: %w", err)
	}

	if tx == nil {
		if err = dbTx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// GetFailedExecutions retrieves failed workflow executions for a tenant
func (r *sqlExecutionRepository) GetFailedExecutions(ctx context.Context, tenantID string) ([]*domain.WorkflowExecution, error) {
	query := `
		SELECT e.id, e.workflow_id, e.tenant_id, e.trigger_id, e.status, e.payload, e.context, e.steps, 
		       e.created_at, e.updated_at, e.started_at, e.completed_at, e.error_message
		FROM workflow_executions e
		JOIN workflow_retries r ON e.id = r.execution_id
		WHERE e.tenant_id = ? AND e.status = ?
		ORDER BY r.next_retry_at ASC
	`

	rows, err := r.db.Query(ctx, query, tenantID, domain.ExecutionStatusFailed)
	if err != nil {
		return nil, fmt.Errorf("failed to query failed executions: %w", err)
	}
	defer rows.Close()

	var executions []*domain.WorkflowExecution
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

// UpdateRetryAttempt updates the retry attempt count and next retry time for a workflow execution
func (r *sqlExecutionRepository) UpdateRetryAttempt(ctx context.Context, executionID string, nextRetry time.Time, tx db.Transaction) error {
	var dbTx db.Transaction
	var err error
	
	if tx != nil {
		dbTx = tx
	} else {
		dbTx, err = r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				dbTx.Rollback()
			}
		}()
	}

	query := `
		UPDATE workflow_retries 
		SET attempt_count = attempt_count + 1, next_retry_at = ?
		WHERE execution_id = ?
	`
	
	_, err = dbTx.Exec(ctx, query, nextRetry, executionID)
	if err != nil {
		return fmt.Errorf("failed to update retry attempt: %w", err)
	}

	if tx == nil {
		if err = dbTx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// CountExecutions counts workflow executions based on filters
func (r *sqlExecutionRepository) CountExecutions(ctx context.Context, tenantID string, filters repos.ExecutionFilters) (int64, error) {
	var args []interface{}
	var conditions []string
	
	// Base query
	query := `SELECT COUNT(*) FROM workflow_executions WHERE tenant_id = ?`
	args = append(args, tenantID)
	
	// Add filters
	if filters.WorkflowID != "" {
		conditions = append(conditions, "workflow_id = ?")
		args = append(args, filters.WorkflowID)
	}

	if filters.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filters.Status)
	}

	if filters.TriggerID != "" {
		conditions = append(conditions, "trigger_id = ?")
		args = append(args, filters.TriggerID)
	}

	// Add conditions to query
	for _, condition := range conditions {
		query += " AND " + condition
	}
	
	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count executions: %w", err)
	}
	
	return count, nil
}

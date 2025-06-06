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

type sqlWorkflowRepository struct {
	db db.Database
}

// NewWorkflowRepository creates a new SQL workflow repository
func NewWorkflowRepository(database db.Database) repos.WorkflowRepository {
	return &sqlWorkflowRepository{db: database}
}

// CreateWorkflow creates a new workflow in the database
func (r *sqlWorkflowRepository) CreateWorkflow(ctx context.Context, workflow *domain.Workflow) (*domain.Workflow, error) {
	stepsJSON, err := json.Marshal(workflow.Steps)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow steps: %w", err)
	}

	triggerJSON, err := json.Marshal(workflow.Trigger)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow trigger: %w", err)
	}

	query := `
		INSERT INTO workflows (id, tenant_id, name, description, status, trigger, steps, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(ctx, query,
		workflow.ID,
		workflow.TenantID,
		workflow.Name,
		workflow.Description,
		workflow.Status,
		triggerJSON,
		stepsJSON,
		workflow.CreatedAt,
		workflow.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return workflow, nil
}

// GetWorkflowByID retrieves a workflow by ID
func (r *sqlWorkflowRepository) GetWorkflowByID(ctx context.Context, workflowID string) (*domain.Workflow, error) {
	query := `
		SELECT id, tenant_id, name, description, status, trigger, steps, created_at, updated_at
		FROM workflows
		WHERE id = ?
	`

	row := r.db.QueryRow(ctx, query, workflowID)
	workflow, err := r.scanWorkflow(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow by ID: %w", err)
	}
	
	return workflow, nil
}

// GetByTriggerIdentifier retrieves a workflow by trigger identifier
func (r *sqlWorkflowRepository) GetByTriggerIdentifier(ctx context.Context, tenantID, triggerIdentifier string) (*domain.Workflow, error) {
	query := `
		SELECT id, tenant_id, name, description, status, trigger, steps, created_at, updated_at
		FROM workflows
		WHERE tenant_id = ? AND trigger->>'identifier' = ?
	`

	row := r.db.QueryRow(ctx, query, tenantID, triggerIdentifier)
	workflow, err := r.scanWorkflow(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow by trigger identifier: %w", err)
	}
	
	return workflow, nil
}

// UpdateWorkflow updates an existing workflow
func (r *sqlWorkflowRepository) UpdateWorkflow(ctx context.Context, workflow *domain.Workflow) (*domain.Workflow, error) {
	workflow.UpdatedAt = time.Now()

	stepsJSON, err := json.Marshal(workflow.Steps)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow steps: %w", err)
	}

	triggerJSON, err := json.Marshal(workflow.Trigger)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow trigger: %w", err)
	}

	query := `
		UPDATE workflows
		SET name = ?, description = ?, status = ?, trigger = ?, steps = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(ctx, query,
		workflow.Name,
		workflow.Description,
		workflow.Status,
		triggerJSON,
		stepsJSON,
		workflow.UpdatedAt,
		workflow.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("workflow not found or no changes made")
	}

	return workflow, nil
}

// DeleteWorkflow deletes a workflow
func (r *sqlWorkflowRepository) DeleteWorkflow(ctx context.Context, workflowID string) error {
	query := `DELETE FROM workflows WHERE id = ?`

	result, err := r.db.Exec(ctx, query, workflowID)
	if err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("workflow not found")
	}

	return nil
}

// GetWorkflowsByTenantID retrieves all workflows for a specific tenant
func (r *sqlWorkflowRepository) GetWorkflowsByTenantID(ctx context.Context, tenantID string) ([]*domain.Workflow, error) {
	query := `
		SELECT id, tenant_id, name, description, status, trigger, steps, created_at, updated_at
		FROM workflows
		WHERE tenant_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflows by tenant: %w", err)
	}
	defer rows.Close()

	var workflows []*domain.Workflow
	for rows.Next() {
		workflow, err := r.scanWorkflow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}
		workflows = append(workflows, workflow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workflow rows: %w", err)
	}

	return workflows, nil
}

// ListWorkflows retrieves workflows with pagination
func (r *sqlWorkflowRepository) ListWorkflows(ctx context.Context, limit, offset int) ([]*domain.Workflow, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM workflows`
	err := r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count workflows: %w", err)
	}

	// Get workflows with pagination
	query := `
		SELECT id, tenant_id, name, description, status, trigger, steps, created_at, updated_at
		FROM workflows
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*domain.Workflow
	for rows.Next() {
		workflow, err := r.scanWorkflow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan workflow: %w", err)
		}
		workflows = append(workflows, workflow)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating workflow rows: %w", err)
	}

	return workflows, total, nil
}

// Internal method kept for compatibility with existing service code
func (r *sqlWorkflowRepository) List(ctx context.Context, tenantID string, filters repos.WorkflowFilters) ([]*domain.Workflow, error) {
	var args []interface{}
	var conditions []string

	// Base query
	query := `
		SELECT id, tenant_id, name, description, status, trigger, steps, created_at, updated_at
		FROM workflows
		WHERE tenant_id = ?
	`
	args = append(args, tenantID)

	// Add filters
	if filters.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filters.Status)
	}

	if filters.Search != "" {
		conditions = append(conditions, "(name ILIKE ? OR description ILIKE ?)")
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Add conditions to query
	if len(conditions) > 0 {
		for i, condition := range conditions {
			if i == 0 {
				query += " AND " + condition
			} else {
				query += " AND " + condition
			}
		}
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
		return nil, fmt.Errorf("failed to query workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*domain.Workflow
	for rows.Next() {
		workflow, err := r.scanWorkflow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}
		workflows = append(workflows, workflow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workflow rows: %w", err)
	}

	return workflows, nil
}

// Count implements the WorkflowRepository.Count method
func (r *sqlWorkflowRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM workflows`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count workflows: %w", err)
	}

	return count, nil
}

// CountWithFilters counts workflows based on filters (internal helper)
func (r *sqlWorkflowRepository) CountWithFilters(ctx context.Context, tenantID string, filters repos.WorkflowFilters) (int64, error) {
	var args []interface{}
	var conditions []string

	// Base query
	query := `
		SELECT COUNT(*)
		FROM workflows
		WHERE tenant_id = ?
	`
	args = append(args, tenantID)

	// Add filters
	if filters.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filters.Status)
	}

	if filters.Search != "" {
		conditions = append(conditions, "(name ILIKE ? OR description ILIKE ?)")
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Add conditions to query
	if len(conditions) > 0 {
		for i, condition := range conditions {
			if i == 0 {
				query += " AND " + condition
			} else {
				query += " AND " + condition
			}
		}
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count workflows: %w", err)
	}

	return count, nil
}

// SetWorkflowActive sets the active status of a workflow
func (r *sqlWorkflowRepository) SetWorkflowActive(ctx context.Context, workflowID string, active bool) error {
	status := "inactive"
	if active {
		status = "active"
	}

	query := `UPDATE workflows SET status = ?, updated_at = ? WHERE id = ?`
	
	result, err := r.db.Exec(ctx, query, status, time.Now(), workflowID)
	if err != nil {
		return fmt.Errorf("failed to set workflow active status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("workflow not found")
	}

	return nil
}

// GetActiveWorkflows retrieves all active workflows
func (r *sqlWorkflowRepository) GetActiveWorkflows(ctx context.Context) ([]*domain.Workflow, error) {
	query := `
		SELECT id, tenant_id, name, description, status, trigger, steps, created_at, updated_at
		FROM workflows
		WHERE status = 'active'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*domain.Workflow
	for rows.Next() {
		workflow, err := r.scanWorkflow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}
		workflows = append(workflows, workflow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workflow rows: %w", err)
	}

	return workflows, nil
}

// GetWorkflowsByTriggerType retrieves workflows by trigger type
func (r *sqlWorkflowRepository) GetWorkflowsByTriggerType(ctx context.Context, triggerType string) ([]*domain.Workflow, error) {
	query := `
		SELECT id, tenant_id, name, description, status, trigger, steps, created_at, updated_at
		FROM workflows
		WHERE trigger->>'type' = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, triggerType)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflows by trigger type: %w", err)
	}
	defer rows.Close()

	var workflows []*domain.Workflow
	for rows.Next() {
		workflow, err := r.scanWorkflow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}
		workflows = append(workflows, workflow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workflow rows: %w", err)
	}

	return workflows, nil
}

// scanWorkflow scans a workflow from a database row
func (r *sqlWorkflowRepository) scanWorkflow(scanner interface{}) (*domain.Workflow, error) {
	var workflow domain.Workflow
	var stepsJSON, triggerJSON []byte

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&workflow.ID,
			&workflow.TenantID,
			&workflow.Name,
			&workflow.Description,
			&workflow.Status,
			&triggerJSON,
			&stepsJSON,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
		)
	case *sql.Rows:
		err = s.Scan(
			&workflow.ID,
			&workflow.TenantID,
			&workflow.Name,
			&workflow.Description,
			&workflow.Status,
			&triggerJSON,
			&stepsJSON,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan workflow: %w", err)
	}

	// Unmarshal JSON fields
	if len(stepsJSON) > 0 {
		if err := json.Unmarshal(stepsJSON, &workflow.Steps); err != nil {
			return nil, fmt.Errorf("failed to unmarshal workflow steps: %w", err)
		}
	}

	if len(triggerJSON) > 0 {
		if err := json.Unmarshal(triggerJSON, &workflow.Trigger); err != nil {
			return nil, fmt.Errorf("failed to unmarshal workflow trigger: %w", err)
		}
	}

	return &workflow, nil
}

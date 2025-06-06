package repos

import (
	"context"
	"time"

	"getnoti.com/internal/workflows/domain"
	"getnoti.com/pkg/db"
)



type ExecutionRepository interface {
	// Execution management
	CreateExecution(ctx context.Context, execution *domain.WorkflowExecution, tx db.Transaction) error
	GetExecutionByID(ctx context.Context, tenantID, executionID string) (*domain.WorkflowExecution, error)
	UpdateExecution(ctx context.Context, execution *domain.WorkflowExecution, tx db.Transaction) error
	ListExecutions(ctx context.Context, tenantID string, filters ExecutionFilters) ([]*domain.WorkflowExecution, error)
	CountExecutions(ctx context.Context, tenantID string, filters ExecutionFilters) (int64, error)
	GetPendingExecutions(ctx context.Context, limit int) ([]*domain.WorkflowExecution, error)
	
	// Step execution management
	CreateStepExecution(ctx context.Context, stepExecution *domain.StepExecution, tx db.Transaction) error
	UpdateStepExecution(ctx context.Context, stepExecution *domain.StepExecution, tx db.Transaction) error
	GetPendingStepExecutions(ctx context.Context, limit int) ([]*domain.StepExecution, error)
	GetDelayedStepExecutions(ctx context.Context, limit int) ([]*domain.StepExecution, error)
	GetStepExecutionsByExecutionID(ctx context.Context, executionID string) ([]*domain.StepExecution, error)
	
	// State Management
	SaveExecutionState(ctx context.Context, executionID string, state interface{}, checkpoint bool, tx db.Transaction) error
	GetExecutionState(ctx context.Context, executionID string) (*domain.WorkflowState, error)
	ClearExecutionState(ctx context.Context, executionID string, tx db.Transaction) error
	
	// Recovery Operations
	RecordFailedExecution(ctx context.Context, executionID string, stepID string, reason string, metadata map[string]interface{}, tx db.Transaction) error
	GetFailedExecutions(ctx context.Context, tenantID string) ([]*domain.WorkflowExecution, error)
	UpdateRetryAttempt(ctx context.Context, executionID string, nextRetry time.Time, tx db.Transaction) error
}

type WorkflowFilters struct {
	Status string
	Search string
	Limit  int
	Offset int
}

type ExecutionFilters struct {
	WorkflowID string
	Status     string
	TriggerID  string
	Limit      int
	Offset     int
}



// WorkflowRepository defines the interface for workflow persistence operations
type WorkflowRepository interface {
	// Workflow CRUD operations
	CreateWorkflow(ctx context.Context, workflow *domain.Workflow) (*domain.Workflow, error)
	GetWorkflowByID(ctx context.Context, workflowID string) (*domain.Workflow, error)
	GetWorkflowsByTenantID(ctx context.Context, tenantID string) ([]*domain.Workflow, error)
	ListWorkflows(ctx context.Context, limit, offset int) ([]*domain.Workflow, int64, error)
	UpdateWorkflow(ctx context.Context, workflow *domain.Workflow) (*domain.Workflow, error)
	DeleteWorkflow(ctx context.Context, workflowID string) error
	Count(ctx context.Context) (int64, error)
	
	// Additional query operations
	List(ctx context.Context, tenantID string, filters WorkflowFilters) ([]*domain.Workflow, error)
	CountWithFilters(ctx context.Context, tenantID string, filters WorkflowFilters) (int64, error)
	
	// Workflow status operations
	SetWorkflowActive(ctx context.Context, workflowID string, active bool) error
	GetActiveWorkflows(ctx context.Context) ([]*domain.Workflow, error)
	GetWorkflowsByTriggerType(ctx context.Context, triggerType string) ([]*domain.Workflow, error)
	GetByTriggerIdentifier(ctx context.Context, tenantID, triggerIdentifier string) (*domain.Workflow, error)
}

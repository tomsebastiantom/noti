package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusPaused    ExecutionStatus = "paused"
)

type WorkflowExecution struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	WorkflowID  uuid.UUID        `json:"workflow_id" db:"workflow_id"`
	TenantID    string           `json:"tenant_id" db:"tenant_id"`
	TriggerID   string           `json:"trigger_id" db:"trigger_id"`
	Status      ExecutionStatus  `json:"status" db:"status"`
	Payload     json.RawMessage  `json:"payload" db:"payload"`
	Context     ExecutionContext `json:"context" db:"context"`
	Steps       []StepExecution  `json:"steps" db:"steps"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
	StartedAt   *time.Time       `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty" db:"completed_at"`
	ErrorMessage string          `json:"error_message,omitempty" db:"error_message"`
}

type ExecutionContext struct {
	UserID      string                 `json:"user_id,omitempty"`
	Subscriber  map[string]interface{} `json:"subscriber,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type StepExecution struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	ExecutionID  uuid.UUID       `json:"execution_id" db:"execution_id"`
	StepID       string          `json:"step_id" db:"step_id"`
	StepType     StepType        `json:"step_type" db:"step_type"`
	Status       ExecutionStatus `json:"status" db:"status"`
	StartedAt    *time.Time      `json:"started_at,omitempty" db:"started_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	Result       json.RawMessage `json:"result,omitempty" db:"result"`
	ErrorMessage string          `json:"error_message,omitempty" db:"error_message"`
	RetryCount   int             `json:"retry_count" db:"retry_count"`
	DelayUntil   *time.Time      `json:"delay_until,omitempty" db:"delay_until"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// NewWorkflowExecution creates a new workflow execution
func NewWorkflowExecution(workflowID uuid.UUID, tenantID, triggerID string, payload json.RawMessage, context ExecutionContext) *WorkflowExecution {
	now := time.Now()
	return &WorkflowExecution{
		ID:         uuid.New(),
		WorkflowID: workflowID,
		TenantID:   tenantID,
		TriggerID:  triggerID,
		Status:     ExecutionStatusPending,
		Payload:    payload,
		Context:    context,
		Steps:      []StepExecution{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Start marks the execution as started
func (e *WorkflowExecution) Start() {
	now := time.Now()
	e.Status = ExecutionStatusRunning
	e.StartedAt = &now
	e.UpdatedAt = now
}

// Complete marks the execution as completed
func (e *WorkflowExecution) Complete() {
	now := time.Now()
	e.Status = ExecutionStatusCompleted
	e.CompletedAt = &now
	e.UpdatedAt = now
}

// Fail marks the execution as failed
func (e *WorkflowExecution) Fail(errorMsg string) {
	now := time.Now()
	e.Status = ExecutionStatusFailed
	e.ErrorMessage = errorMsg
	e.CompletedAt = &now
	e.UpdatedAt = now
}

// AddStepExecution adds a step execution
func (e *WorkflowExecution) AddStepExecution(stepID string, stepType StepType) *StepExecution {
	now := time.Now()
	stepExecution := StepExecution{
		ID:          uuid.New(),
		ExecutionID: e.ID,
		StepID:      stepID,
		StepType:    stepType,
		Status:      ExecutionStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	e.Steps = append(e.Steps, stepExecution)
	e.UpdatedAt = now
	
	return &stepExecution
}

// NewStepExecution creates a new step execution
func NewStepExecution(executionID uuid.UUID, stepID string, stepType StepType) *StepExecution {
	now := time.Now()
	return &StepExecution{
		ID:          uuid.New(),
		ExecutionID: executionID,
		StepID:      stepID,
		StepType:    stepType,
		Status:      ExecutionStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Start marks the step execution as started
func (s *StepExecution) Start() {
	now := time.Now()
	s.Status = ExecutionStatusRunning
	s.StartedAt = &now
	s.UpdatedAt = now
}

// Complete marks the step execution as completed
func (s *StepExecution) Complete(result json.RawMessage) {
	now := time.Now()
	s.Status = ExecutionStatusCompleted
	s.Result = result
	s.CompletedAt = &now
	s.UpdatedAt = now
}

// Fail marks the step execution as failed
func (s *StepExecution) Fail(errorMsg string) {
	now := time.Now()
	s.Status = ExecutionStatusFailed
	s.ErrorMessage = errorMsg
	s.CompletedAt = &now
	s.UpdatedAt = now
}

// SetDelayUntil sets a delay for the step execution
func (s *StepExecution) SetDelayUntil(delayTime time.Time) {
	s.DelayUntil = &delayTime
	s.UpdatedAt = time.Now()
}

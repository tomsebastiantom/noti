package dtos

import (
	"encoding/json"
	"time"

	"getnoti.com/internal/workflows/domain"
)

type CreateWorkflowRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Description string                 `json:"description" validate:"max=1000"`
	Trigger     WorkflowTriggerDTO     `json:"trigger" validate:"required"`
	Steps       []WorkflowStepDTO      `json:"steps" validate:"required,min=1"`
}

type UpdateWorkflowRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Description string                 `json:"description" validate:"max=1000"`
	Trigger     WorkflowTriggerDTO     `json:"trigger" validate:"required"`
	Steps       []WorkflowStepDTO      `json:"steps" validate:"required,min=1"`
}

type WorkflowResponse struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Trigger     WorkflowTriggerDTO     `json:"trigger"`
	Steps       []WorkflowStepDTO      `json:"steps"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type WorkflowTriggerDTO struct {
	Type       string                 `json:"type" validate:"required,oneof=event schedule webhook"`
	Identifier string                 `json:"identifier" validate:"required"`
	Config     map[string]interface{} `json:"config"`
}

type WorkflowStepDTO struct {
	ID         string                 `json:"id,omitempty"`
	Type       string                 `json:"type" validate:"required,oneof=email sms push webhook delay digest condition"`
	Name       string                 `json:"name" validate:"required"`
	Config     map[string]interface{} `json:"config" validate:"required"`
	Conditions []ConditionDTO         `json:"conditions,omitempty"`
	NextSteps  []string               `json:"next_steps,omitempty"`
	Position   int                    `json:"position"`
	Enabled    bool                   `json:"enabled"`
}

type ConditionDTO struct {
	Field    string      `json:"field" validate:"required"`
	Operator string      `json:"operator" validate:"required,oneof=eq ne gt lt gte lte contains in not_in"`
	Value    interface{} `json:"value" validate:"required"`
}

type TriggerWorkflowRequest struct {
	TriggerIdentifier string                 `json:"trigger_identifier" validate:"required"`
	Payload          map[string]interface{} `json:"payload" validate:"required"`
	Context          ExecutionContextDTO    `json:"context"`
}

type ExecutionContextDTO struct {
	UserID     string                 `json:"user_id,omitempty"`
	Subscriber map[string]interface{} `json:"subscriber,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type WorkflowExecutionResponse struct {
	ID          string                    `json:"id"`
	WorkflowID  string                    `json:"workflow_id"`
	TenantID    string                    `json:"tenant_id"`
	TriggerID   string                    `json:"trigger_id"`
	Status      string                    `json:"status"`
	Payload     map[string]interface{}    `json:"payload"`
	Context     ExecutionContextDTO       `json:"context"`
	Steps       []StepExecutionResponse   `json:"steps"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	StartedAt   *time.Time                `json:"started_at,omitempty"`
	CompletedAt *time.Time                `json:"completed_at,omitempty"`
	ErrorMessage string                   `json:"error_message,omitempty"`
}

type StepExecutionResponse struct {
	ID           string                 `json:"id"`
	ExecutionID  string                 `json:"execution_id"`
	StepID       string                 `json:"step_id"`
	StepType     string                 `json:"step_type"`
	Status       string                 `json:"status"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Result       map[string]interface{} `json:"result,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	RetryCount   int                    `json:"retry_count"`
	DelayUntil   *time.Time             `json:"delay_until,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type ListWorkflowsRequest struct {
	Status string `query:"status"`
	Search string `query:"search"`
	Limit  int    `query:"limit" validate:"min=1,max=100"`
	Offset int    `query:"offset" validate:"min=0"`
}

type ListWorkflowsResponse struct {
	Workflows []WorkflowResponse `json:"workflows"`
	Total     int64              `json:"total"`
	Limit     int                `json:"limit"`
	Offset    int                `json:"offset"`
}

type ListExecutionsRequest struct {
	WorkflowID string `query:"workflow_id"`
	Status     string `query:"status"`
	TriggerID  string `query:"trigger_id"`
	Limit      int    `query:"limit" validate:"min=1,max=100"`
	Offset     int    `query:"offset" validate:"min=0"`
}

type ListExecutionsResponse struct {
	Executions []WorkflowExecutionResponse `json:"executions"`
	Total      int64                       `json:"total"`
	Limit      int                         `json:"limit"`
	Offset     int                         `json:"offset"`
}

// Conversion methods
func ToWorkflowResponse(workflow *domain.Workflow) *WorkflowResponse {
	steps := make([]WorkflowStepDTO, len(workflow.Steps))
	for i, step := range workflow.Steps {
		conditions := make([]ConditionDTO, len(step.Conditions))
		for j, condition := range step.Conditions {
			conditions[j] = ConditionDTO{
				Field:    condition.Field,
				Operator: condition.Operator,
				Value:    condition.Value,
			}
		}
		
		steps[i] = WorkflowStepDTO{
			ID:         step.ID,
			Type:       string(step.Type),
			Name:       step.Name,
			Config:     step.Config,
			Conditions: conditions,
			NextSteps:  step.NextSteps,
			Position:   step.Position,
			Enabled:    step.Enabled,
		}
	}

	return &WorkflowResponse{
		ID:          workflow.ID.String(),
		TenantID:    workflow.TenantID,
		Name:        workflow.Name,
		Description: workflow.Description,
		Status:      string(workflow.Status),
		Trigger: WorkflowTriggerDTO{
			Type:       workflow.Trigger.Type,
			Identifier: workflow.Trigger.Identifier,
			Config:     workflow.Trigger.Config,
		},
		Steps:     steps,
		CreatedAt: workflow.CreatedAt,
		UpdatedAt: workflow.UpdatedAt,
	}
}

func ToExecutionResponse(execution *domain.WorkflowExecution) *WorkflowExecutionResponse {
	var payload map[string]interface{}
	if execution.Payload != nil {
		json.Unmarshal(execution.Payload, &payload)
	}

	steps := make([]StepExecutionResponse, len(execution.Steps))
	for i, step := range execution.Steps {
		var result map[string]interface{}
		if step.Result != nil {
			json.Unmarshal(step.Result, &result)
		}
		
		steps[i] = StepExecutionResponse{
			ID:           step.ID.String(),
			ExecutionID:  step.ExecutionID.String(),
			StepID:       step.StepID,
			StepType:     string(step.StepType),
			Status:       string(step.Status),
			StartedAt:    step.StartedAt,
			CompletedAt:  step.CompletedAt,
			Result:       result,
			ErrorMessage: step.ErrorMessage,
			RetryCount:   step.RetryCount,
			DelayUntil:   step.DelayUntil,
			CreatedAt:    step.CreatedAt,
			UpdatedAt:    step.UpdatedAt,
		}
	}

	return &WorkflowExecutionResponse{
		ID:         execution.ID.String(),
		WorkflowID: execution.WorkflowID.String(),
		TenantID:   execution.TenantID,
		TriggerID:  execution.TriggerID,
		Status:     string(execution.Status),
		Payload:    payload,
		Context: ExecutionContextDTO{
			UserID:     execution.Context.UserID,
			Subscriber: execution.Context.Subscriber,
			Variables:  execution.Context.Variables,
			Metadata:   execution.Context.Metadata,
		},
		Steps:        steps,
		CreatedAt:    execution.CreatedAt,
		UpdatedAt:    execution.UpdatedAt,
		StartedAt:    execution.StartedAt,
		CompletedAt:  execution.CompletedAt,
		ErrorMessage: execution.ErrorMessage,
	}
}

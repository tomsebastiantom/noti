package events

import (
	"getnoti.com/internal/shared/events"
)

// Workflow Event Types
const (
	WorkflowCreatedEventType         = "workflow.created"
	WorkflowUpdatedEventType         = "workflow.updated"
	WorkflowDeletedEventType         = "workflow.deleted"
	WorkflowActivatedEventType       = "workflow.activated"
	WorkflowDeactivatedEventType     = "workflow.deactivated"
	WorkflowExecutionStartedEventType = "workflow.execution.started"
	WorkflowExecutionCompletedEventType = "workflow.execution.completed"
	WorkflowExecutionFailedEventType  = "workflow.execution.failed"
	WorkflowStepExecutedEventType    = "workflow.step.executed"
	WorkflowStepFailedEventType      = "workflow.step.failed"
)

// Workflow Domain Events

// WorkflowCreatedEvent is published when a new workflow is created
type WorkflowCreatedEvent struct {
	*events.BaseDomainEvent
	WorkflowID   string                 `json:"workflow_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	TriggerType  string                 `json:"trigger_type"`
	StepCount    int                    `json:"step_count"`
	CreatedBy    string                 `json:"created_by"`
	WorkflowData map[string]interface{} `json:"workflow_data"`
}

// NewWorkflowCreatedEvent creates a new workflow created event
func NewWorkflowCreatedEvent(
	workflowID, tenantID, name, description, triggerType, createdBy string,
	stepCount int,
	workflowData map[string]interface{},
) *WorkflowCreatedEvent {
	payload := map[string]interface{}{
		"workflow_id":   workflowID,
		"name":          name,
		"description":   description,
		"trigger_type":  triggerType,
		"step_count":    stepCount,
		"created_by":    createdBy,
		"workflow_data": workflowData,
	}

	return &WorkflowCreatedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent(WorkflowCreatedEventType, workflowID, tenantID, payload),
		WorkflowID:      workflowID,
		Name:            name,
		Description:     description,
		TriggerType:     triggerType,
		StepCount:       stepCount,
		CreatedBy:       createdBy,
		WorkflowData:    workflowData,
	}
}

// WorkflowExecutionStartedEvent is published when a workflow execution begins
type WorkflowExecutionStartedEvent struct {
	*events.BaseDomainEvent
	ExecutionID   string                 `json:"execution_id"`
	WorkflowID    string                 `json:"workflow_id"`
	WorkflowName  string                 `json:"workflow_name"`
	TriggerID     string                 `json:"trigger_id"`
	TriggerType   string                 `json:"trigger_type"`
	StartedBy     string                 `json:"started_by"`
	ExecutionData map[string]interface{} `json:"execution_data"`
}

// NewWorkflowExecutionStartedEvent creates a new workflow execution started event
func NewWorkflowExecutionStartedEvent(
	executionID, workflowID, tenantID, workflowName, triggerID, triggerType, startedBy string,
	executionData map[string]interface{},
) *WorkflowExecutionStartedEvent {
	payload := map[string]interface{}{
		"execution_id":   executionID,
		"workflow_id":    workflowID,
		"workflow_name":  workflowName,
		"trigger_id":     triggerID,
		"trigger_type":   triggerType,
		"started_by":     startedBy,
		"execution_data": executionData,
	}

	return &WorkflowExecutionStartedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent(WorkflowExecutionStartedEventType, executionID, tenantID, payload),
		ExecutionID:     executionID,
		WorkflowID:      workflowID,
		WorkflowName:    workflowName,
		TriggerID:       triggerID,
		TriggerType:     triggerType,
		StartedBy:       startedBy,
		ExecutionData:   executionData,
	}
}

// WorkflowExecutionCompletedEvent is published when a workflow execution completes successfully
type WorkflowExecutionCompletedEvent struct {
	*events.BaseDomainEvent
	ExecutionID     string                 `json:"execution_id"`
	WorkflowID      string                 `json:"workflow_id"`
	WorkflowName    string                 `json:"workflow_name"`
	CompletedAt     string                 `json:"completed_at"`
	Duration        int64                  `json:"duration_ms"`
	StepsExecuted   int                    `json:"steps_executed"`
	ExecutionResult map[string]interface{} `json:"execution_result"`
}

// NewWorkflowExecutionCompletedEvent creates a new workflow execution completed event
func NewWorkflowExecutionCompletedEvent(
	executionID, workflowID, tenantID, workflowName, completedAt string,
	duration int64, stepsExecuted int,
	executionResult map[string]interface{},
) *WorkflowExecutionCompletedEvent {
	payload := map[string]interface{}{
		"execution_id":     executionID,
		"workflow_id":      workflowID,
		"workflow_name":    workflowName,
		"completed_at":     completedAt,
		"duration_ms":      duration,
		"steps_executed":   stepsExecuted,
		"execution_result": executionResult,
	}

	return &WorkflowExecutionCompletedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent(WorkflowExecutionCompletedEventType, executionID, tenantID, payload),
		ExecutionID:     executionID,
		WorkflowID:      workflowID,
		WorkflowName:    workflowName,
		CompletedAt:     completedAt,
		Duration:        duration,
		StepsExecuted:   stepsExecuted,
		ExecutionResult: executionResult,
	}
}

// WorkflowExecutionFailedEvent is published when a workflow execution fails
type WorkflowExecutionFailedEvent struct {
	*events.BaseDomainEvent
	ExecutionID   string                 `json:"execution_id"`
	WorkflowID    string                 `json:"workflow_id"`
	WorkflowName  string                 `json:"workflow_name"`
	FailedAt      string                 `json:"failed_at"`
	ErrorMessage  string                 `json:"error_message"`
	FailedStepID  string                 `json:"failed_step_id"`
	RetryCount    int                    `json:"retry_count"`
	WillRetry     bool                   `json:"will_retry"`
	ErrorDetails  map[string]interface{} `json:"error_details"`
}

// NewWorkflowExecutionFailedEvent creates a new workflow execution failed event
func NewWorkflowExecutionFailedEvent(
	executionID, workflowID, tenantID, workflowName, failedAt, errorMessage, failedStepID string,
	retryCount int, willRetry bool,
	errorDetails map[string]interface{},
) *WorkflowExecutionFailedEvent {
	payload := map[string]interface{}{
		"execution_id":   executionID,
		"workflow_id":    workflowID,
		"workflow_name":  workflowName,
		"failed_at":      failedAt,
		"error_message":  errorMessage,
		"failed_step_id": failedStepID,
		"retry_count":    retryCount,
		"will_retry":     willRetry,
		"error_details":  errorDetails,
	}

	return &WorkflowExecutionFailedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent(WorkflowExecutionFailedEventType, executionID, tenantID, payload),
		ExecutionID:     executionID,
		WorkflowID:      workflowID,
		WorkflowName:    workflowName,
		FailedAt:        failedAt,
		ErrorMessage:    errorMessage,
		FailedStepID:    failedStepID,
		RetryCount:      retryCount,
		WillRetry:       willRetry,
		ErrorDetails:    errorDetails,
	}
}

// WorkflowStepExecutedEvent is published when a workflow step completes (success or failure)
type WorkflowStepExecutedEvent struct {
	*events.BaseDomainEvent
	ExecutionID   string                 `json:"execution_id"`
	WorkflowID    string                 `json:"workflow_id"`
	StepID        string                 `json:"step_id"`
	StepType      string                 `json:"step_type"`
	StepName      string                 `json:"step_name"`
	Status        string                 `json:"status"` // completed, failed
	ExecutedAt    string                 `json:"executed_at"`
	Duration      int64                  `json:"duration_ms"`
	Result        map[string]interface{} `json:"result"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
}

// NewWorkflowStepExecutedEvent creates a new workflow step executed event
func NewWorkflowStepExecutedEvent(
	executionID, workflowID, tenantID, stepID, stepType, stepName, status, executedAt, errorMessage string,
	duration int64,
	result map[string]interface{},
) *WorkflowStepExecutedEvent {
	payload := map[string]interface{}{
		"execution_id":  executionID,
		"workflow_id":   workflowID,
		"step_id":       stepID,
		"step_type":     stepType,
		"step_name":     stepName,
		"status":        status,
		"executed_at":   executedAt,
		"duration_ms":   duration,
		"result":        result,
		"error_message": errorMessage,
	}

	return &WorkflowStepExecutedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent(WorkflowStepExecutedEventType, executionID, tenantID, payload),
		ExecutionID:     executionID,
		WorkflowID:      workflowID,
		StepID:          stepID,
		StepType:        stepType,
		StepName:        stepName,
		Status:          status,
		ExecutedAt:      executedAt,
		Duration:        duration,
		Result:          result,
		ErrorMessage:    errorMessage,
	}
}

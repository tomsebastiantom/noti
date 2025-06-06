package handlers

import (
	"context"
	"fmt"

	"getnoti.com/internal/shared/events"
	workflowEvents "getnoti.com/internal/workflows/events"
	"getnoti.com/pkg/logger"
)

// WorkflowEventHandlers handles workflow domain events
type WorkflowEventHandlers struct {
	logger     logger.Logger
}

// NewWorkflowEventHandlers creates a new workflow event handlers instance
func NewWorkflowEventHandlers(
	logger logger.Logger,
) *WorkflowEventHandlers {
	return &WorkflowEventHandlers{
		logger:     logger,
	}
}

// HandleWorkflowCreated processes workflow creation events
func (h *WorkflowEventHandlers) HandleWorkflowCreated(ctx context.Context, event events.DomainEvent) error {
	workflowEvent, ok := event.(*workflowEvents.WorkflowCreatedEvent)
	if !ok {
		h.logger.Error("Invalid event type for WorkflowCreated handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()},
			logger.Field{Key: "expected_type", Value: "WorkflowCreatedEvent"},
			logger.Field{Key: "actual_type", Value: fmt.Sprintf("%T", event)})
		return fmt.Errorf("invalid event type: expected WorkflowCreatedEvent, got %T", event)
	}

	h.logger.Info("Processing workflow created event",
		logger.Field{Key: "event_id", Value: workflowEvent.GetEventID()},
		logger.Field{Key: "workflow_id", Value: workflowEvent.WorkflowID},
		logger.Field{Key: "workflow_name", Value: workflowEvent.Name},
		logger.Field{Key: "trigger_type", Value: workflowEvent.TriggerType},
		logger.Field{Key: "tenant_id", Value: workflowEvent.GetTenantID()})

	// No notifications needed for workflow creation in this implementation
	return nil
}

// HandleWorkflowExecutionStarted processes workflow execution start events
func (h *WorkflowEventHandlers) HandleWorkflowExecutionStarted(ctx context.Context, event events.DomainEvent) error {
	executionEvent, ok := event.(*workflowEvents.WorkflowExecutionStartedEvent)
	if !ok {
		h.logger.Error("Invalid event type for WorkflowExecutionStarted handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()})
		return fmt.Errorf("invalid event type: expected WorkflowExecutionStartedEvent, got %T", event)
	}

	h.logger.Info("Processing workflow execution started event",
		logger.Field{Key: "event_id", Value: executionEvent.GetEventID()},
		logger.Field{Key: "execution_id", Value: executionEvent.ExecutionID},
		logger.Field{Key: "workflow_id", Value: executionEvent.WorkflowID},
		logger.Field{Key: "workflow_name", Value: executionEvent.WorkflowName},
		logger.Field{Key: "tenant_id", Value: executionEvent.GetTenantID()})

	// Notify admin users about workflow execution start if needed
	// This could be used for monitoring and alerting
	return nil
}

// HandleWorkflowExecutionCompleted processes workflow execution completion events
func (h *WorkflowEventHandlers) HandleWorkflowExecutionCompleted(ctx context.Context, event events.DomainEvent) error {
	completedEvent, ok := event.(*workflowEvents.WorkflowExecutionCompletedEvent)
	if !ok {
		h.logger.Error("Invalid event type for WorkflowExecutionCompleted handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()})
		return fmt.Errorf("invalid event type: expected WorkflowExecutionCompletedEvent, got %T", event)
	}

	h.logger.Info("Processing workflow execution completed event",
		logger.Field{Key: "event_id", Value: completedEvent.GetEventID()},
		logger.Field{Key: "execution_id", Value: completedEvent.ExecutionID},
		logger.Field{Key: "workflow_id", Value: completedEvent.WorkflowID},
		logger.Field{Key: "workflow_name", Value: completedEvent.WorkflowName},
		logger.Field{Key: "steps_executed", Value: completedEvent.StepsExecuted},
		logger.Field{Key: "tenant_id", Value: completedEvent.GetTenantID()})

	// Send completion notification to admins or stakeholders if needed
	return nil
}

// HandleWorkflowExecutionFailed processes workflow execution failure events
func (h *WorkflowEventHandlers) HandleWorkflowExecutionFailed(ctx context.Context, event events.DomainEvent) error {
	failedEvent, ok := event.(*workflowEvents.WorkflowExecutionFailedEvent)
	if !ok {
		h.logger.Error("Invalid event type for WorkflowExecutionFailed handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()})
		return fmt.Errorf("invalid event type: expected WorkflowExecutionFailedEvent, got %T", event)
	}

	h.logger.Error("Processing workflow execution failed event",
		logger.Field{Key: "event_id", Value: failedEvent.GetEventID()},
		logger.Field{Key: "execution_id", Value: failedEvent.ExecutionID},
		logger.Field{Key: "workflow_id", Value: failedEvent.WorkflowID},
		logger.Field{Key: "workflow_name", Value: failedEvent.WorkflowName},
		logger.Field{Key: "error_message", Value: failedEvent.ErrorMessage},
		logger.Field{Key: "tenant_id", Value: failedEvent.GetTenantID()})

	// We would send an admin notification for workflow failures here
	// But since we removed direct event bus access, we just log it
	h.logger.Info("Workflow failure would trigger notification",
		logger.Field{Key: "workflow_name", Value: failedEvent.WorkflowName},
		logger.Field{Key: "execution_id", Value: failedEvent.ExecutionID})

	return nil
}

// HandleWorkflowStepExecuted processes workflow step execution events
func (h *WorkflowEventHandlers) HandleWorkflowStepExecuted(ctx context.Context, event events.DomainEvent) error {
	stepEvent, ok := event.(*workflowEvents.WorkflowStepExecutedEvent)
	if !ok {
		h.logger.Error("Invalid event type for WorkflowStepExecuted handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()})
		return fmt.Errorf("invalid event type: expected WorkflowStepExecutedEvent, got %T", event)
	}

	h.logger.Info("Processing workflow step executed event",
		logger.Field{Key: "event_id", Value: stepEvent.GetEventID()},
		logger.Field{Key: "execution_id", Value: stepEvent.ExecutionID},
		logger.Field{Key: "workflow_id", Value: stepEvent.WorkflowID},
		logger.Field{Key: "step_id", Value: stepEvent.StepID},
		logger.Field{Key: "step_type", Value: stepEvent.StepType},
		logger.Field{Key: "status", Value: stepEvent.Status},
		logger.Field{Key: "tenant_id", Value: stepEvent.GetTenantID()})

	// For email/SMS/push steps, if they failed, log that we would send notification
	if stepEvent.Status == "failed" && (stepEvent.StepType == "email" || stepEvent.StepType == "sms" || stepEvent.StepType == "push") {
		h.logger.Info("Step failure would trigger notification",
			logger.Field{Key: "step_type", Value: stepEvent.StepType},
			logger.Field{Key: "step_id", Value: stepEvent.StepID},
			logger.Field{Key: "execution_id", Value: stepEvent.ExecutionID})
	}

	return nil
}

// GetHandlerMethods returns a map of event types to handler methods for registration
func (h *WorkflowEventHandlers) GetHandlerMethods() map[string]func(context.Context, events.DomainEvent) error {
	return map[string]func(context.Context, events.DomainEvent) error{
		workflowEvents.WorkflowCreatedEventType:           h.HandleWorkflowCreated,
		workflowEvents.WorkflowExecutionStartedEventType:  h.HandleWorkflowExecutionStarted,
		workflowEvents.WorkflowExecutionCompletedEventType: h.HandleWorkflowExecutionCompleted,
		workflowEvents.WorkflowExecutionFailedEventType:   h.HandleWorkflowExecutionFailed,
		workflowEvents.WorkflowStepExecutedEventType:      h.HandleWorkflowStepExecuted,
	}
}

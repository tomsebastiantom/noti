package services

import (
	"context"
	"fmt"

	notificationEvents "getnoti.com/internal/notifications/events"
	"getnoti.com/internal/shared/events"
	"getnoti.com/pkg/logger"
)

// WorkflowNotificationService handles sending notifications for workflow events
type WorkflowNotificationService struct {
	logger   logger.Logger
	eventBus events.EventBus
}

// NewWorkflowNotificationService creates a new workflow notification service
func NewWorkflowNotificationService(logger logger.Logger, eventBus events.EventBus) *WorkflowNotificationService {
	return &WorkflowNotificationService{
		logger:   logger,
		eventBus: eventBus,
	}
}

// SendWorkflowFailureNotification sends a notification for a workflow execution failure
func (s *WorkflowNotificationService) SendWorkflowFailureNotification(
	ctx context.Context,
	executionID, workflowName, tenantID, errorMessage, failedAt, failedStepID string,
) error {
	notificationEvent := notificationEvents.NewNotificationCreatedEvent(
		fmt.Sprintf("workflow-fail-%s", executionID), // Notification ID
		"admin",                                     // User ID (admin in this case)
		tenantID,                                    // Tenant ID
		"workflow_failure_template",                 // Template ID for the notification
		"email",                                     // Channel (email, sms, etc)
		"high",                                      // Priority
		nil,                                         // Scheduled for (immediate)
		map[string]interface{}{                      // Notification data
			"workflow_name":  workflowName,
			"execution_id":   executionID,
			"error_message":  errorMessage,
			"failure_time":   failedAt,
			"failed_step_id": failedStepID,
		},
	)

	// Publish the notification event
	if err := s.eventBus.PublishAsync(ctx, notificationEvent); err != nil {
		s.logger.Error("Failed to publish workflow failure notification",
			logger.Field{Key: "event_id", Value: notificationEvent.GetEventID()},
			logger.Field{Key: "error", Value: err.Error()})
		return fmt.Errorf("failed to publish workflow failure notification: %w", err)
	}

	s.logger.Info("Published workflow failure notification",
		logger.Field{Key: "notification_event_id", Value: notificationEvent.GetEventID()},
		logger.Field{Key: "workflow_name", Value: workflowName})

	return nil
}

// SendStepFailureNotification sends a notification for a workflow step execution failure
func (s *WorkflowNotificationService) SendStepFailureNotification(
	ctx context.Context,
	executionID, stepID, stepType, workflowID, tenantID, errorMessage, executedAt string,
) error {
	notificationEvent := notificationEvents.NewNotificationCreatedEvent(
		fmt.Sprintf("step-fail-%s-%s", executionID, stepID),
		"admin",                              // User ID (admin)
		tenantID,                             // Tenant ID
		"workflow_step_failure_template",     // Template ID
		"email",                              // Channel
		"medium",                             // Priority
		nil,                                  // Scheduled for (immediate)
		map[string]interface{}{               // Notification data
			"step_id":       stepID,
			"step_type":     stepType,
			"execution_id":  executionID,
			"workflow_id":   workflowID,
			"error_message": errorMessage,
			"failure_time":  executedAt,
		},
	)

	if err := s.eventBus.PublishAsync(ctx, notificationEvent); err != nil {
		s.logger.Error("Failed to publish step failure notification",
			logger.Field{Key: "event_id", Value: notificationEvent.GetEventID()},
			logger.Field{Key: "error", Value: err.Error()})
		return fmt.Errorf("failed to publish step failure notification: %w", err)
	}

	s.logger.Info("Published step failure notification",
		logger.Field{Key: "notification_event_id", Value: notificationEvent.GetEventID()},
		logger.Field{Key: "step_type", Value: stepType},
		logger.Field{Key: "step_id", Value: stepID})

	return nil
}

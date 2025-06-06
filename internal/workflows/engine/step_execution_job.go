package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	notificationServices "getnoti.com/internal/notifications/services"
	"getnoti.com/internal/shared/events"
	"getnoti.com/internal/workflows/domain"
	workflowEvents "getnoti.com/internal/workflows/events"
	"getnoti.com/internal/workflows/repos"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/workerpool"
)

// StepExecutionJob implements the workerpool.Job interface for individual step execution
type StepExecutionJob struct {
	stepExecution       *domain.StepExecution
	workflowStep        *domain.WorkflowStep
	execution           *domain.WorkflowExecution
	workflow            *domain.Workflow
	executionRepo       repos.ExecutionRepository
	eventBus            events.EventBus
	notificationService *notificationServices.NotificationService
	logger              logger.Logger
}

// NewStepExecutionJob creates a new step execution job
func NewStepExecutionJob(
	stepExecution *domain.StepExecution,
	workflowStep *domain.WorkflowStep,
	execution *domain.WorkflowExecution,
	workflow *domain.Workflow,
	executionRepo repos.ExecutionRepository,
	eventBus events.EventBus,
	notificationService *notificationServices.NotificationService,
	logger logger.Logger,
) workerpool.Job {
	return &StepExecutionJob{
		stepExecution:       stepExecution,
		workflowStep:        workflowStep,
		execution:           execution,
		workflow:            workflow,
		executionRepo:       executionRepo,
		eventBus:            eventBus,
		notificationService: notificationService,
		logger:              logger,
	}
}

// Process implements the workerpool.Job interface
func (j *StepExecutionJob) Process(ctx context.Context) error {
	j.logger.InfoContext(ctx, "Processing step execution",
		logger.String("step_execution_id", j.stepExecution.ID.String()),
		logger.String("step_id", j.workflowStep.ID),
		logger.String("step_type", string(j.workflowStep.Type)),
		logger.String("execution_id", j.execution.ID.String()))

	// Start the step execution
	j.stepExecution.Start()
		// Update step execution status to running
	if err := j.executionRepo.UpdateStepExecution(ctx, j.stepExecution, nil); err != nil {
		j.logger.Error("Failed to update step execution status to running",
			logger.String("step_execution_id", j.stepExecution.ID.String()),
			logger.Err(err))
	}

	var result map[string]interface{}
	var stepError error

	// Track execution time
	startTime := time.Now()

	// Execute based on step type
	switch j.workflowStep.Type {
	case domain.StepTypeEmail:
		result, stepError = j.executeEmailStep(ctx)
	case domain.StepTypeSMS:
		result, stepError = j.executeSMSStep(ctx)
	case domain.StepTypePush:
		result, stepError = j.executePushStep(ctx)
	case domain.StepTypeWebhook:
		result, stepError = j.executeWebhookStep(ctx)
	case domain.StepTypeDigest:
		result, stepError = j.executeDigestStep(ctx)
	case domain.StepTypeCondition:
		result, stepError = j.executeConditionStep(ctx)
	default:
		stepError = fmt.Errorf("unsupported step type: %s", j.workflowStep.Type)
	}

	// Calculate execution duration
	duration := time.Since(startTime).Milliseconds()

	// Complete or fail the step
	status := "completed"
	errorMessage := ""
	
	if stepError != nil {
		j.stepExecution.Fail(stepError.Error())
		j.logger.Error("Step execution failed",
			logger.String("step_id", j.workflowStep.ID),
			logger.String("step_execution_id", j.stepExecution.ID.String()),
			logger.Err(stepError))
		status = "failed"
		errorMessage = stepError.Error()
	} else {
		resultBytes, _ := json.Marshal(result)
		j.stepExecution.Complete(resultBytes)
		j.logger.Info("Step execution completed successfully",
			logger.String("step_id", j.workflowStep.ID),
			logger.String("step_execution_id", j.stepExecution.ID.String()))
	}
	// Save step execution final state
	if updateErr := j.executionRepo.UpdateStepExecution(ctx, j.stepExecution, nil); updateErr != nil {
		j.logger.Error("Failed to update step execution final status",
			logger.String("step_execution_id", j.stepExecution.ID.String()),
			logger.Err(updateErr))
	}

	// Publish step execution event
	j.publishStepExecutionEvent(ctx, status, duration, result, errorMessage)
	
	// Process next steps if this step completed successfully
	if stepError == nil {
		j.processNextSteps(ctx)
	}

	return stepError
}

// processNextSteps processes the next steps in the workflow
func (j *StepExecutionJob) processNextSteps(ctx context.Context) {
	if len(j.workflowStep.NextSteps) == 0 {
		j.logger.Debug("No next steps to process",
			logger.String("step_id", j.workflowStep.ID))
		return
	}
	j.logger.Info("Processing next steps",
		logger.String("step_id", j.workflowStep.ID),
		logger.String("next_steps", fmt.Sprintf("%v", j.workflowStep.NextSteps)))

	for _, nextStepID := range j.workflowStep.NextSteps {
		// Find the next step in the workflow
		var nextStep *domain.WorkflowStep
		for _, step := range j.workflow.Steps {
			if step.ID == nextStepID {
				nextStep = &step
				break
			}
		}

		if nextStep == nil {
			j.logger.Error("Next step not found in workflow",
				logger.String("next_step_id", nextStepID),
				logger.String("workflow_id", j.workflow.ID.String()))
			continue
		}

		// Check if this step is enabled and conditions are met
		if !nextStep.Enabled {
			j.logger.Debug("Skipping disabled next step",
				logger.String("next_step_id", nextStepID))
			continue
		}
		// Create step execution for next step
		nextStepExecution := j.execution.AddStepExecution(nextStep.ID, nextStep.Type)
		
		// Save step execution
		if err := j.executionRepo.CreateStepExecution(ctx, nextStepExecution, nil); err != nil {
			j.logger.Error("Failed to create next step execution",
				logger.String("next_step_id", nextStepID),
				logger.Err(err))
			continue
		}
		// Create and submit next step job
		nextStepJob := NewStepExecutionJob(nextStepExecution, nextStep, j.execution, j.workflow, j.executionRepo, j.eventBus, j.notificationService, j.logger)
				// For now, we'll log that the next step should be processed
		// In a real implementation, you might want to submit this to the worker pool
		// or handle it based on your specific workflow execution strategy
		j.logger.Info("Next step ready for processing",
			logger.String("next_step_id", nextStepID),
			logger.String("next_step_execution_id", nextStepExecution.ID.String()))
		
		// You could submit to worker pool here:
		// if err := workerPool.Submit(nextStepJob); err != nil { ... }
		
		// For now, let's process it directly (could cause deep recursion in complex workflows)
		if err := nextStepJob.Process(ctx); err != nil {
			j.logger.Error("Failed to process next step",
				logger.String("next_step_id", nextStepID),
				logger.Err(err))
		}
	}
}

// Step execution methods (same as in workflow_execution_job.go)
func (j *StepExecutionJob) executeEmailStep(ctx context.Context) (map[string]interface{}, error) {
	j.logger.Info("Executing email step", 
		logger.String("step_id", j.workflowStep.ID),
		logger.String("execution_id", j.execution.ID.String()))
	
	// Get recipient from context
	recipient := j.getRecipientFromContext("email")
	if recipient == "" {
		return nil, fmt.Errorf("no email recipient found in context")
	}
	
	// Get template ID from step config
	templateID := j.getTemplateFromConfig()
	if templateID == "" {
		return nil, fmt.Errorf("no template_id specified in step configuration")
	}
	
	// Get priority from config or use default
	priority := "normal"
	if p, ok := j.workflowStep.Config["priority"].(string); ok && p != "" {
		priority = p
	}
	
	// Extract subject and body from config or use defaults
	subject := fmt.Sprintf("Workflow %s notification", j.workflow.Name)
	if s, ok := j.workflowStep.Config["subject"].(string); ok && s != "" {
		subject = s
	}
	
	body := fmt.Sprintf("You have a notification from workflow %s", j.workflow.Name)
	if b, ok := j.workflowStep.Config["body"].(string); ok && b != "" {
		body = b
	}
	
	// Prepare variables for template
	variables := map[string]interface{}{
		"workflow_id":   j.workflow.ID.String(),
		"workflow_name": j.workflow.Name,
		"execution_id":  j.execution.ID.String(),
		"step_id":       j.workflowStep.ID,
		"step_name":     j.workflowStep.Name,
		"context":       j.execution.Context,
	}
	
	// Add custom data if it exists
	if cd, ok := j.workflowStep.Config["custom_data"]; ok {
		variables["custom_data"] = cd
	}
	
	// Create notification request using the notification service
	notificationReq := notificationServices.SendNotificationRequest{
		TenantID:   j.execution.TenantID,
		Channel:    "email",
		Recipients: []string{recipient},
		Subject:    subject,
		Body:       body,
		TemplateID: templateID,
		Variables:  variables,
		Priority:   priority,
	}
	
	// Send notification via the notification service
	response, err := j.notificationService.SendNotification(ctx, notificationReq)
	if err != nil {
		j.logger.Error("Failed to send email notification",
			logger.String("step_id", j.workflowStep.ID),
			logger.String("execution_id", j.execution.ID.String()),
			logger.String("recipient", recipient),
			logger.Err(err))
		return nil, fmt.Errorf("failed to send email notification: %w", err)
	}
	
	j.logger.Info("Email notification sent successfully",
		logger.String("step_id", j.workflowStep.ID),
		logger.String("execution_id", j.execution.ID.String()),
		logger.String("notification_id", response.NotificationID),
		logger.String("recipient", recipient),
		logger.String("template_id", templateID))
	
	return map[string]interface{}{
		"type":            "email",
		"sent":            true,
		"recipient":       recipient,
		"template_id":     templateID,
		"notification_id": response.NotificationID,
		"status":          response.Status,
	}, nil
}

func (j *StepExecutionJob) executeSMSStep(ctx context.Context) (map[string]interface{}, error) {
	j.logger.Info("Executing SMS step", 
		logger.String("step_id", j.workflowStep.ID),
		logger.String("execution_id", j.execution.ID.String()))
	
	// Get recipient from context
	recipient := j.getRecipientFromContext("phone")
	if recipient == "" {
		return nil, fmt.Errorf("no phone recipient found in context")
	}
	
	// Get template ID from step config
	templateID := j.getTemplateFromConfig()
	if templateID == "" {
		return nil, fmt.Errorf("no template_id specified in step configuration")
	}
	
	// Get priority from config or use default
	priority := "normal"
	if p, ok := j.workflowStep.Config["priority"].(string); ok && p != "" {
		priority = p
	}
	
	// Extract subject and body from config or use defaults
	subject := fmt.Sprintf("Workflow %s notification", j.workflow.Name)
	if s, ok := j.workflowStep.Config["subject"].(string); ok && s != "" {
		subject = s
	}
	
	body := fmt.Sprintf("You have a notification from workflow %s", j.workflow.Name)
	if b, ok := j.workflowStep.Config["body"].(string); ok && b != "" {
		body = b
	}
	
	// Prepare variables for template
	variables := map[string]interface{}{
		"workflow_id":   j.workflow.ID.String(),
		"workflow_name": j.workflow.Name,
		"execution_id":  j.execution.ID.String(),
		"step_id":       j.workflowStep.ID,
		"step_name":     j.workflowStep.Name,
		"context":       j.execution.Context,
	}
	
	// Add custom data if it exists
	if cd, ok := j.workflowStep.Config["custom_data"]; ok {
		variables["custom_data"] = cd
	}
	
	// Create notification request using the notification service
	notificationReq := notificationServices.SendNotificationRequest{
		TenantID:   j.execution.TenantID,
		Channel:    "sms",
		Recipients: []string{recipient},
		Subject:    subject,
		Body:       body,
		TemplateID: templateID,
		Variables:  variables,
		Priority:   priority,
	}
	
	// Send notification via the notification service
	response, err := j.notificationService.SendNotification(ctx, notificationReq)
	if err != nil {
		j.logger.Error("Failed to send SMS notification",
			logger.String("step_id", j.workflowStep.ID),
			logger.String("execution_id", j.execution.ID.String()),
			logger.String("recipient", recipient),
			logger.Err(err))
		return nil, fmt.Errorf("failed to send SMS notification: %w", err)
	}
	
	j.logger.Info("SMS notification sent successfully",
		logger.String("step_id", j.workflowStep.ID),
		logger.String("execution_id", j.execution.ID.String()),
		logger.String("notification_id", response.NotificationID),
		logger.String("recipient", recipient),
		logger.String("template_id", templateID))
		return map[string]interface{}{
		"type":            "sms",
		"sent":            true,
		"recipient":       recipient,
		"template_id":     templateID,
		"notification_id": response.NotificationID,
		"status":          response.Status,
	}, nil
}

func (j *StepExecutionJob) executePushStep(ctx context.Context) (map[string]interface{}, error) {
	j.logger.Info("Executing push step", 
		logger.String("step_id", j.workflowStep.ID),
		logger.String("execution_id", j.execution.ID.String()))
	
	// Get device_id from context
	deviceID := j.getRecipientFromContext("device_id")
	if deviceID == "" {
		return nil, fmt.Errorf("no device_id found in context")
	}
	
	// Get template ID from step config
	templateID := j.getTemplateFromConfig()
	if templateID == "" {
		return nil, fmt.Errorf("no template_id specified in step configuration")
	}
	
	// Get priority from config or use default
	priority := "normal"
	if p, ok := j.workflowStep.Config["priority"].(string); ok && p != "" {
		priority = p
	}
	
	// Extract subject and body from config or use defaults
	subject := fmt.Sprintf("Workflow %s notification", j.workflow.Name)
	if s, ok := j.workflowStep.Config["subject"].(string); ok && s != "" {
		subject = s
	}
	
	body := fmt.Sprintf("You have a notification from workflow %s", j.workflow.Name)
	if b, ok := j.workflowStep.Config["body"].(string); ok && b != "" {
		body = b
	}
	
	// Prepare variables for template
	variables := map[string]interface{}{
		"workflow_id":   j.workflow.ID.String(),
		"workflow_name": j.workflow.Name,
		"execution_id":  j.execution.ID.String(),
		"step_id":       j.workflowStep.ID,
		"step_name":     j.workflowStep.Name,
		"context":       j.execution.Context,
	}
	
	// Add custom data if it exists
	if cd, ok := j.workflowStep.Config["custom_data"]; ok {
		variables["custom_data"] = cd
	}
	
	// Create notification request using the notification service
	notificationReq := notificationServices.SendNotificationRequest{
		TenantID:   j.execution.TenantID,
		Channel:    "push",
		Recipients: []string{deviceID},
		Subject:    subject,
		Body:       body,
		TemplateID: templateID,
		Variables:  variables,
		Priority:   priority,
	}
	
	// Send notification via the notification service
	response, err := j.notificationService.SendNotification(ctx, notificationReq)
	if err != nil {
		j.logger.Error("Failed to send push notification",
			logger.String("step_id", j.workflowStep.ID),
			logger.String("execution_id", j.execution.ID.String()),
			logger.String("device_id", deviceID),
			logger.Err(err))
		return nil, fmt.Errorf("failed to send push notification: %w", err)
	}
	
	j.logger.Info("Push notification sent successfully",
		logger.String("step_id", j.workflowStep.ID),
		logger.String("execution_id", j.execution.ID.String()),
		logger.String("notification_id", response.NotificationID),
		logger.String("device_id", deviceID),
		logger.String("template_id", templateID))
	
	return map[string]interface{}{
		"type":            "push",
		"sent":            true,
		"device_id":       deviceID,
		"template_id":     templateID,
		"notification_id": response.NotificationID,
		"status":          response.Status,
	}, nil
}

func (j *StepExecutionJob) executeWebhookStep(ctx context.Context) (map[string]interface{}, error) {
	j.logger.Info("Executing webhook step", 
		logger.String("step_id", j.workflowStep.ID),
		logger.String("execution_id", j.execution.ID.String()))
	
	// Get webhook URL from step config
	webhookURL, _ := j.workflowStep.Config["url"].(string)
	
	return map[string]interface{}{
		"type": "webhook",
		"sent": true,
		"url": webhookURL,
	}, nil
}

func (j *StepExecutionJob) executeDigestStep(ctx context.Context) (map[string]interface{}, error) {
	j.logger.Info("Executing digest step", 
		logger.String("step_id", j.workflowStep.ID),
		logger.String("execution_id", j.execution.ID.String()))
	
	return map[string]interface{}{
		"type": "digest",
		"scheduled": true,
	}, nil
}

func (j *StepExecutionJob) executeConditionStep(ctx context.Context) (map[string]interface{}, error) {
	j.logger.Info("Executing condition step", 
		logger.String("step_id", j.workflowStep.ID),
		logger.String("execution_id", j.execution.ID.String()))
	
	return map[string]interface{}{
		"type": "condition",
		"evaluated": true,
	}, nil
}

// Helper methods
func (j *StepExecutionJob) getRecipientFromContext(field string) string {
	// Check if execution exists
	if j.execution == nil {
		j.logger.Warn("Execution is nil", 
			logger.String("step_id", j.workflowStep.ID))
		return ""
	}

	// Check if subscriber map exists or is empty
	if j.execution.Context.Subscriber == nil || len(j.execution.Context.Subscriber) == 0 {
		j.logger.Warn("Subscriber map is nil or empty in execution context", 
			logger.String("step_id", j.workflowStep.ID),
			logger.String("execution_id", j.execution.ID.String()))
		return ""
	}
	
	// Try to extract the specified field
	if subscriber, ok := j.execution.Context.Subscriber[field].(string); ok {
		if subscriber == "" {
			j.logger.Warn("Empty recipient value found in context", 
				logger.String("field", field),
				logger.String("step_id", j.workflowStep.ID),
				logger.String("execution_id", j.execution.ID.String()))
		}
		return subscriber
	}
	
	j.logger.Warn("Recipient field not found in context", 
		logger.String("field", field),
		logger.String("step_id", j.workflowStep.ID),
		logger.String("execution_id", j.execution.ID.String()))
	return ""
}

func (j *StepExecutionJob) getTemplateFromConfig() string {
	// Check if config exists
	if j.workflowStep.Config == nil {
		j.logger.Warn("Step config is nil", 
			logger.String("step_id", j.workflowStep.ID),
			logger.String("step_type", string(j.workflowStep.Type)))
		return ""
	}
	
	// Extract template ID from config
	templateID, ok := j.workflowStep.Config["template_id"].(string)
	if !ok {
		j.logger.Warn("Template ID not found in step config or not a string", 
			logger.String("step_id", j.workflowStep.ID),
			logger.String("step_type", string(j.workflowStep.Type)))
		return ""
	}
	
	if templateID == "" {
		j.logger.Warn("Empty template ID in step config", 
			logger.String("step_id", j.workflowStep.ID),
			logger.String("step_type", string(j.workflowStep.Type)))
	}
	
	return templateID
}

// publishStepExecutionEvent publishes an event for the step execution
func (j *StepExecutionJob) publishStepExecutionEvent(ctx context.Context, status string, duration int64, result map[string]interface{}, errorMessage string) {
	// Create domain event for step execution
	event := workflowEvents.NewWorkflowStepExecutedEvent(
		j.execution.ID.String(),
		j.workflow.ID.String(),
		j.execution.TenantID,
		j.workflowStep.ID,
		string(j.workflowStep.Type),
		j.workflowStep.Name,
		status,
		time.Now().Format(time.RFC3339),
		errorMessage,
		duration,
		result,
	)

	// Publish the event asynchronously to avoid blocking the workflow execution
	if err := j.eventBus.PublishAsync(ctx, event); err != nil {
		j.logger.Error("Failed to publish step execution event",
			logger.String("step_execution_id", j.stepExecution.ID.String()),
			logger.String("event_id", event.GetEventID()),
			logger.Err(err))
	} else {
		j.logger.Debug("Published step execution event",
			logger.String("step_execution_id", j.stepExecution.ID.String()),
			logger.String("event_id", event.GetEventID()),
			logger.String("status", status))
	}
}

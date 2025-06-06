package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"getnoti.com/internal/workflows/domain"
	"getnoti.com/internal/workflows/repos"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/workerpool"
)

// WorkflowExecutionJob implements the workerpool.Job interface
type WorkflowExecutionJob struct {
	execution     *domain.WorkflowExecution
	workflow      *domain.Workflow
	workflowRepo  repos.WorkflowRepository
	executionRepo repos.ExecutionRepository
	logger        logger.Logger
}

// NewWorkflowExecutionJob creates a new workflow execution job
func NewWorkflowExecutionJob(
	execution *domain.WorkflowExecution,
	workflow *domain.Workflow, 
	workflowRepo repos.WorkflowRepository,
	executionRepo repos.ExecutionRepository,
	logger logger.Logger,
) workerpool.Job {
	return &WorkflowExecutionJob{
		execution:     execution,
		workflow:      workflow,
		workflowRepo:  workflowRepo,
		executionRepo: executionRepo,
		logger:        logger,
	}
}

// Process implements the workerpool.Job interface
func (j *WorkflowExecutionJob) Process(ctx context.Context) error {
	j.logger.InfoContext(ctx, "Processing workflow execution",
		logger.String("execution_id", j.execution.ID.String()),
		logger.String("workflow_id", j.workflow.ID.String()),
		logger.String("tenant_id", j.execution.TenantID))

	// Start the execution
	j.execution.Start()
	
	// Update execution status to running
	if err := j.executionRepo.UpdateExecution(ctx, j.execution, nil); err != nil {
		j.logger.Error("Failed to update execution status to running",
			logger.String("execution_id", j.execution.ID.String()),
			logger.Err(err))
	}

	var executionError error

	// Process workflow steps
	if len(j.workflow.Steps) > 0 {
		executionError = j.processWorkflowSteps(ctx)
	} else {
		j.logger.Warn("Workflow has no steps",
			logger.String("workflow_id", j.workflow.ID.String()))
	}

	// Complete or fail the execution
	if executionError != nil {
		j.execution.Fail(executionError.Error())
		j.logger.Error("Workflow execution failed",
			logger.String("execution_id", j.execution.ID.String()),
			logger.String("workflow_id", j.workflow.ID.String()),
			logger.Err(executionError))
	} else {
		j.execution.Complete()
		j.logger.Info("Workflow execution completed successfully",
			logger.String("execution_id", j.execution.ID.String()),
			logger.String("workflow_id", j.workflow.ID.String()))
	}
	// Save final execution state
	if err := j.executionRepo.UpdateExecution(ctx, j.execution, nil); err != nil {
		j.logger.Error("Failed to update execution final status",
			logger.String("execution_id", j.execution.ID.String()),
			logger.Err(err))
	}

	return executionError
}

// processWorkflowSteps processes the workflow steps in order
func (j *WorkflowExecutionJob) processWorkflowSteps(ctx context.Context) error {
	// Sort steps by position
	steps := make([]domain.WorkflowStep, len(j.workflow.Steps))
	copy(steps, j.workflow.Steps)
	
	// Simple bubble sort by position (could use sort.Slice for better performance)
	for i := 0; i < len(steps)-1; i++ {
		for k := 0; k < len(steps)-i-1; k++ {
			if steps[k].Position > steps[k+1].Position {
				steps[k], steps[k+1] = steps[k+1], steps[k]
			}
		}
	}

	// Process each step
	for _, step := range steps {
		if !step.Enabled {
			j.logger.Debug("Skipping disabled step",
				logger.String("step_id", step.ID),
				logger.String("step_name", step.Name),
				logger.String("execution_id", j.execution.ID.String()))
			continue
		}

		// Check step conditions
		if !j.evaluateStepConditions(step) {
			j.logger.Debug("Step conditions not met, skipping",
				logger.String("step_id", step.ID),
				logger.String("step_name", step.Name),
				logger.String("execution_id", j.execution.ID.String()))
			continue
		}

		// Process the step
		if err := j.processStep(ctx, &step); err != nil {
			return fmt.Errorf("failed to process step %s: %w", step.ID, err)
		}
	}

	return nil
}

// processStep processes a single workflow step
func (j *WorkflowExecutionJob) processStep(ctx context.Context, step *domain.WorkflowStep) error {
	// Create step execution
	stepExecution := j.execution.AddStepExecution(step.ID, step.Type)
	
	// Save step execution
	err := j.executionRepo.CreateStepExecution(ctx, stepExecution, nil)
	if err != nil {
		return fmt.Errorf("failed to create step execution: %w", err)
	}

	j.logger.Info("Processing workflow step",
		logger.String("step_id", step.ID),
		logger.String("step_name", step.Name),
		logger.String("step_type", string(step.Type)),
		logger.String("execution_id", j.execution.ID.String()))

	// Handle delay steps immediately
	if step.Type == domain.StepTypeDelay {
		return j.processDelayStep(ctx, step, stepExecution)
	}

	// For other step types, process immediately
	return j.executeStep(ctx, step, stepExecution)
}

// processDelayStep handles delay steps
func (j *WorkflowExecutionJob) processDelayStep(ctx context.Context, step *domain.WorkflowStep, stepExecution *domain.StepExecution) error {
	// Get delay configuration from step config
	delayMinutes, ok := step.Config["delayMinutes"].(float64)
	if !ok {
		delayMinutes = 5 // Default to 5 minutes
	}

	// Calculate delay time
	delayTime := time.Now().Add(time.Duration(delayMinutes) * time.Minute)
	
	// Set delay on step execution
	stepExecution.SetDelayUntil(delayTime)
	
	j.logger.Info("Scheduling delayed step execution",
		logger.String("step_id", step.ID),
		logger.String("execution_id", j.execution.ID.String()),
		logger.Time("delay_until", delayTime),
		logger.Float64("delay_minutes", delayMinutes))

	// Update step execution with delay
	return j.executionRepo.UpdateStepExecution(ctx, stepExecution, nil)
}

// executeStep executes a step based on its type
func (j *WorkflowExecutionJob) executeStep(ctx context.Context, step *domain.WorkflowStep, stepExecution *domain.StepExecution) error {
	stepExecution.Start()
	
	// Update step execution to running
	if err := j.executionRepo.UpdateStepExecution(ctx, stepExecution, nil); err != nil {
		j.logger.Error("Failed to update step execution status to running",
			logger.String("step_execution_id", stepExecution.ID.String()),
			logger.Err(err))
	}

	var result map[string]interface{}
	var stepError error

	// Execute based on step type
	switch step.Type {
	case domain.StepTypeEmail:
		result, stepError = j.executeEmailStep(ctx, step)
	case domain.StepTypeSMS:
		result, stepError = j.executeSMSStep(ctx, step)
	case domain.StepTypePush:
		result, stepError = j.executePushStep(ctx, step)
	case domain.StepTypeWebhook:
		result, stepError = j.executeWebhookStep(ctx, step)
	case domain.StepTypeDigest:
		result, stepError = j.executeDigestStep(ctx, step)
	case domain.StepTypeCondition:
		result, stepError = j.executeConditionStep(ctx, step)
	default:
		stepError = fmt.Errorf("unsupported step type: %s", step.Type)
	}

	// Complete or fail the step
	if stepError != nil {
		stepExecution.Fail(stepError.Error())
		j.logger.Error("Step execution failed",
			logger.String("step_id", step.ID),
			logger.String("step_execution_id", stepExecution.ID.String()),
			logger.Err(stepError))
	} else {
		resultBytes, _ := json.Marshal(result)
		stepExecution.Complete(resultBytes)
		j.logger.Info("Step execution completed",
			logger.String("step_id", step.ID),
			logger.String("step_execution_id", stepExecution.ID.String()))
	}

	// Save step execution final state
	if err := j.executionRepo.UpdateStepExecution(ctx, stepExecution, nil); err != nil {
		j.logger.Error("Failed to update step execution final status",
			logger.String("step_execution_id", stepExecution.ID.String()),
			logger.Err(err))
	}

	return stepError
}

// evaluateStepConditions evaluates if step conditions are met
func (j *WorkflowExecutionJob) evaluateStepConditions(step domain.WorkflowStep) bool {
	if len(step.Conditions) == 0 {
		return true // No conditions means always execute
	}

	// For now, implement simple AND logic for all conditions
	for _, condition := range step.Conditions {
		if !j.evaluateCondition(condition) {
			return false
		}
	}
	
	return true
}

// evaluateCondition evaluates a single condition
func (j *WorkflowExecutionJob) evaluateCondition(condition domain.Condition) bool {
	// This is a simplified implementation
	// In a real system, you'd want more sophisticated condition evaluation
	// that can access execution context, previous step results, etc.
	
	j.logger.Debug("Evaluating condition",
		logger.String("field", condition.Field),
		logger.String("operator", condition.Operator),
		logger.Any("value", condition.Value))

	// For now, always return true (implement your condition logic here)
	return true
}

// Step execution methods (placeholder implementations)
func (j *WorkflowExecutionJob) executeEmailStep(ctx context.Context, step *domain.WorkflowStep) (map[string]interface{}, error) {
	j.logger.Info("Executing email step", logger.String("step_id", step.ID))
	// TODO: Integrate with your notification system
	return map[string]interface{}{
		"type": "email",
		"sent": true,
		"recipient": "user@example.com",
	}, nil
}

func (j *WorkflowExecutionJob) executeSMSStep(ctx context.Context, step *domain.WorkflowStep) (map[string]interface{}, error) {
	j.logger.Info("Executing SMS step", logger.String("step_id", step.ID))
	// TODO: Integrate with your SMS provider
	return map[string]interface{}{
		"type": "sms",
		"sent": true,
		"recipient": "+1234567890",
	}, nil
}

func (j *WorkflowExecutionJob) executePushStep(ctx context.Context, step *domain.WorkflowStep) (map[string]interface{}, error) {
	j.logger.Info("Executing push step", logger.String("step_id", step.ID))
	// TODO: Integrate with your push notification system
	return map[string]interface{}{
		"type": "push",
		"sent": true,
		"device_id": "device123",
	}, nil
}

func (j *WorkflowExecutionJob) executeWebhookStep(ctx context.Context, step *domain.WorkflowStep) (map[string]interface{}, error) {
	j.logger.Info("Executing webhook step", logger.String("step_id", step.ID))
	// TODO: Integrate with your webhook system
	return map[string]interface{}{
		"type": "webhook",
		"sent": true,
		"url": "https://example.com/webhook",
	}, nil
}

func (j *WorkflowExecutionJob) executeDigestStep(ctx context.Context, step *domain.WorkflowStep) (map[string]interface{}, error) {
	j.logger.Info("Executing digest step", logger.String("step_id", step.ID))
	// TODO: Implement digest functionality
	return map[string]interface{}{
		"type": "digest",
		"scheduled": true,
	}, nil
}

func (j *WorkflowExecutionJob) executeConditionStep(ctx context.Context, step *domain.WorkflowStep) (map[string]interface{}, error) {
	j.logger.Info("Executing condition step", logger.String("step_id", step.ID))
	// TODO: Implement conditional branching
	return map[string]interface{}{
		"type": "condition",
		"evaluated": true,
	}, nil
}

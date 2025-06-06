package engine

import (
	"context"
	"fmt"
	"time"

	"getnoti.com/internal/workflows/domain"
	"getnoti.com/internal/workflows/repos"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/workerpool"
)

// WorkflowRetryJob implements the workerpool.Job interface for retrying failed workflow executions
type WorkflowRetryJob struct {
	execution    *domain.WorkflowExecution
	workflow     *domain.Workflow
	workflowRepo repos.WorkflowRepository
	executionRepo repos.ExecutionRepository
	logger       logger.Logger
}

// NewWorkflowRetryJob creates a new workflow retry job
func NewWorkflowRetryJob(
	execution *domain.WorkflowExecution,
	workflow *domain.Workflow,
	workflowRepo repos.WorkflowRepository,
	executionRepo repos.ExecutionRepository,
	logger logger.Logger,
) workerpool.Job {
	return &WorkflowRetryJob{
		execution:    execution,
		workflow:     workflow,
		workflowRepo: workflowRepo,
		executionRepo: executionRepo,
		logger:       logger,
	}
}

// Process implements the workerpool.Job interface
func (j *WorkflowRetryJob) Process(ctx context.Context) error {
	j.logger.InfoContext(ctx, "Processing workflow retry job",
		logger.String("execution_id", j.execution.ID.String()),
		logger.String("workflow_id", j.workflow.ID.String()),
		logger.String("tenant_id", j.execution.TenantID))
	// Reset execution status for retry
	j.execution.Status = domain.ExecutionStatusRunning
	
	// Update execution in repository
	if err := j.executionRepo.UpdateExecution(ctx, j.execution, nil); err != nil {
		j.logger.Error("Failed to update execution status for retry",
			logger.String("execution_id", j.execution.ID.String()),
			logger.Err(err))
		return err
	}

	// Create a new execution job to handle the retry
	job := NewWorkflowExecutionJob(j.execution, j.workflow, j.workflowRepo, j.executionRepo, j.logger)
	
	// Process the execution directly
	return job.Process(ctx)
}

// DelayedStepJob implements the workerpool.Job interface for processing delayed workflow steps
type DelayedStepJob struct {
	stepExecution *domain.StepExecution
	execution     *domain.WorkflowExecution
	workflow      *domain.Workflow
	workflowRepo  repos.WorkflowRepository
	executionRepo repos.ExecutionRepository
	logger        logger.Logger
}

// NewDelayedStepJob creates a new delayed step execution job
func NewDelayedStepJob(
	stepExecution *domain.StepExecution,
	execution *domain.WorkflowExecution,
	workflow *domain.Workflow,
	workflowRepo repos.WorkflowRepository,
	executionRepo repos.ExecutionRepository,
	logger logger.Logger,
) workerpool.Job {
	return &DelayedStepJob{
		stepExecution: stepExecution,
		execution:     execution,
		workflow:      workflow,
		workflowRepo:  workflowRepo,
		executionRepo: executionRepo,
		logger:        logger,
	}
}

// Process implements the workerpool.Job interface
func (j *DelayedStepJob) Process(ctx context.Context) error {
	j.logger.InfoContext(ctx, "Processing delayed step job",
		logger.String("step_execution_id", j.stepExecution.ID.String()),
		logger.String("execution_id", j.execution.ID.String()),
		logger.String("workflow_id", j.workflow.ID.String()),
		logger.String("tenant_id", j.execution.TenantID))
	// Update step status to running
	j.stepExecution.Status = domain.ExecutionStatusRunning
	now := time.Now()
	j.stepExecution.StartedAt = &now
	
	// Save step execution status
	if err := j.executionRepo.UpdateStepExecution(ctx, j.stepExecution, nil); err != nil {
		j.logger.Error("Failed to update step execution status",
			logger.String("step_execution_id", j.stepExecution.ID.String()),
			logger.Err(err))
		return err
	}

	// Find the step in the workflow
	var step *domain.WorkflowStep
	for _, s := range j.workflow.Steps {
		if s.ID == j.stepExecution.StepID {
			step = &s
			break
		}
	}

	if step == nil {
		err := fmt.Errorf("step not found in workflow: %s", j.stepExecution.StepID)
		j.logger.Error("Step not found in workflow",
			logger.String("step_id", j.stepExecution.StepID),
			logger.String("workflow_id", j.workflow.ID.String()),
			logger.Err(err))
				j.stepExecution.Status = domain.ExecutionStatusFailed
		j.stepExecution.ErrorMessage = err.Error()
		j.executionRepo.UpdateStepExecution(ctx, j.stepExecution, nil)
		return err
	}
	// Execute the step
	var err error
	output, err := executeStep(ctx, step, j.execution, nil, j.logger) // Using nil for input since field doesn't exist
	
	if err != nil {
		j.stepExecution.Status = domain.ExecutionStatusFailed
		j.stepExecution.ErrorMessage = err.Error()
		j.logger.Error("Step execution failed",
			logger.String("step_id", step.ID),
			logger.String("step_execution_id", j.stepExecution.ID.String()),
			logger.Err(err))
	} else {
		j.stepExecution.Status = domain.ExecutionStatusCompleted
		j.stepExecution.Result = output
		j.logger.Info("Step executed successfully",
			logger.String("step_id", step.ID),
			logger.String("step_execution_id", j.stepExecution.ID.String()))
	}
	
	now2 := time.Now()
	j.stepExecution.CompletedAt = &now2
	
	// Save step execution result
	if updateErr := j.executionRepo.UpdateStepExecution(ctx, j.stepExecution, nil); updateErr != nil {
		j.logger.Error("Failed to update step execution result",
			logger.String("step_execution_id", j.stepExecution.ID.String()),
			logger.Err(updateErr))
		// Return the original error if there was one
		if err != nil {
			return err
		}
		return updateErr
	}

	return err
}

// WorkflowExecutionJob implements the workerpool.Job interface for executing workflows
type WorkflowExecutionJob struct {
	execution    *domain.WorkflowExecution
	workflow     *domain.Workflow
	workflowRepo repos.WorkflowRepository
	executionRepo repos.ExecutionRepository
	logger       logger.Logger
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
		execution:    execution,
		workflow:     workflow,
		workflowRepo: workflowRepo,
		executionRepo: executionRepo,
		logger:       logger,
	}
}

// Process implements the workerpool.Job interface
func (j *WorkflowExecutionJob) Process(ctx context.Context) error {
	j.logger.InfoContext(ctx, "Processing workflow execution job",
		logger.String("execution_id", j.execution.ID.String()),
		logger.String("workflow_id", j.workflow.ID.String()),
		logger.String("tenant_id", j.execution.TenantID))

	// Start the execution
	j.execution.Start()
	
	// Update execution in repository
	if err := j.executionRepo.UpdateExecution(ctx, j.execution, nil); err != nil {
		j.logger.Error("Failed to update execution status to running",
			logger.String("execution_id", j.execution.ID.String()),
			logger.Err(err))
		return err
	}

	// Execute workflow steps
	for _, step := range j.workflow.Steps {
		// Create step execution
		stepExecution := j.execution.AddStepExecution(step.ID, step.Type)
		
		// Save step execution
		if err := j.executionRepo.CreateStepExecution(ctx, stepExecution, nil); err != nil {
			j.logger.Error("Failed to create step execution",
				logger.String("step_id", step.ID),
				logger.Err(err))
			continue
		}

		// Execute the step
		stepExecution.Start()
		output, err := executeStep(ctx, &step, j.execution, nil, j.logger)
		
		if err != nil {
			stepExecution.Fail(err.Error())
			j.logger.Error("Step execution failed",
				logger.String("step_id", step.ID),
				logger.String("step_execution_id", stepExecution.ID.String()),
				logger.Err(err))
		} else {
			stepExecution.Complete(output)
			j.logger.Info("Step executed successfully",
				logger.String("step_id", step.ID),
				logger.String("step_execution_id", stepExecution.ID.String()))
		}
		
		// Update step execution
		if updateErr := j.executionRepo.UpdateStepExecution(ctx, stepExecution, nil); updateErr != nil {
			j.logger.Error("Failed to update step execution",
				logger.String("step_execution_id", stepExecution.ID.String()),
				logger.Err(updateErr))
		}
		
		// If step failed, fail the entire execution
		if err != nil {
			j.execution.Fail(fmt.Sprintf("Step %s failed: %v", step.ID, err))
			j.executionRepo.UpdateExecution(ctx, j.execution, nil)
			return err
		}
	}

	// Complete the execution
	j.execution.Complete()
	
	// Update execution in repository
	if err := j.executionRepo.UpdateExecution(ctx, j.execution, nil); err != nil {
		j.logger.Error("Failed to update execution status to completed",
			logger.String("execution_id", j.execution.ID.String()),
			logger.Err(err))
		return err
	}

	j.logger.InfoContext(ctx, "Workflow execution completed successfully",
		logger.String("execution_id", j.execution.ID.String()),
		logger.String("workflow_id", j.workflow.ID.String()))

	return nil
}

// executeStep executes a workflow step based on its type and configuration
// This is a placeholder implementation - in a real application, this would be more complex
func executeStep(ctx context.Context, step *domain.WorkflowStep, execution *domain.WorkflowExecution, input []byte, log logger.Logger) ([]byte, error) {
	// Implement step execution logic based on the step type
	// This is a simplified implementation for demonstration purposes
	switch step.Type {
	case "notification":
		// Here we would trigger a notification using the notification service
		return []byte(`{"success": true, "message": "Notification sent"}`), nil
	case "webhook":
		// Here we would make an HTTP request to a webhook endpoint
		return []byte(`{"success": true, "message": "Webhook called"}`), nil
	case "condition":
		// Here we would evaluate a condition and return a result
		return []byte(`{"result": true}`), nil
	default:
		return nil, fmt.Errorf("unsupported step type: %s", step.Type)
	}
}

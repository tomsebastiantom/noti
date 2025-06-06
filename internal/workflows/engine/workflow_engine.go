package engine

import (
	"context"
	"fmt"
	"time"

	notificationServices "getnoti.com/internal/notifications/services"
	"getnoti.com/internal/shared/events"
	"getnoti.com/internal/workflows/domain"
	"getnoti.com/internal/workflows/repos"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/workerpool"
)

// WorkflowEngine manages workflow execution using the worker pool
type WorkflowEngine struct {
	workflowRepo        repos.WorkflowRepository
	executionRepo       repos.ExecutionRepository
	workerPool          *workerpool.WorkerPool
	logger              logger.Logger
	eventBus            events.EventBus
	notificationService *notificationServices.NotificationService
	stopCh              chan struct{}
	pollInterval        time.Duration
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(
	workflowRepo repos.WorkflowRepository,
	executionRepo repos.ExecutionRepository,
	workerPool *workerpool.WorkerPool,
	logger logger.Logger,
	eventBus events.EventBus,
	notificationService *notificationServices.NotificationService,
	pollInterval time.Duration,
) *WorkflowEngine {
	return &WorkflowEngine{
		workflowRepo:        workflowRepo,
		executionRepo:       executionRepo,
		workerPool:          workerPool,
		logger:              logger,
		eventBus:            eventBus,
		notificationService: notificationService,
		pollInterval:        pollInterval,
		stopCh:              make(chan struct{}),
	}
}

// Start begins the workflow engine polling loop
func (e *WorkflowEngine) Start(ctx context.Context) error {
	e.logger.Info("Starting workflow engine")
	
	go e.pollForPendingExecutions(ctx)
	go e.pollForDelayedSteps(ctx)
	
	return nil
}

// Stop stops the workflow engine
func (e *WorkflowEngine) Stop() {
	e.logger.Info("Stopping workflow engine")
	close(e.stopCh)
}

// TriggerWorkflow creates and starts a new workflow execution
func (e *WorkflowEngine) TriggerWorkflow(ctx context.Context, workflowID string, triggerID string, payload []byte, execCtx domain.ExecutionContext) (*domain.WorkflowExecution, error) {
	// Get the workflow
	workflow, err := e.workflowRepo.GetWorkflowByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	if workflow.Status != domain.WorkflowStatusActive {
		return nil, fmt.Errorf("workflow is not active: %s", workflow.Status)
	}

	// Create new execution
	execution := domain.NewWorkflowExecution(workflow.ID, workflow.TenantID, triggerID, payload, execCtx)
		// Save execution
	err = e.executionRepo.CreateExecution(ctx, execution, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}
	
	savedExecution := execution

	// Submit execution job to worker pool
	job := NewWorkflowExecutionJob(savedExecution, workflow, e.workflowRepo, e.executionRepo, e.logger)
	if err := e.workerPool.Submit(job); err != nil {
		e.logger.Error("Failed to submit workflow execution job",
			logger.String("execution_id", savedExecution.ID.String()),
			logger.String("workflow_id", workflowID),
			logger.Err(err))
				// Mark execution as failed
		savedExecution.Fail("Failed to submit execution job: " + err.Error())
		e.executionRepo.UpdateExecution(ctx, savedExecution, nil)
		return nil, fmt.Errorf("failed to submit execution job: %w", err)
	}

	e.logger.Info("Workflow execution triggered",
		logger.String("execution_id", savedExecution.ID.String()),
		logger.String("workflow_id", workflowID),
		logger.String("trigger_id", triggerID))

	return savedExecution, nil
}

// pollForPendingExecutions checks for pending executions and processes them
func (e *WorkflowEngine) pollForPendingExecutions(ctx context.Context) {
	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	e.logger.Info("Started polling for pending workflow executions")
	
	for {
		select {
		case <-ticker.C:
			if err := e.processPendingExecutions(ctx); err != nil {
				e.logger.Error("Error processing pending executions", logger.Err(err))
			}
		case <-e.stopCh:
			e.logger.Info("Stopped polling for pending workflow executions")
			return
		case <-ctx.Done():
			e.logger.Info("Context done, stopped polling for pending workflow executions")
			return
		}
	}
}

// pollForDelayedSteps checks for delayed steps that are ready to execute
func (e *WorkflowEngine) pollForDelayedSteps(ctx context.Context) {
	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	e.logger.Info("Started polling for delayed workflow steps")
	
	for {
		select {
		case <-ticker.C:
			if err := e.processDelayedSteps(ctx); err != nil {
				e.logger.Error("Error processing delayed steps", logger.Err(err))
			}
		case <-e.stopCh:
			e.logger.Info("Stopped polling for delayed workflow steps")
			return
		case <-ctx.Done():
			e.logger.Info("Context done, stopped polling for delayed workflow steps")
			return
		}
	}
}

// processPendingExecutions finds and processes pending executions
func (e *WorkflowEngine) processPendingExecutions(ctx context.Context) error {
	executions, err := e.executionRepo.GetPendingExecutions(ctx, 50) // Process up to 50 at a time
	if err != nil {
		return fmt.Errorf("failed to get pending executions: %w", err)
	}

	if len(executions) == 0 {
		return nil
	}

	e.logger.Debug("Processing pending executions", logger.Int("count", len(executions)))

	for _, execution := range executions {
		// Get the workflow for this execution
		workflow, err := e.workflowRepo.GetWorkflowByID(ctx, execution.WorkflowID.String())
		if err != nil {
			e.logger.Error("Failed to get workflow for execution",
				logger.String("execution_id", execution.ID.String()),
				logger.String("workflow_id", execution.WorkflowID.String()),
				logger.Err(err))
			continue
		}

		// Submit execution job to worker pool
		job := NewWorkflowExecutionJob(execution, workflow, e.workflowRepo, e.executionRepo, e.logger)
		if err := e.workerPool.Submit(job); err != nil {
			e.logger.Error("Failed to submit pending execution job",
				logger.String("execution_id", execution.ID.String()),
				logger.Err(err))
		}
	}
	
	return nil
}

// processDelayedSteps finds and processes delayed steps that are ready
func (e *WorkflowEngine) processDelayedSteps(ctx context.Context) error {
	delayedSteps, err := e.executionRepo.GetDelayedStepExecutions(ctx, 50) // Process up to 50 at a time
	if err != nil {
		return fmt.Errorf("failed to get ready delayed steps: %w", err)
	}

	if len(delayedSteps) == 0 {
		return nil
	}

	e.logger.Debug("Processing ready delayed steps", logger.Int("count", len(delayedSteps)))
	for _, stepExecution := range delayedSteps {
		// Get the execution and workflow
		executionID := stepExecution.ExecutionID.String()
		
		// First get the execution to get the tenant ID
		executions, err := e.executionRepo.ListExecutions(ctx, "", repos.ExecutionFilters{
			Limit: 1,
		})
		if err != nil || len(executions) == 0 {
			e.logger.Error("Failed to find execution for delayed step",
				logger.String("step_execution_id", stepExecution.ID.String()),
				logger.String("execution_id", executionID),
				logger.Err(err))
			continue
		}
		
		tenantID := executions[0].TenantID
		execution, err := e.executionRepo.GetExecutionByID(ctx, tenantID, executionID)
		if err != nil {
			e.logger.Error("Failed to get execution for delayed step",
				logger.String("step_execution_id", stepExecution.ID.String()),
				logger.String("execution_id", executionID),
				logger.Err(err))
			continue
		}

		workflow, err := e.workflowRepo.GetWorkflowByID(ctx, execution.WorkflowID.String())
		if err != nil {
			e.logger.Error("Failed to get workflow for delayed step",
				logger.String("execution_id", execution.ID.String()),
				logger.String("workflow_id", execution.WorkflowID.String()),
				logger.Err(err))
			continue
		}

		// Find the workflow step
		var workflowStep *domain.WorkflowStep
		for _, step := range workflow.Steps {
			if step.ID == stepExecution.StepID {
				workflowStep = &step
				break
			}
		}

		if workflowStep == nil {
			e.logger.Error("Workflow step not found for delayed step",
				logger.String("step_id", stepExecution.StepID),
				logger.String("workflow_id", workflow.ID.String()))
			continue
		}		// Submit step execution job to worker pool
		job := NewStepExecutionJob(stepExecution, workflowStep, execution, workflow, e.executionRepo, e.eventBus, e.notificationService, e.logger)
		if err := e.workerPool.Submit(job); err != nil {
			e.logger.Error("Failed to submit delayed step execution job",
				logger.String("step_execution_id", stepExecution.ID.String()),
				logger.Err(err))
		}
	}
	return nil
}


// SetPollInterval sets the polling interval for the engine
func (e *WorkflowEngine) SetPollInterval(interval time.Duration) {
	e.pollInterval = interval
}

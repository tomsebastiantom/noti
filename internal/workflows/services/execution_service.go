package services

import (
	"context"
	"fmt"

	"getnoti.com/internal/workflows/domain"
	"getnoti.com/internal/workflows/dtos"
	"getnoti.com/internal/workflows/repos"
	"getnoti.com/pkg/logger"
)

type ExecutionService struct {
	executionRepo repos.ExecutionRepository
	logger        logger.Logger
}

func NewExecutionService(
	executionRepo repos.ExecutionRepository,
	logger logger.Logger,
) *ExecutionService {
	return &ExecutionService{
		executionRepo: executionRepo,
		logger:        logger,
	}
}

func (s *ExecutionService) GetExecution(ctx context.Context, tenantID, executionID string) (*dtos.WorkflowExecutionResponse, error) {
	execution, err := s.executionRepo.GetExecutionByID(ctx, tenantID, executionID)
	if err != nil {
		return nil, err
	}

	return dtos.ToExecutionResponse(execution), nil
}

func (s *ExecutionService) ListExecutions(ctx context.Context, tenantID string, req *dtos.ListExecutionsRequest) (*dtos.ListExecutionsResponse, error) {
	filters := repos.ExecutionFilters{
		WorkflowID: req.WorkflowID,
		Status:     req.Status,
		TriggerID:  req.TriggerID,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	executions, err := s.executionRepo.ListExecutions(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list executions: %w", err)
	}

	total, err := s.executionRepo.CountExecutions(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count executions: %w", err)
	}

	response := &dtos.ListExecutionsResponse{
		Executions: make([]dtos.WorkflowExecutionResponse, len(executions)),
		Total:      total,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	for i, execution := range executions {
		response.Executions[i] = *dtos.ToExecutionResponse(execution)
	}

	return response, nil
}

func (s *ExecutionService) CancelExecution(ctx context.Context, tenantID, executionID string) error {
	s.logger.InfoContext(ctx, "Cancelling workflow execution",
		logger.String("tenant_id", tenantID),
		logger.String("execution_id", executionID))

	execution, err := s.executionRepo.GetExecutionByID(ctx, tenantID, executionID)
	if err != nil {
		return err
	}

	if execution.Status == domain.ExecutionStatusCompleted {
		return domain.ErrExecutionCompleted
	}

	if execution.Status == domain.ExecutionStatusFailed {
		return domain.ErrExecutionFailed
	}
	// Update execution status to cancelled
	execution.Status = domain.ExecutionStatusCancelled
	if err := s.executionRepo.UpdateExecution(ctx, execution, nil); err != nil {
		s.logger.Error("Failed to cancel execution",
			logger.String("tenant_id", tenantID),
			logger.String("execution_id", executionID),
			logger.Err(err))
		return fmt.Errorf("failed to cancel execution: %w", err)
	}

	s.logger.InfoContext(ctx, "Workflow execution cancelled successfully",
		logger.String("tenant_id", tenantID),
		logger.String("execution_id", executionID))

	return nil
}

func (s *ExecutionService) RetryExecution(ctx context.Context, tenantID, executionID string) error {
	s.logger.InfoContext(ctx, "Retrying workflow execution",
		logger.String("tenant_id", tenantID),
		logger.String("execution_id", executionID))

	execution, err := s.executionRepo.GetExecutionByID(ctx, tenantID, executionID)
	if err != nil {
		return err
	}

	if execution.Status != domain.ExecutionStatusFailed {
		return fmt.Errorf("only failed executions can be retried")
	}

	// Reset execution status
	execution.Status = domain.ExecutionStatusPending
	execution.ErrorMessage = ""

	// Reset failed step executions
	for i := range execution.Steps {
		if execution.Steps[i].Status == domain.ExecutionStatusFailed {
			execution.Steps[i].Status = domain.ExecutionStatusPending
			execution.Steps[i].ErrorMessage = ""
			execution.Steps[i].RetryCount++
		}
	}	
	if err := s.executionRepo.UpdateExecution(ctx, execution, nil); err != nil {
		s.logger.Error("Failed to retry execution",
			logger.String("tenant_id", tenantID),
			logger.String("execution_id", executionID),
			logger.Err(err))
		return fmt.Errorf("failed to retry execution: %w", err)
	}

	s.logger.InfoContext(ctx, "Workflow execution retry initiated",
		logger.String("tenant_id", tenantID),
		logger.String("execution_id", executionID))

	return nil
}

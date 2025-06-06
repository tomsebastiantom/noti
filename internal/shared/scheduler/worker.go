package scheduler

import (
	"context"
	"fmt"
	"time"

	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/workerpool"
	"github.com/robfig/cron/v3"
)

// Worker handles scheduling operations and job execution
type Worker struct {
	repository        Repository
	workerPool        *workerpool.WorkerPool
	logger            logger.Logger
	stopCh            chan struct{}
	pollInterval      time.Duration
	executionBatchSize int
}

// NewWorker creates a new scheduler worker
func NewWorker(
	repository Repository,
	workerPool *workerpool.WorkerPool,
	logger logger.Logger,
) *Worker {
	return &Worker{
		repository:         repository,
		workerPool:         workerPool,
		logger:             logger,
		stopCh:             make(chan struct{}),
		pollInterval:       time.Minute, // Poll every minute
		executionBatchSize: 50,          // Process 50 schedules at a time
	}
}

// Start begins the scheduler worker polling loop
func (w *Worker) Start(ctx context.Context) error {
	w.logger.InfoContext(ctx, "Starting scheduler worker",
		logger.Duration("poll_interval", w.pollInterval),
		logger.Int("batch_size", w.executionBatchSize))

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.InfoContext(ctx, "Scheduler worker stopping due to context cancellation")
			return ctx.Err()
		case <-w.stopCh:
			w.logger.InfoContext(ctx, "Scheduler worker stopping due to stop signal")
			return nil
		case <-ticker.C:
			if err := w.processSchedules(ctx); err != nil {
				w.logger.Error("Error processing schedules",
					logger.Err(err))
			}
		}
	}
}

// Stop stops the scheduler worker
func (w *Worker) Stop() {
	close(w.stopCh)
}

// SetPollInterval sets the polling interval for the worker
func (w *Worker) SetPollInterval(interval time.Duration) {
	w.pollInterval = interval
}

// processSchedules finds and processes schedules that are due for execution
func (w *Worker) processSchedules(ctx context.Context) error {
	w.logger.DebugContext(ctx, "Processing due schedules")

	// Get schedules that are due for execution
	now := time.Now()
	schedules, err := w.repository.GetSchedulesDueForExecution(ctx, now, w.executionBatchSize)
	if err != nil {
		return fmt.Errorf("failed to get schedules due for execution: %w", err)
	}

	if len(schedules) == 0 {
		w.logger.DebugContext(ctx, "No schedules due for execution")
		return nil
	}

	w.logger.InfoContext(ctx, "Found schedules due for execution",
		logger.Int("count", len(schedules)))

	// Process each schedule
	for _, schedule := range schedules {		if err := w.processSchedule(ctx, schedule); err != nil {
			w.logger.Error("Failed to process schedule",
				logger.String("schedule_id", schedule.ID.String()),
				logger.String("schedule_name", schedule.Name),
				logger.Err(err))
			continue
		}
	}

	return nil
}

// processSchedule processes a single schedule
func (w *Worker) processSchedule(ctx context.Context, schedule *Schedule) error {
	w.logger.InfoContext(ctx, "Processing schedule",
		logger.String("schedule_id", schedule.ID.String()),
		logger.String("schedule_name", schedule.Name),
		logger.String("schedule_type", string(schedule.Type)))

	// Create execution record
	execution := &ScheduleExecution{
		ScheduleID: schedule.ID,
		Status:     ExecutionStatusPending,
	}

	execution, err := w.repository.CreateExecution(ctx, execution)
	if err != nil {
		return fmt.Errorf("failed to create execution record: %w", err)
	}

	// Create and submit job to worker pool
	job := NewScheduleJob(schedule, execution, w.repository, w.logger)
	
	if err := w.workerPool.Submit(job); err != nil {
		// Update execution to failed if we can't submit the job
		execution.Status = ExecutionStatusFailed
		execution.ErrorMessage = stringPtr(fmt.Sprintf("Failed to submit job: %v", err))
				if _, updateErr := w.repository.UpdateExecution(ctx, execution); updateErr != nil {
			w.logger.Error("Failed to update execution after job submission failure",
				logger.String("execution_id", execution.ID.String()),
				logger.Err(updateErr))
		}
		
		return fmt.Errorf("failed to submit job to worker pool: %w", err)
	}
	// Update schedule's next execution time
	if err := w.updateNextExecution(ctx, schedule); err != nil {
		w.logger.Error("Failed to update next execution time",
			logger.String("schedule_id", schedule.ID.String()),
			logger.Err(err))
		// Don't return error here as the job has been submitted successfully
	}

	w.logger.InfoContext(ctx, "Schedule processed successfully",
		logger.String("schedule_id", schedule.ID.String()),
		logger.String("execution_id", execution.ID.String()))

	return nil
}

// updateNextExecution calculates and updates the next execution time for a schedule
func (w *Worker) updateNextExecution(ctx context.Context, schedule *Schedule) error {
	// Parse cron expression
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	cronSchedule, err := parser.Parse(schedule.CronExpression)
	if err != nil {
		return fmt.Errorf("failed to parse cron expression '%s': %w", schedule.CronExpression, err)
	}

	// Calculate next execution time
	now := time.Now()
	nextExecution := cronSchedule.Next(now)
	// Update in database
	if err := w.repository.UpdateScheduleNextExecution(ctx, schedule.ID.String(), nextExecution); err != nil {
		return fmt.Errorf("failed to update next execution time: %w", err)
	}

	w.logger.DebugContext(ctx, "Updated next execution time",
		logger.String("schedule_id", schedule.ID.String()),
		logger.Time("next_execution", nextExecution))

	return nil
}

// ProcessRetries processes failed executions that should be retried
func (w *Worker) ProcessRetries(ctx context.Context, maxRetries int) error {
	w.logger.DebugContext(ctx, "Processing failed executions for retry")

	// Get pending executions (which could include retries)
	executions, err := w.repository.GetPendingExecutions(ctx, w.executionBatchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending executions: %w", err)
	}

	if len(executions) == 0 {
		w.logger.DebugContext(ctx, "No pending executions found")
		return nil
	}

	w.logger.InfoContext(ctx, "Found pending executions",
		logger.Int("count", len(executions)))
	// Process each execution
	for _, execution := range executions {
		// Get the associated schedule
		schedule, err := w.repository.GetScheduleByID(ctx, execution.ScheduleID.String())
		if err != nil {
			w.logger.Error("Failed to get schedule for execution",
				logger.String("execution_id", execution.ID.String()),
				logger.String("schedule_id", execution.ScheduleID.String()),
				logger.Err(err))
			continue
		}

		// Create retry job
		retryJob := NewRetryJob(execution, schedule, w.repository, w.logger)
		
		if err := w.workerPool.Submit(retryJob); err != nil {
			w.logger.Error("Failed to submit retry job",
				logger.String("execution_id", execution.ID.String()),
				logger.Err(err))
			continue
		}

		w.logger.InfoContext(ctx, "Retry job submitted",
			logger.String("execution_id", execution.ID.String()),
			logger.String("schedule_id", execution.ScheduleID.String()))
	}
	return nil
}

// GetStats returns current scheduler statistics
func (w *Worker) GetStats(ctx context.Context) (*ScheduleStats, error) {
	return w.repository.GetScheduleStats(ctx)
}

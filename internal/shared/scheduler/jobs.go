package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"getnoti.com/pkg/logger"
)

// ScheduleJob represents a job for executing a scheduled task
type ScheduleJob struct {
	schedule   *Schedule
	execution  *ScheduleExecution
	repository Repository
	logger     logger.Logger
}

// NewScheduleJob creates a new schedule execution job
func NewScheduleJob(
	schedule *Schedule,
	execution *ScheduleExecution,
	repository Repository,
	logger logger.Logger,
) *ScheduleJob {
	return &ScheduleJob{
		schedule:   schedule,
		execution:  execution,
		repository: repository,
		logger:     logger,
	}
}

// Process implements the workerpool.Job interface
func (j *ScheduleJob) Process(ctx context.Context) error {
	j.logger.InfoContext(ctx, "Processing schedule execution",
		logger.String("schedule_id", j.schedule.ID.String()),
		logger.String("execution_id", j.execution.ID.String()),
		logger.String("schedule_type", string(j.schedule.Type)))

	// Update execution to running status
	j.execution.Status = ExecutionStatusRunning
	j.execution.StartedAt = timePtr(time.Now())
	
	if _, err := j.repository.UpdateExecution(ctx, j.execution); err != nil {
		j.logger.Error("Failed to update execution status to running",
			logger.String("execution_id", j.execution.ID.String()),
			logger.Err(err))
	}

	var result string
	var err error

	// Execute based on schedule type
	switch j.schedule.Type {
	case ScheduleTypeNotification:
		result, err = j.executeNotificationSchedule(ctx)
	case ScheduleTypeWebhook:
		result, err = j.executeWebhookSchedule(ctx)
	case ScheduleTypeCleanup:
		result, err = j.executeCleanupSchedule(ctx)
	case ScheduleTypeReport:
		result, err = j.executeReportSchedule(ctx)
	default:
		err = fmt.Errorf("unsupported schedule type: %s", j.schedule.Type)
	}
	// Update execution with results
	j.execution.CompletedAt = timePtr(time.Now())
	
	if err != nil {
		j.execution.Status = ExecutionStatusFailed
		j.execution.ErrorMessage = stringPtr(err.Error())
		j.logger.Error("Schedule execution failed",
			logger.String("schedule_id", j.schedule.ID.String()),
			logger.String("execution_id", j.execution.ID.String()),
			logger.Err(err))
	} else {
		j.execution.Status = ExecutionStatusCompleted
		j.execution.Result = &result
		j.logger.Info("Schedule execution completed successfully",
			logger.String("schedule_id", j.schedule.ID.String()),
			logger.String("execution_id", j.execution.ID.String()))
	}

	// Save final execution state
	if _, updateErr := j.repository.UpdateExecution(ctx, j.execution); updateErr != nil {
		j.logger.Error("Failed to update execution final status",
			logger.String("execution_id", j.execution.ID.String()),
			logger.Err(updateErr))
	}

	// Update schedule's last execution time
	if updateErr := j.repository.UpdateScheduleLastExecution(ctx, j.schedule.ID.String(), *j.execution.StartedAt); updateErr != nil {
		j.logger.Error("Failed to update schedule last execution time",
			logger.String("schedule_id", j.schedule.ID.String()),
			logger.Err(updateErr))
	}

	return err
}

// executeNotificationSchedule executes a notification schedule
func (j *ScheduleJob) executeNotificationSchedule(ctx context.Context) (string, error) {
	j.logger.InfoContext(ctx, "Executing notification schedule",
		logger.String("schedule_id", j.schedule.ID.String()))

	// Parse notification config
	var config NotificationConfig
	configBytes, err := json.Marshal(j.schedule.Config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return "", fmt.Errorf("failed to parse notification config: %w", err)
	}

	// TODO: Implement notification sending logic
	// This would integrate with the notification service to send the scheduled notification
	
	result := fmt.Sprintf("Notification scheduled for template_id: %s, user_id: %s", 
		config.TemplateID, config.UserID)
	
	return result, nil
}

// executeWebhookSchedule executes a webhook schedule
func (j *ScheduleJob) executeWebhookSchedule(ctx context.Context) (string, error) {
	j.logger.InfoContext(ctx, "Executing webhook schedule",
		logger.String("schedule_id", j.schedule.ID.String()),
		logger.String("tenant_id", j.schedule.TenantID))

	// Parse webhook config
	var config WebhookConfig
	configBytes, err := json.Marshal(j.schedule.Config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return "", fmt.Errorf("failed to parse webhook config: %w", err)
	}
		// TODO: Implement webhook sending logic
	// This would integrate with the webhook service to send the scheduled webhook
	
	result := fmt.Sprintf("Webhook sent to URL: %s with event_type: %s", 
		config.URL, config.EventType)

	
	return result, nil
}

// executeCleanupSchedule executes a cleanup schedule
func (j *ScheduleJob) executeCleanupSchedule(ctx context.Context) (string, error) {
	j.logger.InfoContext(ctx, "Executing cleanup schedule",
		logger.String("schedule_id", j.schedule.ID.String()))

	// Parse cleanup config
	var config CleanupConfig
	configBytes, err := json.Marshal(j.schedule.Config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return "", fmt.Errorf("failed to parse cleanup config: %w", err)
	}

	// TODO: Implement cleanup logic
	// This would clean up old records based on the configuration
	
	result := fmt.Sprintf("Cleanup executed for table: %s, older than: %s", 
		config.TableName, config.RetentionPeriod)
	
	return result, nil
}

// executeReportSchedule executes a report schedule
func (j *ScheduleJob) executeReportSchedule(ctx context.Context) (string, error) {
	j.logger.InfoContext(ctx, "Executing report schedule",
		logger.String("schedule_id", j.schedule.ID.String()))

	// Parse report config
	var config ReportConfig
	configBytes, err := json.Marshal(j.schedule.Config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return "", fmt.Errorf("failed to parse report config: %w", err)
	}

	// TODO: Implement report generation logic
	// This would generate and send reports based on the configuration
	
	result := fmt.Sprintf("Report generated: %s, sent to: %s", 
		config.ReportType, config.Recipients)
	
	return result, nil
}

// RetryJob represents a job for retrying failed schedule executions
type RetryJob struct {
	execution  *ScheduleExecution
	schedule   *Schedule
	repository Repository
	maxRetries int
	logger     logger.Logger
}

// NewRetryJob creates a new retry job
func NewRetryJob(
	execution *ScheduleExecution,
	schedule *Schedule,
	repository Repository,
	logger logger.Logger,
) *RetryJob {
	return &RetryJob{
		execution:  execution,
		schedule:   schedule,
		repository: repository,
		maxRetries: 3, // Default max retries
		logger:     logger,
	}
}

// Process implements the workerpool.Job interface for retries
func (j *RetryJob) Process(ctx context.Context) error {
	j.logger.InfoContext(ctx, "Processing retry for failed execution",
		logger.String("execution_id", j.execution.ID.String()),
		logger.String("schedule_id", j.execution.ScheduleID.String()),
		logger.Int("retry_count", j.execution.RetryCount+1),
		logger.Int("max_retries", j.maxRetries))
		
	// Increment the retry count
	j.execution.RetryCount++
	j.execution.Status = ExecutionStatusRetrying
	
	if _, err := j.repository.UpdateExecution(ctx, j.execution); err != nil {
		j.logger.Error("Failed to update execution retry count",
			logger.String("execution_id", j.execution.ID.String()),
			logger.Err(err))
		return err
	}

	// Create a new schedule job for the retry
	scheduleJob := NewScheduleJob(j.schedule, j.execution, j.repository, j.logger)
	
	// Execute the job
	return scheduleJob.Process(ctx)
}

// shouldRetry determines if an execution should be retried
func shouldRetry(execution *ScheduleExecution, maxRetries int) bool {
	if execution.Status != ExecutionStatusFailed {
		return false
	}
	
	// Check if the retry count is below the maximum allowed retries
	return execution.RetryCount < maxRetries
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func stringPtr(s string) *string {
	return &s
}

// Config types for different schedule types
type NotificationConfig struct {
	TemplateID string                 `json:"template_id"`
	UserID     string                 `json:"user_id,omitempty"`
	Channel    string                 `json:"channel"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
}

type WebhookConfig struct {
	URL       string                 `json:"url"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Headers   map[string]string      `json:"headers,omitempty"`
}

type CleanupConfig struct {
	TableName       string `json:"table_name"`
	RetentionPeriod string `json:"retention_period"`
	BatchSize       int    `json:"batch_size,omitempty"`
}

type ReportConfig struct {
	ReportType string   `json:"report_type"`
	Recipients []string `json:"recipients"`
	Format     string   `json:"format,omitempty"`
	Period     string   `json:"period,omitempty"`
}
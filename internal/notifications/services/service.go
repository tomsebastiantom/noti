package notifications

import (
	"context"
	"fmt"
	"time"

	"getnoti.com/internal/notifications/domain"
	repos "getnoti.com/internal/notifications/repos"
	tenantServices "getnoti.com/internal/tenants/services"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/workerpool"
)

// NotificationService handles notification-related business operations
type NotificationService struct {
	notificationRepo  repos.NotificationRepository
	tenantService     *tenantServices.TenantService
	connectionManager *db.Manager
	queueManager      *queue.QueueManager
	workerPoolManager *workerpool.WorkerPoolManager
	logger            logger.Logger
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	notificationRepo repos.NotificationRepository,
	tenantService *tenantServices.TenantService,
	connectionManager *db.Manager,
	queueManager *queue.QueueManager,
	workerPoolManager *workerpool.WorkerPoolManager,
	logger logger.Logger,
) *NotificationService {
	return &NotificationService{
		notificationRepo:  notificationRepo,
		tenantService:     tenantService,
		connectionManager: connectionManager,
		queueManager:      queueManager,
		workerPoolManager: workerPoolManager,
		logger:            logger,
	}
}

// SendNotificationRequest represents a notification sending request
type SendNotificationRequest struct {
	TenantID     string                 `json:"tenant_id"`
	Channel      string                 `json:"channel"`
	Recipients   []string               `json:"recipients"`
	Subject      string                 `json:"subject"`
	Body         string                 `json:"body"`
	TemplateID   string                 `json:"template_id,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	Priority     string                 `json:"priority,omitempty"`
	ScheduledFor string                 `json:"scheduled_for,omitempty"`
}

// SendNotificationResponse represents the response after sending notification
type SendNotificationResponse struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

// SendNotification sends a notification for a tenant
func (s *NotificationService) SendNotification(ctx context.Context, req SendNotificationRequest) (*SendNotificationResponse, error) {
	s.logger.InfoContext(ctx, "Processing send notification request",
		logger.String("tenant_id", req.TenantID),
		logger.String("channel", req.Channel))

	// Validate tenant access
	err := s.tenantService.ValidateTenantAccess(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant validation failed: %w", err)
	}
	// Create notification domain object
	notification := &domain.Notification{
		ID:       req.TenantID + "_" + req.Channel, // Generate proper ID
		TenantID: req.TenantID,
		UserID:   req.Recipients[0], // Use first recipient for now
		Type:     req.Channel,
		Channel:  req.Channel,
		Status:   "pending",
		Content:  req.Body,
	}
	// Set optional fields
	if req.TemplateID != "" {
		notification.TemplateID = req.TemplateID
	}
	if req.Variables != nil {
		// Convert map[string]interface{} to []TemplateVariable
		variables := make([]domain.TemplateVariable, 0, len(req.Variables))
		for key, value := range req.Variables {
			variables = append(variables, domain.TemplateVariable{
				Key:   key,
				Value: fmt.Sprintf("%v", value), // Convert to string
			})
		}
		notification.Variables = variables
	}

	// Save notification to repository	err = s.notificationRepo.CreateNotification(ctx, notification)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to save notification",
			logger.String("tenant_id", req.TenantID),
			logger.Err(err))
		return nil, fmt.Errorf("failed to save notification: %w", err)
	}

	// Queue notification for processing
	err = s.queueNotification(ctx, notification)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to queue notification",
			logger.String("notification_id", notification.ID),
			logger.Err(err))
		return nil, fmt.Errorf("failed to queue notification: %w", err)
	}

	s.logger.InfoContext(ctx, "Notification queued successfully",
		logger.String("notification_id", notification.ID),
		logger.String("tenant_id", req.TenantID))

	return &SendNotificationResponse{
		NotificationID: notification.ID,
		Status:         "queued",
		Message:        "Notification queued for processing",
	}, nil
}

// GetNotification retrieves a notification by ID for a tenant
func (s *NotificationService) GetNotification(ctx context.Context, tenantID, notificationID string) (*domain.Notification, error) {
	s.logger.DebugContext(ctx, "Getting notification",
		logger.String("tenant_id", tenantID),
		logger.String("notification_id", notificationID))

	// Validate tenant access
	err := s.tenantService.ValidateTenantAccess(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant validation failed: %w", err)
	}

	// Get notification from repository
	notification, err := s.notificationRepo.GetNotificationByID(ctx, notificationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Verify notification belongs to tenant
	if notification.TenantID != tenantID {
		return nil, fmt.Errorf("notification does not belong to tenant")
	}

	return notification, nil
}

// queueNotification queues a notification for processing
func (s *NotificationService) queueNotification(ctx context.Context, notification *domain.Notification) error {
	// Get queue for tenant
	notificationQueue, err := s.queueManager.GetOrCreateQueue(fmt.Sprintf("notifications_%s", notification.TenantID))
	if err != nil {
		return fmt.Errorf("failed to get queue: %w", err)
	}
	
	// Queue the notification
	message := queue.Message{
		Body:      []byte(notification.Content),
		Headers:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}
	
	err = notificationQueue.Publish(ctx, "notifications", notification.TenantID, message)
	if err != nil {
		return fmt.Errorf("failed to publish notification: %w", err)
	}

	return nil
}

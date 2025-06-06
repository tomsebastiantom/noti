package container

import (
	"context"
	"fmt"
	"time"

	notificationHandlers "getnoti.com/internal/notifications/events/handlers"
	notificationServices "getnoti.com/internal/notifications/services"
	providerServices "getnoti.com/internal/providers/services"
	sharedEvents "getnoti.com/internal/shared/events"
	templateServices "getnoti.com/internal/templates/services"
	tenantHandlers "getnoti.com/internal/tenants/events/handlers"
	tenantServices "getnoti.com/internal/tenants/services"
	webhookHandlers "getnoti.com/internal/webhooks/events/handlers"
	webhookServices "getnoti.com/internal/webhooks/services"
	workflowEngine "getnoti.com/internal/workflows/engine"
	workflowHandlers "getnoti.com/internal/workflows/events/handlers"
	workflowServices "getnoti.com/internal/workflows/services"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/webhook"
	"getnoti.com/pkg/workerpool"
)

// initializeServices sets up all application services
func (c *ServiceContainer) initializeServices() error {
	c.logger.Info("Initializing application services")
		// Initialize hybrid event bus (leverages existing infrastructure)
	c.eventBus = sharedEvents.NewHybridEventBus(
		c.dbManager,
		c.workerPoolManager,
		c.logger,
	)
	c.logger.Info("Hybrid event bus initialized successfully")
		// Initialize tenant service
	c.tenantService = tenantServices.NewTenantService(
		c.tenantRepo,
		c.userRepo,
		c.dbManager,
		c.configResolver,
		c.logger,
	)
	c.logger.Info("Tenant service initialized successfully")	// Initialize user preference service
	c.userPreferenceService = tenantServices.NewUserPreferenceService(
		c.dbManager,
		c.logger,
		c.repositoryFactory,
	)
	c.logger.Info("User preference service initialized successfully")
		// Initialize notification service with event bus
	c.notificationService = notificationServices.NewNotificationService(
		c.notificationRepo,
		c.tenantService,
		c.dbManager,
		c.queueManager,
		c.workerPoolManager,
		c.logger,
	)
	c.logger.Info("Notification service initialized successfully")
		// Initialize template service
	c.templateService = templateServices.NewTemplateService(
		c.templateRepo,
		c.tenantService,
		c.logger,
	)
	c.logger.Info("Template service initialized successfully")
	
	// Initialize provider service with event bus
	var notificationQueue queue.Queue
	
	// Try to get a notification queue but don't fail if not configured
	if c.config.Queue.URL != "" {
		var err error
		notificationQueue, err = c.queueManager.GetOrCreateQueue("notifications")
		if err != nil {
			c.logger.Warn("Failed to initialize notification queue, proceeding without queuing support",
				logger.Field{Key: "error", Value: err.Error()})
		}
	} else {
		c.logger.Warn("No queue URL configured, notifications will be processed synchronously")
	}
		c.providerService = providerServices.NewProviderService(
		c.providerRepo,
		c.tenantService,
		c.credentialManager,
		c.cache,
		c.providerFactory,
		notificationQueue,
		c.workerPoolManager,
		c.userPreferenceService,
		c.logger,
	)
	c.logger.Info("Provider service initialized successfully")	// Initialize webhook service with event bus
	webhookSecurityManager := webhook.NewSecurityManager()
	c.webhookService = webhookServices.NewWebhookService(
		c.dbManager,
		c.queueManager,
		webhookSecurityManager,
		c.webhookSender,
		c.logger,
		c.repositoryFactory,
	)
	c.logger.Info("Webhook service initialized successfully")

	// Initialize workflow service
	c.workflowService = workflowServices.NewWorkflowService(
		c.workflowRepo,
		c.executionRepo,
		c.logger,
	)
	c.logger.Info("Workflow service initialized successfully")
	// Initialize workflow engine with event bus
	workflowWorkerPool := c.workerPoolManager.GetOrCreatePool(workerpool.WorkerPoolConfig{
		Name:           "workflow_engine",
		InitialWorkers: 5,
		MaxJobs:        100,
		MinWorkers:     2,
		MaxWorkers:     10,
		ScaleFactor:    1.5,
		IdleTimeout:    5 * time.Minute,
		ScaleInterval:  30 * time.Second,
	})
	c.workflowEngine = workflowEngine.NewWorkflowEngine(
		c.workflowRepo,
		c.executionRepo,
		workflowWorkerPool,
		c.logger,
		c.eventBus,
		c.notificationService,
		30 * time.Second, // Set poll interval to 30 seconds
	)
	
	c.logger.Info("Workflow engine initialized successfully")
	
	// Start the workflow engine
	if err := c.workflowEngine.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start workflow engine: %w", err)
	}
	c.logger.Info("Workflow engine started")
	
	// Start the event bus
	if err := c.eventBus.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	// Register event handlers after all services are initialized
	if err := c.registerEventHandlers(); err != nil {
		return fmt.Errorf("failed to register event handlers: %w", err)
	}

	c.logger.Info("Application services initialization completed successfully")
	return nil
}

// registerEventHandlers registers all domain event handlers with the event bus
func (c *ServiceContainer) registerEventHandlers() error {
	c.logger.Info("Registering domain event handlers")

	// Create notification event handlers
	notificationHandlerInstance := notificationHandlers.NewNotificationEventHandlers(c.logger)
	for eventType, handler := range notificationHandlerInstance.GetHandlerMethods() {
		if err := c.eventBus.Subscribe(eventType, handler); err != nil {
			return fmt.Errorf("failed to register notification handler for %s: %w", eventType, err)
		}
		c.logger.Info("Registered notification event handler", 
			logger.Field{Key: "event_type", Value: eventType})
	}

	// Create tenant event handlers  
	tenantHandlerInstance := tenantHandlers.NewTenantEventHandlers(c.logger)
	for eventType, handler := range tenantHandlerInstance.GetHandlerMethods() {
		if err := c.eventBus.Subscribe(eventType, handler); err != nil {
			return fmt.Errorf("failed to register tenant handler for %s: %w", eventType, err)
		}
		c.logger.Info("Registered tenant event handler", 
			logger.Field{Key: "event_type", Value: eventType})
	}

	// Create webhook event handlers
	webhookHandlerInstance := webhookHandlers.NewWebhookEventHandlers(c.logger)
	for eventType, handler := range webhookHandlerInstance.GetHandlerMethods() {
		if err := c.eventBus.Subscribe(eventType, handler); err != nil {
			return fmt.Errorf("failed to register webhook handler for %s: %w", eventType, err)
		}
		c.logger.Info("Registered webhook event handler", 
			logger.Field{Key: "event_type", Value: eventType})
	}
		// Create workflow event handlers
	workflowHandlerInstance := workflowHandlers.NewWorkflowEventHandlers(c.logger)
	for eventType, handler := range workflowHandlerInstance.GetHandlerMethods() {
		if err := c.eventBus.Subscribe(eventType, handler); err != nil {
			return fmt.Errorf("failed to register workflow handler for %s: %w", eventType, err)
		}
		c.logger.Info("Registered workflow event handler", 
			logger.Field{Key: "event_type", Value: eventType})
	}

	c.logger.Info("All domain event handlers registered successfully")
	return nil
}

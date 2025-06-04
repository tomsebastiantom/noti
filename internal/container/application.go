package container

import (
	notificationServices "getnoti.com/internal/notifications/services"
	providerServices "getnoti.com/internal/providers/services"
	templateServices "getnoti.com/internal/templates/services"
	tenantServices "getnoti.com/internal/tenants/services"
	webhookServices "getnoti.com/internal/webhooks/services"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/webhook"
)

// initializeServices sets up all application services
func (c *ServiceContainer) initializeServices() error {
	c.logger.Info("Initializing application services")
	// Initialize tenant service
	c.tenantService = tenantServices.NewTenantService(
		c.tenantRepo,
		c.userRepo,
		c.dbManager,
		c.configResolver,
		c.logger,
	)
	c.logger.Info("Tenant service initialized successfully")
	// Initialize notification service  
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
		// Initialize provider service
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
		c.logger,
	)
	c.logger.Info("Provider service initialized successfully")	// Initialize webhook service
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

	c.logger.Info("Application services initialization completed successfully")
	return nil
}

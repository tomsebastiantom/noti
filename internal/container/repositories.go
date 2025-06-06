package container

import (
	notificationRepos "getnoti.com/internal/notifications/repos/implementations"
	"getnoti.com/internal/providers/infra/providers"
	providerRepos "getnoti.com/internal/providers/repos/implementations"
	templateRepos "getnoti.com/internal/templates/repos/implementations"
	tenantRepos "getnoti.com/internal/tenants/repos/implementations"
	webhookRepos "getnoti.com/internal/webhooks/repos/implementations"
	workflowRepos "getnoti.com/internal/workflows/repos/implementations"
)

// initializeRepositories sets up all domain repositories
func (c *ServiceContainer) initializeRepositories() error {
	c.logger.Info("Initializing repositories")

	c.tenantRepo = tenantRepos.NewTenantRepository(c.mainDB, c.credentialManager)
	c.logger.Info("Tenant repository initialized successfully")
	// Initialize user repository (uses main DB)
	c.userRepo = tenantRepos.NewUserRepository(c.mainDB)
	c.logger.Info("User repository initialized successfully")
	// Initialize provider repository (uses main DB for provider metadata)
	c.providerRepo = providerRepos.NewProviderRepository(c.mainDB)
	c.logger.Info("Provider repository initialized successfully")
	
	// Initialize notification repository (uses main DB for notification metadata)
	c.notificationRepo = notificationRepos.NewNotificationRepository(c.mainDB)
	c.logger.Info("Notification repository initialized successfully")
	
	// Initialize template repository (uses main DB for template metadata)
	c.templateRepo = templateRepos.NewTemplateRepository(c.mainDB)
	c.logger.Info("Template repository initialized successfully")
	
	// Initialize webhook repository (uses main DB for webhook metadata)
	c.webhookRepo = webhookRepos.NewWebhookRepository(c.mainDB)
	c.logger.Info("Webhook repository initialized successfully")

	// Initialize provider factory (needs provider repo, cache, and credential manager)
	c.providerFactory = providers.NewProviderFactory(
		c.cache,
		c.providerRepo,
		c.credentialManager,	)
	c.logger.Info("Provider factory initialized successfully")
	
	// Initialize workflow repositories
	c.workflowRepo = workflowRepos.NewWorkflowRepository(c.mainDB)
	c.logger.Info("Workflow repository initialized successfully")
	
	c.executionRepo = workflowRepos.NewExecutionRepository(c.mainDB)
	c.logger.Info("Workflow execution repository initialized successfully")
	
	// Initialize repository factory for tenant-specific repositories
	c.repositoryFactory = NewRepositoryFactory(
		c.dbManager,
		c.credentialManager,
		c.logger,
	)
	c.logger.Info("Repository factory initialized successfully")

	c.logger.Info("Repository initialization completed successfully")
	return nil
}

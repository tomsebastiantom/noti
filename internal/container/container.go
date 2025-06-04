package container

import (
	"context"
	"fmt"

	"getnoti.com/config"
	notificationRepos "getnoti.com/internal/notifications/repos"
	notificationServices "getnoti.com/internal/notifications/services"
	"getnoti.com/internal/providers/infra/providers"
	providerRepos "getnoti.com/internal/providers/repos"
	providerServices "getnoti.com/internal/providers/services"
	templateRepos "getnoti.com/internal/templates/repos"
	templateServices "getnoti.com/internal/templates/services"
	tenantRepos "getnoti.com/internal/tenants/repos"
	tenantServices "getnoti.com/internal/tenants/services"
	webhookRepos "getnoti.com/internal/webhooks/repos"
	webhookServices "getnoti.com/internal/webhooks/services"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/credentials"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/webhook"
	"getnoti.com/pkg/workerpool"
)

// ServiceContainer manages all application dependencies
type ServiceContainer struct {
	// Configuration
	config *config.Config
	logger logger.Logger

	// Infrastructure Services
	mainDB            db.Database
	dbManager         *db.Manager
	cache             *cache.GenericCache
	credentialManager *credentials.Manager
	queueManager      *queue.QueueManager
	workerPoolManager *workerpool.WorkerPoolManager
	// Infrastructure Interfaces
	configResolver    db.ConfigResolver
	// Provider Factory (needed for provider service)
	providerFactory *providers.ProviderFactory
	
	// Webhook Infrastructure
	webhookSender *webhook.Sender

	// Application Services
	tenantService       *tenantServices.TenantService
	notificationService *notificationServices.NotificationService
	templateService     *templateServices.TemplateService
	providerService     *providerServices.ProviderService
	webhookService      *webhookServices.WebhookService
	userPreferenceService *tenantServices.UserPreferenceService
	
	// Repositories
	tenantRepo       tenantRepos.TenantsRepository
	userRepo         tenantRepos.UserRepository
	notificationRepo notificationRepos.NotificationRepository
	templateRepo     templateRepos.TemplateRepository
	webhookRepo      webhookRepos.WebhookRepository
	repositoryFactory *RepositoryFactory
	providerRepo     providerRepos.ProviderRepository
}

// NewServiceContainer creates and initializes the service container
func NewServiceContainer(cfg *config.Config, log logger.Logger) (*ServiceContainer, error) {
	container := &ServiceContainer{
		config: cfg,
		logger: log,
	}

	// Initialize in order: Infrastructure -> Repositories -> Services
	if err := container.initializeInfrastructure(); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	if err := container.initializeRepositories(); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	if err := container.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	return container, nil
}

// Service Getters
func (c *ServiceContainer) GetTenantService() *tenantServices.TenantService {
	return c.tenantService
}

func (c *ServiceContainer) GetNotificationService() *notificationServices.NotificationService {
	return c.notificationService
}

func (c *ServiceContainer) GetTemplateService() *templateServices.TemplateService {
	return c.templateService
}

func (c *ServiceContainer) GetProviderService() *providerServices.ProviderService {
	return c.providerService
}

func (c *ServiceContainer) GetWebhookService() *webhookServices.WebhookService {
	return c.webhookService
}

func (c *ServiceContainer) GetUserPreferenceService() *tenantServices.UserPreferenceService {
	return c.userPreferenceService
}

func (c *ServiceContainer) GetWebhookSender() *webhook.Sender {
	return c.webhookSender
}

// Repository Getters
func (c *ServiceContainer) GetTenantRepository() tenantRepos.TenantsRepository {
	return c.tenantRepo
}

func (c *ServiceContainer) GetUserRepository() tenantRepos.UserRepository {
	return c.userRepo
}

func (c *ServiceContainer) GetNotificationRepository() notificationRepos.NotificationRepository {
	return c.notificationRepo
}

func (c *ServiceContainer) GetTemplateRepository() templateRepos.TemplateRepository {
	return c.templateRepo
}

func (c *ServiceContainer) GetProviderRepository() providerRepos.ProviderRepository {
	return c.providerRepo
}

func (c *ServiceContainer) GetWebhookRepository() webhookRepos.WebhookRepository {
	return c.webhookRepo
}

// Infrastructure Getters
func (c *ServiceContainer) GetInfrastructure() (*Infrastructure, error) {
	return &Infrastructure{
		MainDB:            c.mainDB,
		DBManager:         c.dbManager,
		Cache:             c.cache,
		CredentialManager: c.credentialManager,
		QueueManager:      c.queueManager,
		WorkerPoolManager: c.workerPoolManager,
		Logger:            c.logger,
	}, nil
}

// Update repository getters to use the factory

func (c *ServiceContainer) GetNotificationRepositoryForTenant(tenantID string) (notificationRepos.NotificationRepository, error) {
    return c.repositoryFactory.GetNotificationRepositoryForTenant(tenantID)
}

func (c *ServiceContainer) GetTemplateRepositoryForTenant(tenantID string) (templateRepos.TemplateRepository, error) {
    return c.repositoryFactory.GetTemplateRepositoryForTenant(tenantID)
}

func (c *ServiceContainer) GetProviderRepositoryForTenant(tenantID string) (providerRepos.ProviderRepository, error) {
    return c.repositoryFactory.GetProviderRepositoryForTenant(tenantID)
}

func (c *ServiceContainer) GetWebhookRepositoryForTenant(tenantID string) (webhookRepos.WebhookRepository, error) {
    return c.repositoryFactory.GetWebhookRepositoryForTenant(tenantID)
}

func (c *ServiceContainer) GetUserPreferenceRepositoryForTenant(tenantID string) (tenantRepos.UserPreferenceRepository, error) {
    return c.repositoryFactory.GetUserPreferenceRepositoryForTenant(tenantID)
}

func (c *ServiceContainer) GetTenantPreferenceRepositoryForTenant(tenantID string) (tenantRepos.UserPreferenceRepository, error) {
    return c.repositoryFactory.GetUserPreferenceRepositoryForTenant(tenantID)
}
// Infrastructure holds infrastructure components
type Infrastructure struct {
	MainDB            db.Database
	DBManager         *db.Manager
	Cache             *cache.GenericCache
	CredentialManager *credentials.Manager
	QueueManager      *queue.QueueManager
	WorkerPoolManager *workerpool.WorkerPoolManager
	Logger            logger.Logger
}

// Cleanup gracefully shuts down all services
func (c *ServiceContainer) Cleanup(ctx context.Context) error {
	var errors []error

	// Close infrastructure in reverse order
	if c.workerPoolManager != nil {
		if err := c.workerPoolManager.Shutdown(); err != nil {
			errors = append(errors, err)
		}
	}
	if c.queueManager != nil {
		if err := c.queueManager.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if c.dbManager != nil {
		if err := c.dbManager.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if c.mainDB != nil {
		if err := c.mainDB.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return errors[0] // Return first error
	}

	return nil
}

// HealthCheck verifies all services are healthy
func (c *ServiceContainer) HealthCheck(ctx context.Context) error {
	// Check main database
	if err := c.mainDB.Ping(ctx); err != nil {
		return fmt.Errorf("main database unhealthy: %w", err)
	}

	// Check credential manager
	if !c.credentialManager.IsHealthy() {
		return fmt.Errorf("credential manager unhealthy")
	}

	// Check queue manager
	if !c.queueManager.IsHealthy() {
		return fmt.Errorf("queue manager unhealthy")
	}

	// Check worker pool manager
	if !c.workerPoolManager.IsHealthy() {
		return fmt.Errorf("worker pool manager unhealthy")
	}

	return nil
}

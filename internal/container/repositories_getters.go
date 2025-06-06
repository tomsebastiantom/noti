package container

import (
	"fmt"

	"getnoti.com/pkg/credentials"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/logger"

	notificationRepos "getnoti.com/internal/notifications/repos"
	notificationImpl "getnoti.com/internal/notifications/repos/implementations"
	providerRepos "getnoti.com/internal/providers/repos"
	providerImpl "getnoti.com/internal/providers/repos/implementations"
	schedulerRepos "getnoti.com/internal/shared/scheduler"
	templateRepos "getnoti.com/internal/templates/repos"
	templateImpl "getnoti.com/internal/templates/repos/implementations"
	tenantRepos "getnoti.com/internal/tenants/repos"
	tenantImpl "getnoti.com/internal/tenants/repos/implementations"
	webhookRepos "getnoti.com/internal/webhooks/repos"
	webhookImpl "getnoti.com/internal/webhooks/repos/implementations"
)

// RepositoryFactory creates tenant-specific repositories
type RepositoryFactory struct {
    dbManager         *db.Manager
    credentialManager *credentials.Manager
    logger            logger.Logger
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(
    dbManager *db.Manager,
    credentialManager *credentials.Manager,
    logger logger.Logger,
) *RepositoryFactory {
    return &RepositoryFactory{
        dbManager:         dbManager,
        credentialManager: credentialManager,
        logger:           logger,
    }
}

// GetNotificationRepositoryForTenant creates a notification repository for a tenant
func (f *RepositoryFactory) GetNotificationRepositoryForTenant(tenantID string) (notificationRepos.NotificationRepository, error) {
    // Get tenant DB connection
    db, err := f.dbManager.GetDatabaseConnection(tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to get database for tenant %s: %w", tenantID, err)
    }
    
    return notificationImpl.NewNotificationRepository(db), nil
}

// GetTemplateRepositoryForTenant creates a template repository for a tenant
func (f *RepositoryFactory) GetTemplateRepositoryForTenant(tenantID string) (templateRepos.TemplateRepository, error) {
    // Get tenant DB connection
    db, err := f.dbManager.GetDatabaseConnection(tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to get database for tenant %s: %w", tenantID, err)
    }
    
    return templateImpl.NewTemplateRepository(db), nil
}

// GetProviderRepositoryForTenant creates a provider repository for a tenant
func (f *RepositoryFactory) GetProviderRepositoryForTenant(tenantID string) (providerRepos.ProviderRepository, error) {
    // Get tenant DB connection
    db, err := f.dbManager.GetDatabaseConnection(tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to get database for tenant %s: %w", tenantID, err)
    }
    
    return providerImpl.NewProviderRepository(db), nil
}

// GetWebhookRepositoryForTenant creates a webhook repository for a tenant
func (f *RepositoryFactory) GetWebhookRepositoryForTenant(tenantID string) (webhookRepos.WebhookRepository, error) {
    // Get tenant DB connection
    db, err := f.dbManager.GetDatabaseConnection(tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to get database for tenant %s: %w", tenantID, err)
    }
    
    return webhookImpl.NewWebhookRepository(db), nil
}

// GetSchedulerRepository creates a scheduler repository using the admin database
// Note: Scheduler uses admin database for better scalability across tenants
func (f *RepositoryFactory) GetSchedulerRepository() (schedulerRepos.Repository, error) {
    // Get admin/main DB connection instead of tenant-specific DB
    mainDB := f.dbManager.GetMainDatabase()
    if mainDB == nil {
        return nil, fmt.Errorf("main database is not initialized")
    }
    
    return schedulerRepos.NewSchedulerRepository(mainDB), nil
}

func (f *RepositoryFactory) GetUserPreferenceRepositoryForTenant(tenantID string) (tenantRepos.UserPreferenceRepository, error) {
    
    db, err := f.dbManager.GetDatabaseConnection(tenantID)

    if err != nil {
        return nil, fmt.Errorf("failed to get database for tenant %s: %w", tenantID, err)
    }

    return tenantImpl.NewUserPreferenceRepository(db), nil
}

// GetTenantPreferenceRepositoryForTenant creates a tenant preference repository for a tenant
func (f *RepositoryFactory) GetTenantPreferenceRepositoryForTenant(tenantID string) (tenantRepos.TenantPreferenceRepository, error) {
    db, err := f.dbManager.GetDatabaseConnection(tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to get database for tenant %s: %w", tenantID, err)
    }

    return tenantImpl.NewTenantPreferenceRepository(db), nil
}
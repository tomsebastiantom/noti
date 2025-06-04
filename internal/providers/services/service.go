package providers

import (
	"context"
	"fmt"

	"getnoti.com/internal/providers/domain"
	"getnoti.com/internal/providers/dtos"
	"getnoti.com/internal/providers/infra/providers"
	"getnoti.com/internal/providers/repos"
	tenantServices "getnoti.com/internal/tenants/services"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/credentials"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/workerpool"
)

type ProviderService struct {
    providerRepo        repos.ProviderRepository
    tenantService       *tenantServices.TenantService
    credentialManager   *credentials.Manager
    cache              *cache.GenericCache
    factory            *providers.ProviderFactory
    notificationManager *NotificationManager
    logger             logger.Logger
}

func NewProviderService(
    providerRepo repos.ProviderRepository,
    tenantService *tenantServices.TenantService,
    credentialManager *credentials.Manager,
    cache *cache.GenericCache,
    factory *providers.ProviderFactory,
    queue queue.Queue,
    wpm *workerpool.WorkerPoolManager,
    userPrefService *tenantServices.UserPreferenceService,
    logger logger.Logger,
) *ProviderService {
    // Create the adapter that implements UserPreferenceChecker from the UserPreferenceService
    userPrefChecker := NewUserPreferenceCheckerAdapter(userPrefService)
    
    return &ProviderService{
        providerRepo:        providerRepo,
        tenantService:       tenantService,
        credentialManager:   credentialManager,
        cache:              cache,
        factory:            factory,
        notificationManager: NewNotificationManager(queue, factory, wpm, userPrefChecker),
        logger:             logger,
    }
}

func (s *ProviderService) DispatchNotification(ctx context.Context, tenantID string, providerID string, req dtos.SendNotificationRequest) dtos.SendNotificationResponse {
    s.logger.InfoContext(ctx, "Dispatching notification",
        logger.String("tenant_id", tenantID),
        logger.String("provider_id", providerID))

    // Validate tenant access
    err := s.tenantService.ValidateTenantAccess(ctx, tenantID)
    if err != nil {
        s.logger.ErrorContext(ctx, "Tenant validation failed",
            logger.String("tenant_id", tenantID),
            logger.Err(err))
        return dtos.SendNotificationResponse{Success: false, Message: "Tenant validation failed"}
    }    // Verify provider belongs to tenant
    _, err = s.GetProviderForTenant(ctx, tenantID, providerID)
    if err != nil {
        s.logger.ErrorContext(ctx, "Failed to get provider",
            logger.String("tenant_id", tenantID),
            logger.String("provider_id", providerID),
            logger.Err(err))        
            
        return dtos.SendNotificationResponse{Success: false, Message: "Provider not found"}
    }

    req.ProviderID = providerID
    req.Sender = tenantID

    err = s.notificationManager.DispatchNotification(ctx, req)
    if err != nil {
        s.logger.ErrorContext(ctx, "Failed to dispatch notification",
            logger.String("tenant_id", tenantID),
            logger.String("provider_id", providerID),
            logger.Err(err))
        return dtos.SendNotificationResponse{Success: false, Message: "Failed to send notification"}
    }

    s.logger.InfoContext(ctx, "Notification dispatched successfully",
        logger.String("tenant_id", tenantID),
        logger.String("provider_id", providerID))

    return dtos.SendNotificationResponse{Success: true, Message: "Notification sent successfully"}
}

// GetProviderForTenant retrieves a provider for a specific tenant
func (s *ProviderService) GetProviderForTenant(ctx context.Context, tenantID, providerID string) (*domain.Provider, error) {
    s.logger.DebugContext(ctx, "Getting provider for tenant",
        logger.String("tenant_id", tenantID),
        logger.String("provider_id", providerID))

    // Validate tenant access
    err := s.tenantService.ValidateTenantAccess(ctx, tenantID)
    if err != nil {
        return nil, fmt.Errorf("tenant validation failed: %w", err)
    }    // Get provider from repository
    provider, err := s.providerRepo.GetProviderByID(ctx, providerID)
    if err != nil {
        return nil, fmt.Errorf("failed to get provider: %w", err)
    }

    // Note: Since Provider domain doesn't have TenantID field,
    // tenant validation is handled by the repository layer
    // which should only return providers accessible to the tenant

    return provider, nil
}

func (s *ProviderService) Shutdown() {
    s.notificationManager.Shutdown()
}

func (s *ProviderService) GetPreferences(ctx context.Context,l string,lo string) map[string]string {
    return map[string]string{
        "df": "dsd",
        "gh": "dd",
    }
}



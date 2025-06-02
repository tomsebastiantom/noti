package tenants

import (
	"context"
	"fmt"

	"getnoti.com/internal/tenants/domain"
	repos "getnoti.com/internal/tenants/repos"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/logger"
)

// TenantService handles tenant-related business operations
type TenantService struct {
	tenantRepo        repos.TenantsRepository
	userRepo          repos.UserRepository
	connectionManager *db.Manager
	configResolver    db.ConfigResolver
	logger            logger.Logger
}

// NewTenantService creates a new tenant service
func NewTenantService(
	tenantRepo repos.TenantsRepository,
	userRepo repos.UserRepository,
	connectionManager *db.Manager,
	configResolver db.ConfigResolver,
	logger logger.Logger,
) *TenantService {
	return &TenantService{
		tenantRepo:        tenantRepo,
		userRepo:          userRepo,
		connectionManager: connectionManager,
		configResolver:    configResolver,
		logger:            logger,
	}
}

// GetTenantDatabase returns a database connection for a tenant
func (s *TenantService) GetTenantDatabase(ctx context.Context, tenantID string) (db.Database, error) {
	s.logger.DebugContext(ctx, "Getting tenant database",
		logger.String("tenant_id", tenantID))

	// Get database connection for tenant
	database, err := s.connectionManager.GetConnection(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection for tenant %s: %w", tenantID, err)
	}

	return database, nil
}

// ValidateTenantAccess validates if a tenant exists and is accessible
func (s *TenantService) ValidateTenantAccess(ctx context.Context, tenantID string) error {
	s.logger.DebugContext(ctx, "Validating tenant access",
		logger.String("tenant_id", tenantID))

	// Check if tenant exists
	_, err := s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("tenant validation failed for %s: %w", tenantID, err)
	}

	// Verify database connectivity
	_, err = s.GetTenantDatabase(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("tenant database access failed for %s: %w", tenantID, err)
	}

	s.logger.DebugContext(ctx, "Tenant access validated successfully",
		logger.String("tenant_id", tenantID))

	return nil
}

// GetTenantConfig retrieves database configuration for a tenant
func (s *TenantService) GetTenantConfig(ctx context.Context, tenantID string) (*db.DatabaseConfig, error) {
	s.logger.DebugContext(ctx, "Getting tenant configuration",
		logger.String("tenant_id", tenantID))

	config, err := s.configResolver.ResolveConfig(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration for tenant %s: %w", tenantID, err)
	}

	return config, nil
}

// CreateTenant creates a new tenant with database setup
func (s *TenantService) CreateTenant(ctx context.Context, tenant domain.Tenant) error {
	s.logger.InfoContext(ctx, "Creating new tenant",
		logger.String("tenant_id", tenant.ID),
		logger.String("tenant_name", tenant.Name))

	// Create tenant record
	err := s.tenantRepo.CreateTenant(ctx, tenant)
	if err != nil {
		return fmt.Errorf("failed to create tenant %s: %w", tenant.ID, err)
	}
	// Initialize tenant database if needed
	_, err = s.GetTenantDatabase(ctx, tenant.ID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to initialize tenant database",
			logger.String("tenant_id", tenant.ID),
			logger.String("error", err.Error()))
		// Note: We might want to rollback tenant creation here
		return fmt.Errorf("failed to initialize database for tenant %s: %w", tenant.ID, err)
	}
	s.logger.InfoContext(ctx, "Tenant created successfully",
		logger.String("tenant_id", tenant.ID))

	return nil
}

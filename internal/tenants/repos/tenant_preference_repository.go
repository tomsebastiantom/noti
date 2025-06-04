package repository

import (
	"context"

	"getnoti.com/internal/tenants/domain"
)

type TenantPreferenceRepository interface {
	// Create a new tenant preference
	CreateTenantPreference(ctx context.Context, preference domain.TenantPreference) error
	
	// Get tenant preference by tenant ID
	GetTenantPreferenceByTenantID(ctx context.Context, tenantID string) (domain.TenantPreference, error)
	
	// Update an existing tenant preference
	UpdateTenantPreference(ctx context.Context, preference domain.TenantPreference) error
}

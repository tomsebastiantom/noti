package createtenants

import (
    "context"
    "getnoti.com/internal/tenants/domain"
    repository "getnoti.com/internal/tenants/repos"
)

type CreateTenantsUseCase interface {
    Execute(ctx context.Context, input CreateTenantsRequest) (CreateTenantsResponse, error)
}

type createTenantsUseCase struct {
    repo repository.TenantRepository
}

func NewCreateTenantsUseCase(repo repository.TenantRepository) CreateTenantsUseCase {
    return &createTenantsUseCase{
        repo: repo,
    }
}

func (uc *createTenantsUseCase) Execute(ctx context.Context, input CreateTenantsRequest) (CreateTenantsResponse, error) {
    var successTenants []string
    var failedTenants []FailedTenant

    for _, tenant := range input.Tenants {
        if tenant.ID == "" || tenant.Name == "" {
            failedTenants = append(failedTenants, FailedTenant{ID: tenant.ID, Error: ErrMissingRequiredFields.Error()})
            continue
        }

        domainTenant := domain.Tenant{
            ID:             tenant.ID,
            Name:           tenant.Name,
      
            Preferences:    make(map[string]domain.ChannelPreference),
        }

        for key, pref := range tenant.Preferences {
            domainTenant.Preferences[key] = domain.ChannelPreference{
                ChannelName: domain.NotificationChannel(pref.ChannelName),
                Enabled:     pref.Enabled,
                ProviderID:  pref.ProviderID,
            }
        }

        err := uc.repo.CreateTenant(ctx, domainTenant)
        if err != nil {
            failedTenants = append(failedTenants, FailedTenant{ID: tenant.ID, Error: err.Error()})
        } else {
            successTenants = append(successTenants, tenant.ID)
        }
    }

    if len(successTenants) == 0 {
        return CreateTenantsResponse{}, ErrAllTenantsCreationFailed
    }

    return CreateTenantsResponse{
        SuccessTenants: successTenants,
        FailedTenants:  failedTenants,
    }, nil
}

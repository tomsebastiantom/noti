package tenants

import (
    "context"
    "getnoti.com/internal/tenants/repos"
)

type TenantService struct {
    repo repository.TenantPreferenceRepository
}

func NewTenantService(repo repository.TenantPreferenceRepository) *TenantService {
    return &TenantService{repo: repo}
}

func (s *TenantService) GetPreferences(ctx context.Context, tenantID string, channel string) ((map[string]string), error) {
    return s.repo.GetPreferenceByChannel(ctx, tenantID, channel)
}

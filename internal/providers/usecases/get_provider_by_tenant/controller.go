package getproviderbytenant

import (
    "context"
)

type GetProviderByTenantController struct {
    useCase GetProviderByTenantUseCase
}

func NewGetProviderByTenantController(useCase GetProviderByTenantUseCase) *GetProviderByTenantController {
    return &GetProviderByTenantController{useCase: useCase}
}

func (c *GetProviderByTenantController) GetProviderByTenant(ctx context.Context, req GetProviderByTenantRequest) (GetProviderByTenantResponse, error) {
    return c.useCase.Execute(ctx, req)
}
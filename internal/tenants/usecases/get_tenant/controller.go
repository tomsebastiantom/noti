package gettenant

import (
    "context"
)

type GetTenantController struct {
    useCase GetTenantUseCase
}

func NewGetTenantController(useCase GetTenantUseCase) *GetTenantController {
    return &GetTenantController{useCase: useCase}
}

func (c *GetTenantController) GetTenant(ctx context.Context, req GetTenantRequest) (GetTenantResponse, error) {
    return c.useCase.Execute(ctx, req)
}

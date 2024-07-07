package gettemplates

import (
    "context"
)

type GetTemplatesByTenantController struct {
    useCase GetTemplatesByTenantUseCase
}

func NewGetTemplatesByTenantController(useCase GetTemplatesByTenantUseCase) *GetTemplatesByTenantController {
    return &GetTemplatesByTenantController{useCase: useCase}
}

func (c *GetTemplatesByTenantController) GetTemplatesByTenant(ctx context.Context, req GetTemplatesByTenantRequest) (GetTemplatesByTenantResponse, error) {
    return c.useCase.Execute(ctx, req)
}

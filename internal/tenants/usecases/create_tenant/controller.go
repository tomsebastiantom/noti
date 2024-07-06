package createtenant

import (
    "context"
)

// Controller to handle the use case logic
type CreateTenantController struct {
    useCase CreateTenantUseCase
}

func NewCreateTenantController(useCase CreateTenantUseCase) *CreateTenantController {
    return &CreateTenantController{useCase: useCase}
}

func (c *CreateTenantController) CreateTenant(ctx context.Context, req CreateTenantInput) (CreateTenantOutput, error) {
    return c.useCase.Execute(ctx, req)
}

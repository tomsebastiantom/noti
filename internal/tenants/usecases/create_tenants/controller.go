package createtenants

import (
    "context"
)

type CreateTenantsController struct {
    useCase CreateTenantsUseCase
}

func NewCreateTenantsController(useCase CreateTenantsUseCase) *CreateTenantsController {
    return &CreateTenantsController{useCase: useCase}
}

func (c *CreateTenantsController) CreateTenants(ctx context.Context, req CreateTenantsRequest) (CreateTenantsResponse, error) {
    return c.useCase.Execute(ctx, req)
}

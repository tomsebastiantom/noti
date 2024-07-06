package gettenants

import (
    "context"
)

type GetTenantsController struct {
    useCase GetTenantsUseCase
}

func NewGetTenantsController(useCase GetTenantsUseCase) *GetTenantsController {
    return &GetTenantsController{useCase: useCase}
}

func (c *GetTenantsController) Execute(ctx context.Context, req GetTenantsRequest) (GetTenantsResponse, error) {
    return c.useCase.Execute(ctx, req)
}

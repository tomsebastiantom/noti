package getproviders

import (
    "context"
)

type GetProvidersController struct {
    useCase GetProvidersUseCase
}

func NewGetProvidersController(useCase GetProvidersUseCase) *GetProvidersController {
    return &GetProvidersController{useCase: useCase}
}

func (c *GetProvidersController) GetProviders(ctx context.Context, req GetProvidersRequest) (GetProvidersResponse, error) {
    return c.useCase.Execute(ctx, req)
}
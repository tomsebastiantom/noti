package getprovider

import (
    "context"
)

type GetProviderController struct {
    useCase GetProviderUseCase
}

func NewGetProviderController(useCase GetProviderUseCase) *GetProviderController {
    return &GetProviderController{useCase: useCase}
}

func (c *GetProviderController) GetProvider(ctx context.Context, req GetProviderRequest) (GetProviderResponse, error) {
    return c.useCase.Execute(ctx, req)
}
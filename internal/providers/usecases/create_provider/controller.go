package createprovider

import (
    "context"
)

type CreateProviderController struct {
    useCase CreateProviderUseCase
}

func NewCreateProviderController(useCase CreateProviderUseCase) *CreateProviderController {
    return &CreateProviderController{useCase: useCase}
}

func (c *CreateProviderController) CreateProvider(ctx context.Context, req CreateProviderRequest) (CreateProviderResponse, error) {
    return c.useCase.Execute(ctx, req)
}
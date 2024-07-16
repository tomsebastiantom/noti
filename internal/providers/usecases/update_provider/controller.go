package updateprovider

import (
    "context"
)

type UpdateProviderController struct {
    useCase UpdateProviderUseCase
}

func NewUpdateProviderController(useCase UpdateProviderUseCase) *UpdateProviderController {
    return &UpdateProviderController{useCase: useCase}
}

func (c *UpdateProviderController) UpdateProvider(ctx context.Context, req UpdateProviderRequest) (UpdateProviderResponse, error) {
    return c.useCase.Execute(ctx, req)
}
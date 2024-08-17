
package updatetenant

import (
    "context"
)

type UpdateTenantController struct {
    useCase UpdateTenantUseCase
}

func NewUpdateTenantController(useCase UpdateTenantUseCase) *UpdateTenantController {
    return &UpdateTenantController{useCase: useCase}
}

func (c *UpdateTenantController) UpdateTenant(ctx context.Context, req UpdateTenantRequest) (UpdateTenantResponse,error) {
    return c.useCase.Execute(ctx, req)
}

package updateusers

import (
    "context"
)

type UpdateUsersController struct {
    useCase UpdateUsersUseCase
}

func NewUpdateUsersController(useCase UpdateUsersUseCase) *UpdateUsersController {
    return &UpdateUsersController{useCase: useCase}
}

func (c *UpdateUsersController) UpdateTenant(ctx context.Context, req UpdateUsersRequest) (UpdateUsersResponse, error) {
    return c.useCase.Execute(ctx, req)
}




package updateuser

import (
    "context"
)

type UpdateUserController struct {
    useCase UpdateUserUseCase
}

func NewUpdateUserController(useCase UpdateUserUseCase) *UpdateUserController {
    return &UpdateUserController{useCase: useCase}
}

func (c *UpdateUserController) UpdateUser(ctx context.Context, req UpdateUserRequest) (UpdateUserResponse, error) {
    return c.useCase.Execute(ctx, req)
}




package createuser

import (
    "context"
)

type CreateUserController struct {
    useCase CreateUserUseCase
}

func NewCreateUserController(useCase CreateUserUseCase) *CreateUserController {
    return &CreateUserController{useCase: useCase}
}

func (c *CreateUserController) CreateUser(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error) {
    return c.useCase.Execute(ctx, req)
}

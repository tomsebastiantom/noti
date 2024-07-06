package usecase

import (
    "context"
)

type CreateUserController struct {
    useCase CreateUserUseCase
}

func NewCreateUserController(useCase CreateUserUseCase) *CreateUserController {
    return &CreateUserController{useCase: useCase}
}

func (c *CreateUserController) Handle(ctx context.Context, req CreateUserInput) (CreateUserOutput, error) {
    return c.useCase.Execute(ctx, req)
}

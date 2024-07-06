package createusers

import (
    "context"
)

type CreateUsersController struct {
    useCase CreateUsersUseCase
}

func NewCreateUsersController(useCase CreateUsersUseCase) *CreateUsersController {
    return &CreateUsersController{useCase: useCase}
}

func (c *CreateUsersController) Handle(ctx context.Context, req CreateUsersInput) (CreateUsersOutput, error) {
    return c.useCase.Execute(ctx, req)
}

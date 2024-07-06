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

func (c *CreateUsersController) CreateUsers(ctx context.Context, req CreateUsersRequest) (CreateUsersResponse, error) {
    return c.useCase.Execute(ctx, req)
}

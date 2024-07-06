package getusers

import (
    "context"
)

type GetUsersController struct {
    useCase GetUsersUseCase
}

func NewGetUsersController(useCase GetUsersUseCase) *GetUsersController {
    return &GetUsersController{useCase: useCase}
}

func (c *GetUsersController) GetUsers(ctx context.Context, req GetUsersRequest) (GetUsersResponse, error) {
    return c.useCase.Execute(ctx, req)
}

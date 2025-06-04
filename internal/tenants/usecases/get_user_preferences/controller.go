package getuserpreferences

import (
	"context"
)

type GetUserPreferencesController struct {
	useCase GetUserPreferencesUseCase
}

func NewGetUserPreferencesController(useCase GetUserPreferencesUseCase) *GetUserPreferencesController {
	return &GetUserPreferencesController{useCase: useCase}
}

func (c *GetUserPreferencesController) GetUserPreferences(ctx context.Context, req GetUserPreferencesRequest) (GetUserPreferencesResponse, error) {
	return c.useCase.Execute(ctx, req)
}

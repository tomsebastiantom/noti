package updateuserpreferences

import (
	"context"
)

type UpdateUserPreferencesController struct {
	useCase UpdateUserPreferencesUseCase
}

func NewUpdateUserPreferencesController(useCase UpdateUserPreferencesUseCase) *UpdateUserPreferencesController {
	return &UpdateUserPreferencesController{useCase: useCase}
}

func (c *UpdateUserPreferencesController) UpdateUserPreferences(ctx context.Context, req UpdateUserPreferencesRequest) (UpdateUserPreferencesResponse, error) {
	return c.useCase.Execute(ctx, req)
}

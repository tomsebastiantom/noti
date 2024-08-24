package gettemplates

import (
    "context"
)

type GetTemplatesController struct {
    useCase GetTemplatesUseCase
}

func NewGetTemplatesController(useCase GetTemplatesUseCase) *GetTemplatesController {
    return &GetTemplatesController{useCase: useCase}
}

func (c *GetTemplatesController) GetTemplates(ctx context.Context, req GetTemplatesRequest) (GetTemplatesResponse, error) {
    return c.useCase.Execute(ctx, req)
}

package gettemplate

import (
    "context"
)

type GetTemplateController struct {
    useCase GetTemplateUseCase
}

func NewGetTemplateController(useCase GetTemplateUseCase) *GetTemplateController {
    return &GetTemplateController{useCase: useCase}
}

func (c *GetTemplateController) GetTemplate(ctx context.Context, req GetTemplateRequest) GetTemplateResponse {
    return c.useCase.Execute(ctx, req)
}

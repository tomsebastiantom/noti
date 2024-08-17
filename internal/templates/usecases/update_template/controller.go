package updatetemplate

import (
    "context"
)

type UpdateTemplateController struct {
    useCase UpdateTemplateUseCase
}

func NewUpdateTemplateController(useCase UpdateTemplateUseCase) *UpdateTemplateController {
    return &UpdateTemplateController{useCase: useCase}
}

func (c *UpdateTemplateController) UpdateTemplate(ctx context.Context, req UpdateTemplateRequest) (UpdateTemplateResponse,error) {
    return c.useCase.Execute(ctx, req)
}

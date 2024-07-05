package createtemplate

import (
    "context"

    
)

type CreateTemplateController struct {
    useCase CreateTemplateUseCase
}

func NewCreateTemplateController(useCase CreateTemplateUseCase) *CreateTemplateController {
    return &CreateTemplateController{useCase: useCase}
}

func (c *CreateTemplateController) CreateTemplate(ctx context.Context, req CreateTemplateRequest) CreateTemplateResponse {
    return c.useCase.Execute(ctx, req)
}

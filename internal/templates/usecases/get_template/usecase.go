package gettemplate

import (
    "context"
    "getnoti.com/internal/templates/repos"
)

type GetTemplateUseCase interface {
    Execute(ctx context.Context, req GetTemplateRequest) GetTemplateResponse
}

type getTemplateUseCase struct {
    repository repos.TemplateRepository
}

func NewGetTemplateUseCase(repository repos.TemplateRepository) GetTemplateUseCase {
    return &getTemplateUseCase{repository: repository}
}

func (uc *getTemplateUseCase) Execute(ctx context.Context, req GetTemplateRequest) GetTemplateResponse {
    tmpl, err := uc.repository.GetTemplateByID(ctx, req.TemplateID)
    if err != nil {
        return GetTemplateResponse{Success: false, Message: err.Error()}
    }

    return GetTemplateResponse{
        Template: *tmpl,
        Success:  true,
        Message:  "Template retrieved successfully",
    }
}
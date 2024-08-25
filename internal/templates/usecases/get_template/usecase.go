package gettemplate

import (
    "context"
    "getnoti.com/internal/templates/repos"
)

type GetTemplateUseCase interface {
    Execute(ctx context.Context, req GetTemplateRequest) (GetTemplateResponse,error)
}

type getTemplateUseCase struct {
    repository repos.TemplateRepository
}

func NewGetTemplateUseCase(repository repos.TemplateRepository) GetTemplateUseCase {
    return &getTemplateUseCase{repository: repository}
}

func (uc *getTemplateUseCase) Execute(ctx context.Context, req GetTemplateRequest) (GetTemplateResponse,error) {
    tmpl, err := uc.repository.GetTemplateByID(ctx, req.ID)
    if err != nil {
        return GetTemplateResponse{Success: false, Message: err.Error()},err
    }

    return GetTemplateResponse{
        Template: *tmpl,
        Success:  true,
        Message:  "Template retrieved successfully",
    },nil
}
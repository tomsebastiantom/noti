package createtemplate

import (
    "context"

   "getnoti.com/internal/shared/utils"
    "getnoti.com/internal/templates/domain/template"
    "getnoti.com/internal/templates/repos"
)

type CreateTemplateUseCase interface {
    Execute(ctx context.Context, req CreateTemplateRequest) CreateTemplateResponse
}

type createTemplateUseCase struct {
    repository repos.TemplateRepository
}

func NewCreateTemplateUseCase(repository repos.TemplateRepository) CreateTemplateUseCase {
    return &createTemplateUseCase{repository: repository}
}

func (uc *createTemplateUseCase) Execute(ctx context.Context, req CreateTemplateRequest) CreateTemplateResponse {
    tmplID, err := utils.GenerateUUID()
    if err != nil {
        return CreateTemplateResponse{Success: false, Message: ErrUnexpected.Error()}
    }

    tmpl := &template.Template{
        ID:        tmplID,
        TenantID:  req.TenantID,
        Name:      req.Name,
        Content:   req.Content,
        IsPublic:  req.IsPublic,
        Variables: req.Variables,
    }

    err = uc.repository.CreateTemplate(ctx, tmpl)
    if err != nil {
        return CreateTemplateResponse{Success: false, Message: ErrTemplateCreationFailed.Error()}
    }

    return CreateTemplateResponse{Success: true, Message: "Template created successfully"}
}
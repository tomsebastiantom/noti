package gettemplates

import (
    "context"
    "getnoti.com/internal/templates/repos"
)

type GetTemplatesByTenantUseCase interface {
    Execute(ctx context.Context, req GetTemplatesByTenantRequest) (GetTemplatesByTenantResponse, error)
}

type getTemplatesByTenantUseCase struct {
    repo repos.TemplateRepository
}

func NewGetTemplatesByTenantUseCase(repo repos.TemplateRepository) GetTemplatesByTenantUseCase {
    return &getTemplatesByTenantUseCase{
        repo: repo,
    }
}

func (uc *getTemplatesByTenantUseCase) Execute(ctx context.Context, req GetTemplatesByTenantRequest) (GetTemplatesByTenantResponse, error) {
    templates, err := uc.repo.GetTemplatesByTenantID(ctx, req.TenantID)
    if err != nil {
        return GetTemplatesByTenantResponse{}, err
    }
    return GetTemplatesByTenantResponse{Templates: templates}, nil
}

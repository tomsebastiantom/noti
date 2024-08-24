package gettemplates

import (
    "context"
    "getnoti.com/internal/templates/repos"
)

type GetTemplatesUseCase interface {
    Execute(ctx context.Context, req GetTemplatesRequest) (GetTemplatesResponse, error)
}

type getTemplatesUseCase struct {
    repo repos.TemplateRepository
}

func NewGetTemplatesUseCase(repo repos.TemplateRepository) GetTemplatesUseCase {
    return &getTemplatesUseCase{
        repo: repo,
    }
}

func (uc *getTemplatesUseCase) Execute(ctx context.Context, req GetTemplatesRequest) (GetTemplatesResponse, error) {
    templates, err := uc.repo.GetTemplates(ctx)
    if err != nil {
        return GetTemplatesResponse{}, err
    }
    return GetTemplatesResponse{Templates: templates}, nil
}

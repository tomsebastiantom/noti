package repos

import (
    "context"
    "getnoti.com/internal/templates/domain"
)

type TemplateRepository interface {
    CreateTemplate(ctx context.Context, tmpl *domain.Template) error
    GetTemplateByID(ctx context.Context, templateID string) (*domain.Template, error)
    UpdateTemplate(ctx context.Context, tmpl *domain.Template) error
    GetTemplatesByTenantID(ctx context.Context, templateID string)([]domain.Template, error)
}

package repos

import (
    "context"
    "getnoti.com/internal/templates/domain/template"
)

type TemplateRepository interface {
    CreateTemplate(ctx context.Context, tmpl *template.Template) error
    GetTemplateByID(ctx context.Context, templateID string) (*template.Template, error)
    UpdateTemplate(ctx context.Context, tmpl *template.Template) error
}

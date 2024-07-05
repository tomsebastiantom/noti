package updatetemplate

import (
	"getnoti.com/internal/templates/domain/template"
)

type UpdateTemplateRequest struct {
    TemplateID string
    TenantID   *string
    Name       *string
    Content    *string
    IsPublic   *bool
    Variables  *[]string
}

type UpdateTemplateResponse struct {
    Template template.Template
    Success  bool
    Message  string
}
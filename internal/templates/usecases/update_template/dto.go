package updatetemplate

import (
	"getnoti.com/internal/templates/domain"
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
    Template domain.Template
    Success  bool
    Message  string
}
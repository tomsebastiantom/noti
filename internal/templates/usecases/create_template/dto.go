package createtemplate

import (
    "getnoti.com/internal/templates/domain"
)

type CreateTemplateRequest struct {
    Name      string
    Content   string
    IsPublic  bool
    Variables []string
}



type CreateTemplateResponse struct {
    Success bool
    Template domain.Template
    Message string
}
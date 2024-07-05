package gettemplate

import (

    "getnoti.com/internal/templates/domain/template"

)

type GetTemplateRequest struct {
    TemplateID string
}

type GetTemplateResponse struct {
    Template template.Template
    Success  bool
    Message  string
}
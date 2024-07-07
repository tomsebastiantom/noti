package gettemplate

import (

    "getnoti.com/internal/templates/domain"

)

type GetTemplateRequest struct {
    TemplateID string
}

type GetTemplateResponse struct {
    Template domain.Template
    Success  bool
    Message  string
}
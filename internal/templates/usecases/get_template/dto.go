package gettemplate

import (
	"getnoti.com/internal/templates/domain"
)

type GetTemplateRequest struct {
	ID string
}

func (r *GetTemplateRequest) SetID(id string) {
	r.ID = id
}

type GetTemplateResponse struct {
	Template domain.Template
	Success  bool
	Message  string
}

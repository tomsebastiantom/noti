package updatetemplate

import (
	"getnoti.com/internal/templates/domain"
)

type UpdateTemplateRequest struct {
	ID        string
	Name      *string
	Content   *string
	IsPublic  *bool
	Variables *[]string
}

func (r *UpdateTemplateRequest) SetID(id string) {
	r.ID = id
}

type UpdateTemplateResponse struct {
	Template domain.Template
	Success  bool
	Message  string
}

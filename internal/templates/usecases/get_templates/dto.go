package gettemplates

import (
	"getnoti.com/internal/templates/domain"
)

type GetTemplatesRequest struct {
	TenantID string
}

func (r *GetTemplatesRequest) SetTenantID(id string) {
	r.TenantID = id
}

type GetTemplatesResponse struct {
	Templates []domain.Template `json:"templates"`
}

package gettemplates

import (
    "getnoti.com/internal/templates/domain"
)

type GetTemplatesByTenantRequest struct {
    TenantID string `json:"tenant_id"`
}

type GetTemplatesByTenantResponse struct {
    Templates []domain.Template `json:"templates"`
}

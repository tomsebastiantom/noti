package getusers

import (
    "getnoti.com/internal/tenants/domain"
)

type GetUsersRequest struct {
    UserID   string
    TenantID string
}

func (r *GetUsersRequest) SetTenantID(id string) {
    r.TenantID = id
}
type GetUsersResponse struct {
    Users []domain.User
}

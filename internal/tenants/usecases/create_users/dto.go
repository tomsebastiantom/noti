package createusers

import (

    "getnoti.com/internal/tenants/domain"
   
)
type CreateUsersRequest struct {
    TenantID string
    Users []domain.User
}

type CreateUsersResponse struct {
    SuccessUsers []domain.User
    FailedUsers  []FailedUser
}

func (r *CreateUsersRequest) SetTenantID(id string) {
    r.TenantID = id
}

type FailedUser struct {
    UserID string
    Reason string
}

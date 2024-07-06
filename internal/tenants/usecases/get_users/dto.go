package get_users

import (
    "getnoti.com/internal/tenants/domain"
)

type GetUsersRequest struct {
    UserID   string
    TenantID string
}

type GetUsersResponse struct {
    Users []domain.User
}

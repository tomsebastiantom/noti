package repository

import (
    "context"
    "getnoti.com/internal/tenants/domain"
)

type UserRepository interface {
    CreateUser(ctx context.Context, user domain.User) error
    GetUserByID(ctx context.Context, userid string) (user domain.User, error error)
    UpdateUser(ctx context.Context, user domain.User) error
    GetUsersByTenantID(ctx context.Context, tenantid string) (users []domain.User, error error)
}

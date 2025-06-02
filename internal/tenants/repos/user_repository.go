package repository

import (
	"context"

	"getnoti.com/internal/tenants/domain"
)

type UserRepository interface {
    CreateUser(ctx context.Context, user domain.User) error
    GetUserByID(ctx context.Context, userid string) (domain.User, error)
    UpdateUser(ctx context.Context, user domain.User) error
    GetUsers(ctx context.Context) ([]domain.User, error)
}

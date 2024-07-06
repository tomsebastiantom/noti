package get_users

import (
    "context"
    "errors"
    "getnoti.com/internal/tenants/domain"
    repository "getnoti.com/internal/tenants/repos"
)

type GetUsersUseCase interface {
    Execute(ctx context.Context, input GetUsersRequest) (GetUsersResponse, error)
}

type getUsersUseCase struct {
    repo repository.UserRepository

}

func NewGetUsersUseCase(repo repository.UserRepository) GetUsersUseCase {
    return &getUsersUseCase{
        repo: repo,
    }
}

func (uc *getUsersUseCase) Execute(ctx context.Context, req GetUsersRequest) (GetUsersResponse, error) {
    if req.UserID != "" {
        user, err := uc.repo.GetUserByID(ctx, req.UserID)
        if err != nil {
            return GetUsersResponse{}, err
        }
        return GetUsersResponse{Users: []domain.User{user}}, nil
    } else if req.TenantID != "" {
        users, err := uc.repo.GetUsersByTenantID(ctx, req.TenantID)
        if err != nil {
            return GetUsersResponse{}, err
        }
        return GetUsersResponse{Users: users}, nil
    }
    return GetUsersResponse{}, errors.New("invalid request")
}

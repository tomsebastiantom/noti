package usecase

import (
    "context"
    "getnoti.com/internal/tenants/domain"
    repository "getnoti.com/internal/tenants/repos"
)

type CreateUserUseCase interface {
    Execute(ctx context.Context, input CreateUserInput) (CreateUserOutput, error)
}


type createUserUseCase struct {
    repo repository.UserRepository
}

func NewCreateUserUseCase(repo repository.UserRepository) CreateUserUseCase {
    return &createUserUseCase{
        repo: repo,
    }
}

func (uc *createUserUseCase) Execute(ctx context.Context, input CreateUserInput) (CreateUserOutput, error) {
    user := domain.User{
        ID:            input.ID,
        TenantID:      input.TenantID,
        Email:         input.Email,
        PhoneNumber:   input.PhoneNumber,
        DeviceID:      input.DeviceID,
        WebPushToken:  input.WebPushToken,
        Consents:      input.Consents,
        PreferredMode: input.PreferredMode,
    }

    err := uc.repo.CreateUser(ctx, user)
    if err != nil {
        return CreateUserOutput{}, err
    }

    return CreateUserOutput{User: user}, nil
}

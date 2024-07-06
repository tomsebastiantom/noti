package updateuser

import (
    "context"

    "getnoti.com/internal/tenants/domain"
    repository "getnoti.com/internal/tenants/repos"
)

type UpdateUserUseCase interface {
    Execute(ctx context.Context, input UpdateUserRequest) (UpdateUserResponse, error)
}



type updateUserUseCase struct {
    repo repository.UserRepository
}

func NewUpdateUserUseCase(repo repository.UserRepository) UpdateUserUseCase {
    return &updateUserUseCase{
        repo: repo,
    }
}

func (uc *updateUserUseCase) Execute(ctx context.Context, input UpdateUserRequest) (UpdateUserResponse, error) {
    user := domain.User{
        ID:            input.UserID,
        TenantID:      input.TenantID,
        Email:         input.Email,
        PhoneNumber:   input.PhoneNumber,
        DeviceID:      input.DeviceID,
        WebPushToken:  input.WebPushToken,
        Consents:      input.Consents,
        PreferredMode: input.PreferredMode,
    }

    err := uc.repo.UpdateUser(ctx, user)
    if err != nil {
        return UpdateUserResponse{Success: false, Message: err.Error()}, err
    }

    return UpdateUserResponse{Success: true, Message: "User updated successfully"}, nil
}

package updateusers

import (
    "context"

    "getnoti.com/internal/tenants/domain"
    repository "getnoti.com/internal/tenants/repos"
)

type UpdateUsersUseCase interface {
    Execute(ctx context.Context, input UpdateUsersRequest) (UpdateUsersResponse, error)
}



type updateUsersUseCase struct {
    repo repository.UserRepository
}

func NewUpdateUsersUseCase(repo repository.UserRepository) UpdateUsersUseCase {
    return &updateUsersUseCase{
        repo: repo,
    }
}

func (uc *updateUsersUseCase) Execute(ctx context.Context, input UpdateUsersRequest) (UpdateUsersResponse, error) {
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
        return UpdateUsersResponse{Success: false, Message: err.Error()}, err
    }

    return UpdateUsersResponse{Success: true, Message: "User updated successfully"}, nil
}

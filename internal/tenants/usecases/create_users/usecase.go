package createusers

import (
    "context"
    repository"getnoti.com/internal/tenants/repos"
)

type CreateUsersUseCase interface {
    Execute(ctx context.Context, input CreateUsersRequest) (CreateUsersResponse, error)
}





type createUsersUseCase struct {
    repo repository.UserRepository
}

func NewCreateUsersUseCase(repo repository.UserRepository) CreateUsersUseCase {
    return &createUsersUseCase{
        repo: repo,
    }
}

func (uc *createUsersUseCase) Execute(ctx context.Context, input CreateUsersRequest) (CreateUsersResponse, error) {
    var output CreateUsersResponse
    for _, user := range input.Users {
        if user.ID == ""  {
            output.FailedUsers = append(output.FailedUsers, FailedUser{
                UserID: user.ID,
                Reason: "Missing user ID",
            })
            continue
        }
        err := uc.repo.CreateUser(ctx, user)
        if err != nil {
            output.FailedUsers = append(output.FailedUsers, FailedUser{
                UserID: user.ID,
                Reason: err.Error(),
            })
        } else {
            output.SuccessUsers = append(output.SuccessUsers, user)
        }
    }
    return output, nil
}

package createuser

import (
	"context"
	"getnoti.com/internal/shared/utils"
	"getnoti.com/internal/tenants/domain"
	repository "getnoti.com/internal/tenants/repos"
)

type CreateUserUseCase interface {
	Execute(ctx context.Context, input CreateUserRequest) (CreateUserResponse, error)
}

type createUserUseCase struct {
	repo repository.UserRepository
}

func NewCreateUserUseCase(repo repository.UserRepository) CreateUserUseCase {
	return &createUserUseCase{
		repo: repo,
	}
}

func (uc *createUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error) {

	ID := utils.GenerateUUID()

	user := domain.User{
		ID:            ID,
		Email:         req.Email,
		PhoneNumber:   req.PhoneNumber,
		DeviceID:      req.DeviceID,
		WebPushToken:  req.WebPushToken,
		Consents:      req.Consents,
		PreferredMode: req.PreferredMode,
	}

	err := uc.repo.CreateUser(ctx, user)
	if err != nil {
		return CreateUserResponse{}, err
	}

	return CreateUserResponse{User: user}, nil
}

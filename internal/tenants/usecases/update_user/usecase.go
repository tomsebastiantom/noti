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
   
    existingUser, err := uc.repo.GetUserByID(ctx, input.ID) 
    if err != nil {
        return UpdateUserResponse{Success: false, Message: err.Error()}, err
    }


    updatedUser := domain.User{
        ID:            existingUser.ID, // Ensure we keep the original ID
        Email:         ifNotEmpty(input.Email, existingUser.Email),
        PhoneNumber:   ifNotEmpty(input.PhoneNumber, existingUser.PhoneNumber),
        DeviceID:      ifNotEmpty(input.DeviceID, existingUser.DeviceID),
    }


   

    // Update the user in the repository
    err = uc.repo.UpdateUser(ctx, updatedUser)
    if err != nil {
        return UpdateUserResponse{Success: false, Message: err.Error()}, err
    }

    return UpdateUserResponse{Success: true, Message: "User updated successfully"}, nil
}

// Helper function to use the new value if it's not empty, otherwise use the current value
func ifNotEmpty(new, current string) string {
    if new != "" {
        return new
    }
    return current
}

package createusers

import (

    "getnoti.com/internal/tenants/domain"
   
)
type CreateUsersInput struct {
    Users []domain.User
}

type CreateUsersOutput struct {
    SuccessUsers []domain.User
    FailedUsers  []FailedUser
}



type FailedUser struct {
    UserID string
    Reason string
}

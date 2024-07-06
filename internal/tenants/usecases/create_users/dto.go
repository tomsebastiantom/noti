package createusers

import (

    "getnoti.com/internal/tenants/domain"
   
)
type CreateUsersRequest struct {
    Users []domain.User
}

type CreateUsersResponse struct {
    SuccessUsers []domain.User
    FailedUsers  []FailedUser
}



type FailedUser struct {
    UserID string
    Reason string
}

package updateuser

import "errors"

var (
    ErrUserNotFound = errors.New("user not found")
    ErrUpdateFailed = errors.New("failed to update user")
)

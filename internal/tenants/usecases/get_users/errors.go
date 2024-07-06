package get_users

import "errors"

var (
    ErrUserNotFound = errors.New("user not found")
    ErrTenantNotFound = errors.New("tenant not found")
)

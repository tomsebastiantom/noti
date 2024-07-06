package getusers

import "errors"

var (
    ErrUserNotFound = errors.New("user not found")
    ErrTenantNotFound = errors.New("tenant not found")
)

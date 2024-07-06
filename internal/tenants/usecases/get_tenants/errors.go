package gettenants

import "errors"

var (
    ErrTenantNotFound = errors.New("tenant not found")
    ErrInternal       = errors.New("internal error")
)

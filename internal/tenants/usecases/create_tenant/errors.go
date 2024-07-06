package createtenant

import "errors"

// Custom error definitions
var (
    ErrTenantAlreadyExists = errors.New("tenant already exists")
    ErrInvalidTenantData   = errors.New("invalid tenant data")
)

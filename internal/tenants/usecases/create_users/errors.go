package createusers

import "errors"

var (
    ErrMissingUserID    = errors.New("missing user ID")
    ErrMissingTenantID  = errors.New("missing tenant ID")
    ErrUserCreationFail = errors.New("user creation failed")
)

package getuserpreferences

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrTenantNotFound = errors.New("tenant not found")
	ErrPreferencesNotFound = errors.New("user preferences not found")
)

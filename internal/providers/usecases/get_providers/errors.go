package getproviders

import "errors"

var (
    ErrInvalidTenantID = errors.New("invalid tenant ID")
    ErrNoProvidersFound = errors.New("no providers found for the given tenant ID")
)
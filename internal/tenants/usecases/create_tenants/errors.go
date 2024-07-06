package createtenants

import "errors"

var (
    ErrAllTenantsCreationFailed = errors.New("all tenants creation failed")
    ErrMissingRequiredFields    = errors.New("ID and Name are required fields")
)

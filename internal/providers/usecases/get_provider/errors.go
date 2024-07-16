package getprovider

import "errors"

var (
    ErrProviderNotFound = errors.New("provider not found")
    ErrInvalidProviderID = errors.New("invalid provider ID")
)
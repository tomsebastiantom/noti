package createprovider

import "errors"

var (
    ErrInvalidProviderName     = errors.New("invalid provider name")
    ErrInvalidChannels         = errors.New("invalid channels")
    ErrProviderAlreadyExists   = errors.New("provider already exists")
    ErrFailedToCreateProvider  = errors.New("failed to create provider")
)
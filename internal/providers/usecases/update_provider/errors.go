package updateprovider

import "errors"

var (
    ErrProviderNotFound       = errors.New("provider not found")
    ErrInvalidProviderName    = errors.New("invalid provider name")
    ErrInvalidChannels        = errors.New("invalid channels")
    ErrFailedToUpdateProvider = errors.New("failed to update provider")
)
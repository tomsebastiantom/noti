package vault

import "errors"

var (
    ErrVaultNotInitialized     = errors.New("vault provider not initialized")
    ErrInvalidCredentialFormat = errors.New("invalid credential data format")
    ErrCredentialNotFound      = errors.New("credential not found")
    ErrVaultConnectionFailed   = errors.New("vault connection failed")
    ErrInvalidVaultConfig      = errors.New("invalid vault configuration")
)
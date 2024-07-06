
package updatetenant

import (
    "errors"
)

var (
    ErrTenantNotFound = errors.New("tenant not found")
    ErrUpdateFailed   = errors.New("failed to update tenant")
)

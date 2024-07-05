
package sendnotification

import "errors"

var (
    ErrNotificationCreationFailed = errors.New("notification creation failed")
    ErrNotificationNotFound       = errors.New("notification not found")
    ErrUnexpected                 = errors.New("unexpected error occurred")
)

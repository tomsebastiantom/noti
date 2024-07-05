package gettemplate

import "errors"

var (
    ErrTemplateNotFound = errors.New("template not found")
    ErrUnexpected       = errors.New("unexpected error occurred")
)

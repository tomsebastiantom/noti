package updatetemplate

import "errors"

var (
    ErrTemplateUpdateFailed = errors.New("template update failed")
    ErrTemplateNotFound     = errors.New("template not found")
    ErrUnexpected           = errors.New("unexpected error occurred")
)
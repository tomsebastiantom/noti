package postgres

import "errors"

var (
    ErrTemplateNotFound = errors.New("template not found")
    ErrTemplateCreateFailed = errors.New("template creation failed")
    ErrTemplateUpdateFailed = errors.New("template update failed")
    ErrUnexpected = errors.New("unexpected error occurred")
)

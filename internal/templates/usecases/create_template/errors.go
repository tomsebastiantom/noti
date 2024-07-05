package createtemplate

import "errors"

var (
    ErrTemplateCreationFailed = errors.New("template creation failed")
    ErrUnexpected             = errors.New("unexpected error occurred")
)

package gettemplates

import "errors"

var (
    ErrTemplateNotFound = errors.New("template not found")
    ErrInvalidTenantID  = errors.New("invalid tenant ID")
)

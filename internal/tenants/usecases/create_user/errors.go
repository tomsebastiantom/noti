package usecase

import "errors"

var (
    ErrInvalidInput = errors.New("invalid input")
    ErrUserExists   = errors.New("user already exists")
    ErrCreateFailed = errors.New("failed to create user")
)

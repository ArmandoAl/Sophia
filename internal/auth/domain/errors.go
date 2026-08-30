package domain

import "errors"

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInternalServerError = errors.New("internal server error")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidName         = errors.New("name is required")
	ErrInvalidEmail        = errors.New("email is required")
	ErrWeakPassword        = errors.New("password must be at least 8 characters")
)

package user

import "errors"

var (
	ErrNotFound      = errors.New("user not found")
	ErrEmailExists   = errors.New("email already exists")
	ErrInvalidInput  = errors.New("invalid input")
)

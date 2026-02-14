package service

import "errors"

var (
	ErrCarNotFound = errors.New("car not found")
	ErrValidation  = errors.New("validation failed")
)

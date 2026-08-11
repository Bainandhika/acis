package errors

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrForbidden         = errors.New("forbidden operation")
	ErrBadRequest        = errors.New("invalid request parameters")
	ErrConflict          = errors.New("resource conflict")
	ErrInternalServer    = errors.New("internal server error")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

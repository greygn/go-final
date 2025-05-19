package common

import "errors"

var (
	// ErrNotFound is returned when a requested resource is not found
	ErrNotFound = errors.New("not found")

	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidToken is returned when a JWT token is invalid
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired is returned when a JWT token has expired
	ErrTokenExpired = errors.New("token expired")

	// ErrUnauthorized is returned when a user is not authorized to perform an action
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrDuplicateKey is returned when a unique constraint is violated
	ErrDuplicateKey = errors.New("duplicate key")

	// ErrInternalServer is returned when an internal server error occurs
	ErrInternalServer = errors.New("internal server error")
)

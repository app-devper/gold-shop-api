package entity

import "errors"

// Common domain errors
var (
	ErrNotFound            = errors.New("entity not found")
	ErrDuplicateKey        = errors.New("duplicate key error")
	ErrInvalidInput        = errors.New("invalid input")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidStatus       = errors.New("invalid status for this operation")
	ErrExpired             = errors.New("item has expired")
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrInsufficientPoints  = errors.New("insufficient points")
)

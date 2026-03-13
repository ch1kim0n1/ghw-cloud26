package services

import (
	"errors"
	"net/http"
)

var ErrPlaceholderClient = errors.New("provider client not implemented in phase 0")

type AppError struct {
	Status  int
	Code    string
	Message string
	Details map[string]any
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(status int, code, message string, details map[string]any, err error) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
		Err:     err,
	}
}

func InvalidRequest(code, message string, details map[string]any) *AppError {
	return NewAppError(http.StatusBadRequest, code, message, details, nil)
}

func ResourceNotFound(message string, details map[string]any, err error) *AppError {
	return NewAppError(http.StatusNotFound, "RESOURCE_NOT_FOUND", message, details, err)
}

func StorageFailure(message string, details map[string]any, err error) *AppError {
	return NewAppError(http.StatusInternalServerError, "STORAGE_ERROR", message, details, err)
}

func DatabaseFailure(message string, details map[string]any, err error) *AppError {
	return NewAppError(http.StatusInternalServerError, "DATABASE_ERROR", message, details, err)
}

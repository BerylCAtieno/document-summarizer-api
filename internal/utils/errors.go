package utils

import "net/http"

type AppError struct {
	StatusCode int
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewBadRequestError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

func NewNotFoundError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Message:    message,
	}
}

func NewInternalError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		Message:    message,
	}
}

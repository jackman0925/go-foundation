package errors

import stderrors "errors"

const (
	// CodeOK represents a successful operation.
	CodeOK = 0
	// CodeBadRequest represents invalid input.
	CodeBadRequest = 40000
	// CodeUnauthorized represents missing or invalid authentication.
	CodeUnauthorized = 40100
	// CodeForbidden represents a forbidden operation.
	CodeForbidden = 40300
	// CodeNotFound represents a missing resource.
	CodeNotFound = 40400
	// CodeInternal represents an internal server error.
	CodeInternal = 50000
)

// AppError carries a stable code and public message while optionally wrapping a cause.
type AppError struct {
	Code    int
	Message string
	cause   error
}

// Error returns the public error message.
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap returns the original cause for errors.Is and errors.As.
func (e *AppError) Unwrap() error {
	return e.cause
}

// New creates an AppError without an underlying cause.
func New(code int, message string) error {
	return &AppError{Code: code, Message: message}
}

// Wrap creates an AppError and preserves the underlying cause.
func Wrap(code int, message string, cause error) error {
	return &AppError{Code: code, Message: message, cause: cause}
}

// CodeOf returns an AppError code or CodeInternal for unknown errors.
func CodeOf(err error) int {
	if err == nil {
		return CodeOK
	}

	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr.Code
	}
	return CodeInternal
}

// MessageOf returns an AppError message or the error string for unknown errors.
func MessageOf(err error) string {
	if err == nil {
		return "ok"
	}

	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr.Message
	}
	return err.Error()
}

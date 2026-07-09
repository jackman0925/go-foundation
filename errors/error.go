package errors

import stderrors "errors"

const (
	// CodeOK 表示操作成功。
	CodeOK = 0
	// CodeBadRequest 表示请求参数无效。
	CodeBadRequest = 40000
	// CodeUnauthorized 表示缺少认证或认证无效。
	CodeUnauthorized = 40100
	// CodeForbidden 表示操作被禁止。
	CodeForbidden = 40300
	// CodeNotFound 表示资源不存在。
	CodeNotFound = 40400
	// CodeInternal 表示内部服务错误。
	CodeInternal = 50000
)

// AppError 携带稳定错误码和公开错误信息，并可包装原始错误。
type AppError struct {
	Code    int
	Message string
	cause   error
}

// Error 返回公开错误信息。
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap 返回原始错误，用于 errors.Is 和 errors.As。
func (e *AppError) Unwrap() error {
	return e.cause
}

// New 创建不包含原始错误的 AppError。
func New(code int, message string) error {
	return &AppError{Code: code, Message: message}
}

// Wrap 创建 AppError 并保留原始错误。
func Wrap(code int, message string, cause error) error {
	return &AppError{Code: code, Message: message, cause: cause}
}

// CodeOf 返回 AppError 错误码；未知错误返回 CodeInternal。
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

// MessageOf 返回 AppError 错误信息；未知错误返回 error 字符串。
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

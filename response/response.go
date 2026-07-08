package response

import foundationErrors "github.com/jackman0925/go-foundation/errors"

// Response is the standard API response shape.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PageData is the standard paginated response payload.
type PageData struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	PageNo   int   `json:"pageNo"`
	PageSize int   `json:"pageSize"`
}

// NewOK creates a successful response.
func NewOK(data any) Response {
	return Response{
		Code:    foundationErrors.CodeOK,
		Message: "ok",
		Data:    data,
	}
}

// NewFail creates a failed response from an error.
func NewFail(err error) Response {
	return Response{
		Code:    foundationErrors.CodeOf(err),
		Message: foundationErrors.MessageOf(err),
	}
}

// NewPage creates a successful paginated response.
func NewPage(list any, total int64, page int, pageSize int) Response {
	return NewOK(PageData{
		List:     list,
		Total:    total,
		PageNo:   page,
		PageSize: pageSize,
	})
}

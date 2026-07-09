package response

import foundationErrors "github.com/jackman0925/go-foundation/errors"

// Response 是标准 API 响应结构。
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PageData 是标准分页响应数据结构。
type PageData struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	PageNo   int   `json:"pageNo"`
	PageSize int   `json:"pageSize"`
}

// NewOK 创建成功响应。
func NewOK(data any) Response {
	return Response{
		Code:    foundationErrors.CodeOK,
		Message: "ok",
		Data:    data,
	}
}

// NewFail 根据 error 创建失败响应。
func NewFail(err error) Response {
	return Response{
		Code:    foundationErrors.CodeOf(err),
		Message: foundationErrors.MessageOf(err),
	}
}

// NewPage 创建成功分页响应。
func NewPage(list any, total int64, page int, pageSize int) Response {
	return NewOK(PageData{
		List:     list,
		Total:    total,
		PageNo:   page,
		PageSize: pageSize,
	})
}

// internal/model/response.go
package model

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PageInfo 分页信息
type PageInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	List     interface{} `json:"list"`
	PageInfo PageInfo    `json:"page_info"`
}

// 响应码常量
const (
	CodeSuccess         = 0
	CodeError           = 1
	CodeInvalidParams   = 400
	CodeUnauthorized    = 401
	CodeForbidden       = 403
	CodeNotFound        = 404
	CodeTooManyRequests = 429
	CodeInternalError   = 500
	CodeServiceUnavailable = 503
)

// 响应消息常量
const (
	MsgSuccess             = "success"
	MsgError               = "error"
	MsgInvalidParams       = "invalid parameters"
	MsgUnauthorized        = "unauthorized"
	MsgForbidden           = "forbidden"
	MsgNotFound            = "resource not found"
	MsgTooManyRequests     = "too many requests"
	MsgInternalError       = "internal server error"
	MsgServiceUnavailable  = "service unavailable"
)

// Success 成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code:    CodeSuccess,
		Message: MsgSuccess,
		Data:    data,
	}
}

// Error 错误响应
func Error(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewPaginatedResponse 创建分页响应
func NewPaginatedResponse(list interface{}, page, pageSize int, total int64) *PaginatedResponse {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginatedResponse{
		List: list,
		PageInfo: PageInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

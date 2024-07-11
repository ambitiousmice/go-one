package cust_error

import "fmt"

// CustomError 是自定义的错误类型
type CustomError struct {
	ErrorCode    int32  // 错误码
	ErrorMessage string // 错误信息
}

// 实现 error 接口的 Error 方法
func (e *CustomError) Error() string {
	return fmt.Sprintf("Error: %d - %s", e.ErrorCode, e.ErrorMessage)
}

// NewCustomError 创建一个新的 CustomError
func NewCustomError(code int32, message string) error {
	return &CustomError{
		ErrorCode:    code,
		ErrorMessage: message,
	}
}

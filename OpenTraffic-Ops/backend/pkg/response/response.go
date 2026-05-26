package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CODE_TAG = "code"
	MSG_TAG  = "msg"
	DATA_TAG = "data"
)

// HttpStatus HTTP状态码常量
type HttpStatus int

const (
	SUCCESS        HttpStatus = 200
	CREATED        HttpStatus = 201
	ACCEPTED       HttpStatus = 202
	NO_CONTENT     HttpStatus = 204
	MOVED_PERM     HttpStatus = 301
	BAD_REQUEST    HttpStatus = 400
	UNAUTHORIZED   HttpStatus = 401
	FORBIDDEN      HttpStatus = 403
	NOT_FOUND      HttpStatus = 404
	BAD_METHOD     HttpStatus = 405
	CONFLICT       HttpStatus = 409
	ERROR          HttpStatus = 500
	BAD_GATEWAY    HttpStatus = 502
	SERVICE_UNAV   HttpStatus = 503
	WARN           HttpStatus = 601
)

// Result 统一响应结构
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Result{
		Code: int(SUCCESS),
		Msg:  "操作成功",
		Data: data,
	})
}

// SuccessWithMsg 返回成功响应（自定义消息）
func SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Result{
		Code: int(SUCCESS),
		Msg:  msg,
		Data: data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Result{
		Code: int(ERROR),
		Msg:  msg,
	})
}

// ErrorWithCode 返回错误响应（自定义状态码）
func ErrorWithCode(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Result{
		Code: code,
		Msg:  msg,
	})
}

// Warn 返回警告响应
func Warn(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Result{
		Code: int(WARN),
		Msg:  msg,
	})
}

// Unauthorized 返回未授权响应
func Unauthorized(c *gin.Context, msg string) {
	if msg == "" {
		msg = "登录状态已过期"
	}
	c.JSON(http.StatusOK, Result{
		Code: int(UNAUTHORIZED),
		Msg:  msg,
	})
}

// Forbidden 返回禁止访问响应
func Forbidden(c *gin.Context, msg string) {
	if msg == "" {
		msg = "没有权限，请联系管理员授权"
	}
	c.JSON(http.StatusOK, Result{
		Code: int(FORBIDDEN),
		Msg:  msg,
	})
}

// PageResult 分页响应结构
// 注意：total/rows 与 code/msg 平铺在同一层（参照若依规范），
// 前端通过 res.total / res.rows 直接取值，不再嵌套到 data 字段。
type PageResult struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Total int64       `json:"total"`
	Rows  interface{} `json:"rows"`
}

// SuccessPage 返回分页成功响应
func SuccessPage(c *gin.Context, total int64, rows interface{}) {
	if rows == nil {
		rows = []interface{}{}
	}
	c.JSON(http.StatusOK, PageResult{
		Code:  int(SUCCESS),
		Msg:   "查询成功",
		Total: total,
		Rows:  rows,
	})
}

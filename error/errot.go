/**
 * @Author $
 * @Description //TODO $
 * @Date $ $
 * @Param $
 * @return $
 **/

package eeor

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Error struct {
	StatusCode int    `json:"-"`
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
}

var (
	Success     = NewError(http.StatusOK, 0, "success")
	ServerError = NewError(http.StatusInternalServerError, 200500, "System exception. Please try again later")
	NotFound    = NewError(http.StatusNotFound, 200404, http.StatusText(http.StatusNotFound))
)

func OtherError(message string) *Error {
	return NewError(http.StatusForbidden, 100403, message)
}
func (e *Error) Error() string {
	return e.Msg
}

func NewError(statusCode, Code int, msg string) *Error {
	return &Error{
		StatusCode: statusCode,
		Code:       Code,
		Msg:        msg,
	}
}

// HandleNotFound 404处理
func HandleNotFound(c *gin.Context) {
	err := NotFound
	c.JSON(err.StatusCode, err)
	return
}

// ErrHandler /** 全局异常处理
func ErrHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var Err *Error

				if e, ok := err.(*Error); ok {
					Err = e
				} else if e, ok := err.(error); ok {
					Err = OtherError(e.Error())
				} else {
					Err = ServerError
				}
				// 记录一个错误的日志
				dataType, _ := json.Marshal(c.Request.PostForm)
				dataString := string(dataType)
				start := time.Now()
				path := c.Request.URL.Path
				query := c.Request.URL.RawQuery
				cost := time.Since(start)
				zap.L().Info(path,
					zap.Int("status", c.Writer.Status()),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("query", query),
					zap.String("ip", c.ClientIP()),
					zap.String("client-agent", c.Request.UserAgent()),
					zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
					zap.Duration("cost", cost),
					zap.String("PostForm", dataString),
				)

				c.Next()
				c.JSON(Err.StatusCode, Err)
				return
			}
		}()
		c.Next()
	}
}

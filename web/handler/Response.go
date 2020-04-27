package handler

import "github.com/gin-gonic/gin"

const (
	FAIL    = 0
	SUCCESS = 1
)

func NewResponse(code int, message string, data interface{}) gin.H {
	return gin.H{"code": code, "message": message, "data": data}
}

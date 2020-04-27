package web

import (
	"github.com/gin-gonic/gin"
	"github.com/sduwh/vcode-judger/web/handler"
)

func GetRouter() *gin.Engine{
	router := gin.Default()
	handler.TestHandlerRoutes(router)
	return router
}
package router

import (
	// "net/http"

	"github.com/gin-gonic/gin"
)

func RouterInit() *gin.Engine {
	ginRouter := gin.Default()
	ginRouter.LoadHTMLGlob("HTML/*")

	ginRouter.GET("/", StartHandler)
	ginRouter.POST("/login", LoginHandler)

	return ginRouter
}
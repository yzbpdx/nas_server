package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RouterInit() *gin.Engine {
	ginRouter := gin.Default()
	ginRouter.LoadHTMLGlob("HTML/*")

	ginRouter.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Hello, Gin!",
		})
	})

	return ginRouter
}
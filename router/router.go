package router

import (
	// "net/http"

	"net/http"

	"github.com/gin-gonic/gin"
)

func RouterInit() *gin.Engine {
	ginRouter := gin.Default()
	ginRouter.LoadHTMLGlob("HTML/*")

	ginRouter.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/login")
	})
	ginRouter.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "login.html", gin.H{})
	})
	ginRouter.GET("/home", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "home.html", gin.H{})
	})
	ginRouter.GET("/root", RootFolderHandler)

	ginRouter.POST("/login", LoginHandler)
	ginRouter.POST("/folder", ClickFolderHandler)

	return ginRouter
}

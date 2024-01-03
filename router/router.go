package router

import (
	// "net/http"

	"net/http"

	"github.com/gin-gonic/gin"
)

func RouterInit(folderName string) *gin.Engine {
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
	ginRouter.GET("/register", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "register.html", gin.H{})
	})
	ginRouter.GET("/root", func(ctx *gin.Context) {
		RootFolderHandler(ctx, folderName)
	})

	ginRouter.POST("/login", LoginHandler)
	ginRouter.POST("/folder", ClickFolderHandler)
	ginRouter.POST("/download", DownloadHandler)
	ginRouter.POST("/upload", UploadHandler)
	ginRouter.POST("/create", CreateFolderHandler)
	ginRouter.POST("/register", RegisterHandler)
	ginRouter.POST("/fileInfo", FileInfoHandler)

	return ginRouter
}

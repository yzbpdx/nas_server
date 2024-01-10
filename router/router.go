package router

import (
	// "net/http"

	"net/http"
	"path/filepath"

	"nas_server/conf"
	"nas_server/logs"

	"github.com/gin-gonic/gin"
)

func RouterInit(serverConfig *config.ServerConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	ginRouter := gin.Default()
	ginRouter.LoadHTMLGlob("html/*")

	ginRouter.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/login")
	})
	ginRouter.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "login.html", gin.H{})
	})
	ginRouter.GET("/home/:username", func(ctx *gin.Context) {
		userName := ctx.Param("username")
		logs.GetInstance().Logger.Infof("username is %s", userName)
		ctx.HTML(http.StatusOK, "home.html", gin.H{})
	})
	ginRouter.GET("/register", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "register.html", gin.H{})
	})
	ginRouter.GET(serverConfig.HomeUrl + "/root", func(ctx *gin.Context) {
		userName := ctx.Param("username")
		folderName := filepath.Join(serverConfig.RootFolder, userName)
		RootFolderHandler(ctx, folderName)
	})

	ginRouter.POST("/login", LoginHandler)
	ginRouter.POST(serverConfig.HomeUrl + "/folder", ClickFolderHandler)
	ginRouter.POST(serverConfig.HomeUrl + "/download", DownloadHandler)
	ginRouter.POST(serverConfig.HomeUrl + "/upload", UploadHandler)
	ginRouter.POST(serverConfig.HomeUrl + "/create", CreateFolderHandler)
	ginRouter.POST("/register", RegisterHandler)
	ginRouter.POST("/fileInfo", FileInfoHandler)
	ginRouter.POST(serverConfig.HomeUrl + "/rename", RenameHandler)
	ginRouter.POST(serverConfig.HomeUrl + "/delete", DeleteHandler)

	return ginRouter
}

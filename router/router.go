package router

import (
	// "net/http"

	"net/http"
	"path/filepath"

	"nas_server/conf"
	"nas_server/logs"

	"github.com/gin-gonic/gin"
)

func RouterInit(serverConfig *config.ServerConfig, dockerRegistryPost string) *gin.Engine {
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
		var folderName string
		if userName == "share" {
			folderName = serverConfig.ShareFolder
		} else {
			folderName = filepath.Join(serverConfig.RootFolder, userName)
		}
		RootFolderHandler(ctx, folderName)
	})
	ginRouter.GET(serverConfig.HomeUrl + "/download/info", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "trans.html", gin.H{})
	})
	ginRouter.GET(serverConfig.HomeUrl + "/download/info/list", DownloadProgressHandler)
	ginRouter.GET(serverConfig.HomeUrl + "/domain", DomainHandler)
	ginRouter.GET(serverConfig.HomeUrl + "/docker", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "docker.html", gin.H{})
	})
	ginRouter.GET(serverConfig.HomeUrl + "/docker/list", func(ctx *gin.Context) {
		DockerListHandler(ctx, dockerRegistryPost)
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
	ginRouter.POST(serverConfig.HomeUrl + "/download/pause", PauseDownloadHandler)
	ginRouter.POST(serverConfig.HomeUrl + "/download/resume", ResumeDownloadHandler)
	ginRouter.POST(serverConfig.HomeUrl + "/download/cancel", CancelDownloadHandler)

	return ginRouter
}

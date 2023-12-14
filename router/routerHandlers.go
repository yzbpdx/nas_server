package router

import (
	"nas_server/logs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RouterHandlers struct {

}

func StartHandler(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{})
}

func LoginHandler(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	logs.GetInstance().Infof("username: %s, password: %s", username, password)

	if username == "" || password == "" {
		logs.GetInstance().Errorf("username or password is empty")
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "用户名或密码不能为空"})
		return
	} else if username == "dyf" {
		ctx.HTML(http.StatusOK, "home.html", gin.H{"username": username})
	}
}
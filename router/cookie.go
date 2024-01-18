package router

import (
	"nas_server/logs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CookieVerify(ctx *gin.Context) bool {
	userName, err := ctx.Cookie("username")
	if err != nil || userName == "" {
		ctx.Redirect(http.StatusFound, "/login")
		logs.GetInstance().Logger.Warnf("user %s cookie false: %s", userName, err)
		return false
	}
	return true
}
package router

import (
	"nas_server/logs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type RouterHandlers struct {
}

func LoginHandler(ctx *gin.Context) {
	var loginform LoginForm
	if err := ctx.ShouldBindJSON(&loginform); err != nil {
		logs.GetInstance().Warnf("cannot bing loginForm json")
	}
	logs.GetInstance().Infof("username: %s, password: %s", loginform.Username, loginform.PassWord)

	if loginform.Username == "" || loginform.PassWord == "" {
		logs.GetInstance().Errorf("username or password is empty")
		ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username or password is empty"})
		return
	} else if loginform.Username == "dyf" {
		logs.GetInstance().Infof("login success with %s", loginform.Username)
		ctx.JSON(http.StatusOK, gin.H{})
	}
}

func RootFolderHandler(ctx *gin.Context) {
	respFolders := make([]string, 0)
	respFiles := make([]string, 0)
	folderName := "."

	if err := getFiles(folderName, &respFolders, &respFiles); err != nil {
		logs.GetInstance().Errorf("read root dir error: %s", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"folders":     respFolders,
		"files":       respFiles,
		"currentPath": folderName + "/",
	})
}

func ClickFolderHandler(ctx *gin.Context) {
	var folderName RequestFolder
	if err := ctx.ShouldBindJSON(&folderName); err != nil {
		logs.GetInstance().Warnf("cannot bing folderName json")
	}
	logs.GetInstance().Infof("folderName: %s", folderName.FolderName)

	respFolders := make([]string, 0)
	respFiles := make([]string, 0)
	if err := getFiles(folderName.FolderName, &respFolders, &respFiles); err != nil {
		logs.GetInstance().Errorf("read %s error: %s", folderName.FolderName, err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"folders":     respFolders,
		"files":       respFiles,
		"currentPath": folderName.FolderName + "/",
	})
}

func getFiles(folderName string, respFolders, respFiles *[]string) error {
	files, err := os.ReadDir(folderName)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			*respFolders = append(*respFolders, file.Name())
		} else {
			*respFiles = append(*respFiles, file.Name())
		}
	}

	return nil
}

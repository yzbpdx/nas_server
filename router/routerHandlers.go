package router

import (
	"io"
	"nas_server/logs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

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
	logs.GetInstance().Infof("get root dir success!")

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

func DownloadHandler(ctx *gin.Context) {
	var downloadForm DownloadForm
	if err := ctx.ShouldBindJSON(&downloadForm); err != nil {
		logs.GetInstance().Warnf("cannot bind downloadForm json")
	}
	logs.GetInstance().Infof("downloadForm: %+v", downloadForm)

	ctx.Header("Content-Disposition", "attachment; filename="+downloadForm.FileName)
    ctx.Header("Content-Type", "application/octet-stream")
	ctx.File(downloadForm.FilePath + downloadForm.FileName)
}

func UploadHandler(ctx *gin.Context) {
	var uploadForm UploadForm
	if err := ctx.ShouldBind(&uploadForm); err != nil {
		logs.GetInstance().Errorf("cannot bind uploadForm: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind uploadForm"})
		return
	} else {
		logs.GetInstance().Infof("uploadFolder is %s, fileName is %s", uploadForm.UploadFolder, uploadForm.FileName)
	}

	file, err := uploadForm.File.Open()
	if err != nil {
		logs.GetInstance().Errorf("file open error: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file open error"})
		return
	}
	defer file.Close()

	newFile, err := os.Create(uploadForm.UploadFolder + uploadForm.FileName)
	if err != nil {
		logs.GetInstance().Errorf("new file create error: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "new file create error"})
		return
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, file)
	if err != nil {
		logs.GetInstance().Errorf("copy file error: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "copy file error"})
		return
	}

	logs.GetInstance().Infof("upload file success!")
	ctx.JSON(http.StatusOK, gin.H{})
}

func CreateFolderHandler(ctx *gin.Context) {
	var createFolder CreateFolder
	if err := ctx.ShouldBindJSON(&createFolder); err != nil {
		logs.GetInstance().Errorf("cannot bind createFolder: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind createFolder"})
	}
	logs.GetInstance().Infof("%+v", createFolder)

	if err := os.Mkdir(createFolder.CurrentPath + createFolder.FolderName, os.ModePerm); err != nil {
		logs.GetInstance().Errorf("create folder error: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "create folder error"})
	}

	ctx.JSON(http.StatusOK, gin.H{})
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

package router

import (
	"context"
	"io"
	"nas_server/gorm"
	"nas_server/logs"
	"nas_server/redis"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func LoginHandler(ctx *gin.Context) {
	var loginForm UserForm
	if err := ctx.ShouldBindJSON(&loginForm); err != nil {
		logs.GetInstance().Warnf("cannot bind loginForm json")
	}
	logs.GetInstance().Infof("login with username: %s, password: %s", loginForm.Username, loginForm.PassWord)

	if loginForm.Username == "" || loginForm.PassWord == "" {
		logs.GetInstance().Errorf("username or password is empty")
		ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username or password is empty"})
		return
	} else {
		client := redis.GetClient()
		passwordInDB, err := client.Get(context.Background(), loginForm.Username).Result()
		if redis.CheckNil(err) {
			logs.GetInstance().Warnf("%s not exists", loginForm.Username)
			ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username not found"})
			return
		} else if err != nil {
			logs.GetInstance().Errorf("redis error %s", err)
			ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "redis error"})
			return
		}
		if passwordInDB == loginForm.PassWord {
			logs.GetInstance().Infof("login success with %s", loginForm.Username)
			ctx.JSON(http.StatusOK, gin.H{})
		} else {
			logs.GetInstance().Warnf("password incorrect with %s", loginForm.Username)
			ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username dismatch password"})
			return
		}
	}
}

func RegisterHandler(ctx *gin.Context)  {
	var registerForm UserForm
	if err := ctx.ShouldBindJSON(&registerForm); err != nil {
		logs.GetInstance().Warnf("cannot bind userForm json")
	}
	logs.GetInstance().Infof("register with username: %s, password: %s", registerForm.Username, registerForm.PassWord)
	
	if registerForm.Username == "" || registerForm.PassWord == "" {
		logs.GetInstance().Errorf("username or password is empty")
		ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username or password is empty"})
		return
	} else {
		redisClient := redis.GetClient()
		mysqlClinet := gorm.GetClient()
		if exists, err := redisClient.Exists(context.Background(), registerForm.Username).Result(); err != nil {
			logs.GetInstance().Errorf("redis error %s", err)
			ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "server error"})
			return
		} else {
			if exists == 1 {
				logs.GetInstance().Infof("register %s exists", registerForm.Username)
				ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username exists"})
				return
			} else {
				var count int
				err := mysqlClinet.QueryRow("select COUNT(*) FROM user WHERE username = ?", registerForm.Username).Scan(&count)
				if err != nil {
					logs.GetInstance().Errorf("mysql error %s", err)
					ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "server error"})
					return
				}
				if count > 0 {
					logs.GetInstance().Infof("register %s exists", registerForm.Username)
					ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username exists"})
					return
				}
			}
		}
		_, err := redisClient.Set(context.Background(), registerForm.Username, registerForm.PassWord, 24 * time.Hour).Result()
		if err != nil {
			logs.GetInstance().Errorf("redis err %s", err)
			ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "server error"})
			return
		}
		go func() {
			_, err := mysqlClinet.Exec("INSERT INTO user (username, password) VALUES (?, ?)", registerForm.Username, registerForm.PassWord)
			if err != nil {
				logs.GetInstance().Errorf("mysql error %s", err)
			}
		}()
		logs.GetInstance().Infof("register success with %s", registerForm.Username)
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

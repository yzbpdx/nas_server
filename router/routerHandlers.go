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
		logs.GetInstance().Logger.Warnf("cannot bind loginForm json")
	}
	logs.GetInstance().Logger.Infof("login with username: %s, password: %s", loginForm.Username, loginForm.PassWord)

	if loginForm.Username == "" || loginForm.PassWord == "" {
		logs.GetInstance().Logger.Errorf("username or password is empty")
		ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username or password is empty"})
		return
	} else {
		client := redis.GetClient()
		passwordInDB, err := client.Get(context.Background(), loginForm.Username).Result()
		if redis.CheckNil(err) {
			logs.GetInstance().Logger.Warnf("%s not exists", loginForm.Username)
			ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username not found"})
			return
		} else if err != nil {
			logs.GetInstance().Logger.Errorf("redis error %s", err)
			ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "redis error"})
			return
		}
		if passwordInDB == loginForm.PassWord {
			logs.GetInstance().Logger.Infof("login success with %s", loginForm.Username)
			ctx.JSON(http.StatusOK, gin.H{})
		} else {
			logs.GetInstance().Logger.Warnf("password incorrect with %s", loginForm.Username)
			ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username dismatch password"})
			return
		}
	}
}

func RegisterHandler(ctx *gin.Context)  {
	var registerForm UserForm
	if err := ctx.ShouldBindJSON(&registerForm); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind userForm json")
	}
	logs.GetInstance().Logger.Infof("register with username: %s, password: %s", registerForm.Username, registerForm.PassWord)
	
	if registerForm.Username == "" || registerForm.PassWord == "" {
		logs.GetInstance().Logger.Errorf("username or password is empty")
		ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username or password is empty"})
		return
	} else {
		redisClient := redis.GetClient()
		mysqlClinet := gorm.GetClient()
		if exists, err := redisClient.Exists(context.Background(), registerForm.Username).Result(); err != nil {
			logs.GetInstance().Logger.Errorf("redis error %s", err)
			ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "server error"})
			return
		} else {
			if exists == 1 {
				logs.GetInstance().Logger.Infof("register %s exists", registerForm.Username)
				ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username exists"})
				return
			} else {
				var count int
				err := mysqlClinet.QueryRow("select COUNT(*) FROM user WHERE username = ?", registerForm.Username).Scan(&count)
				if err != nil {
					logs.GetInstance().Logger.Errorf("mysql error %s", err)
					ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "server error"})
					return
				}
				if count > 0 {
					logs.GetInstance().Logger.Infof("register %s exists", registerForm.Username)
					ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "username exists"})
					return
				}
			}
		}
		_, err := redisClient.Set(context.Background(), registerForm.Username, registerForm.PassWord, 24 * time.Hour).Result()
		if err != nil {
			logs.GetInstance().Logger.Errorf("redis err %s", err)
			ctx.JSON(http.StatusBadRequest, ErrorResp{ErrorMsg: "server error"})
			return
		}
		go func() {
			_, err := mysqlClinet.Exec("INSERT INTO user (username, password) VALUES (?, ?)", registerForm.Username, registerForm.PassWord)
			if err != nil {
				logs.GetInstance().Logger.Errorf("mysql error %s", err)
			}
		}()
		logs.GetInstance().Logger.Infof("register success with %s", registerForm.Username)
		ctx.JSON(http.StatusOK, gin.H{})
	}
}

func RootFolderHandler(ctx *gin.Context, folderName string) {
	respFolders := make([]string, 0)
	respFiles := make([]string, 0)

	if err := getFiles(folderName, &respFolders, &respFiles); err != nil {
		logs.GetInstance().Logger.Errorf("read root dir error: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "read root dir error"})
		return
	}
	logs.GetInstance().Logger.Infof("get root dir success!")

	ctx.JSON(http.StatusOK, gin.H{
		"folders":     respFolders,
		"files":       respFiles,
		"currentPath": folderName + "/",
	})
}

func ClickFolderHandler(ctx *gin.Context) {
	var folderName RequestFolder
	if err := ctx.ShouldBindJSON(&folderName); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bing folderName json")
	}
	logs.GetInstance().Logger.Infof("folderName: %s", folderName.FolderName)

	respFolders := make([]string, 0)
	respFiles := make([]string, 0)
	if err := getFiles(folderName.FolderName, &respFolders, &respFiles); err != nil {
		logs.GetInstance().Logger.Errorf("read %s error: %s", folderName.FolderName, err)
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
		logs.GetInstance().Logger.Warnf("cannot bind downloadForm json")
	}
	logs.GetInstance().Logger.Infof("downloadForm: %+v", downloadForm)

	ctx.Header("Content-Disposition", "attachment; filename="+downloadForm.FileName)
    ctx.Header("Content-Type", "application/octet-stream")
	ctx.File(downloadForm.FilePath + downloadForm.FileName)
}

func UploadHandler(ctx *gin.Context) {
	var uploadForm UploadForm
	if err := ctx.ShouldBind(&uploadForm); err != nil {
		logs.GetInstance().Logger.Errorf("cannot bind uploadForm: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind uploadForm"})
		return
	} else {
		logs.GetInstance().Logger.Infof("uploadFolder is %s, fileName is %s", uploadForm.UploadFolder, uploadForm.FileName)
	}

	file, err := uploadForm.File.Open()
	if err != nil {
		logs.GetInstance().Logger.Errorf("file open error: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file open error"})
		return
	}
	defer file.Close()

	newFile, err := os.Create(uploadForm.UploadFolder + uploadForm.FileName)
	if err != nil {
		logs.GetInstance().Logger.Errorf("new file create error: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "new file create error"})
		return
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, file)
	if err != nil {
		logs.GetInstance().Logger.Errorf("copy file error: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "copy file error"})
		return
	}

	logs.GetInstance().Logger.Infof("upload file success!")
	ctx.JSON(http.StatusOK, gin.H{})
}

func CreateFolderHandler(ctx *gin.Context) {
	var createFolder CreateFolder
	if err := ctx.ShouldBindJSON(&createFolder); err != nil {
		logs.GetInstance().Logger.Errorf("cannot bind createFolder: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind createFolder"})
	}
	logs.GetInstance().Logger.Infof("%+v", createFolder)

	if err := os.Mkdir(createFolder.CurrentPath + createFolder.FolderName, os.ModePerm); err != nil {
		logs.GetInstance().Logger.Errorf("create folder error: %s", err)
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

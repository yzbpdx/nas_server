package router

import (
	"context"
	"fmt"
	"io"
	"nas_server/gorm"
	"nas_server/logs"
	"nas_server/redis"
	"net/http"
	"os"
	"strconv"
	"strings"
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
		redisClient := redis.GetClient()
		mysqlClinet := gorm.GetClient()
		passwordInDB, err := redisClient.Get(context.Background(), loginForm.Username).Result()
		if redis.CheckNil(err) {
			err = mysqlClinet.QueryRow("SELECT password FROM user WHERE username = ?", loginForm.Username).Scan(&passwordInDB)
			if err != nil {
				logs.GetInstance().Logger.Warnf("user %s not register", loginForm.Username)
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "you need register first"})
				return
			}
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

func FileInfoHandler(ctx *gin.Context) {
	var downloadForm DownloadForm
	if err := ctx.ShouldBindJSON(&downloadForm); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind downloadForm json")
	}
	logs.GetInstance().Logger.Infof("downloadForm: %+v", downloadForm)
	file, err := os.Open(downloadForm.FilePath + downloadForm.FileName)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "server open file error"})
		logs.GetInstance().Logger.Errorf("file open error %s", err)
		return
	}
	defer file.Close()
	stat, _ := file.Stat()
	ctx.Writer.Header().Set("fileSize", strconv.FormatInt(stat.Size(), 10))
	// ctx.JSON(http.StatusOK, gin.H{"fileSize": strconv.FormatInt(stat.Size(), 10)})
}

func DownloadHandler(ctx *gin.Context) {
	var downloadForm DownloadForm
	if err := ctx.ShouldBindJSON(&downloadForm); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind downloadForm json")
	}
	logs.GetInstance().Logger.Infof("downloadForm: %+v", downloadForm)

	// ctx.Header("Content-Disposition", "attachment; filename="+downloadForm.FileName)
    // ctx.Header("Content-Type", "application/octet-stream")
	// ctx.File(downloadForm.FilePath + downloadForm.FileName)
	file, err := os.Open(downloadForm.FilePath + downloadForm.FileName)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "server open file error"})
		logs.GetInstance().Logger.Errorf("file open error %s", err)
		return
	}
	defer file.Close()
	stat, _ := file.Stat()
	ctx.Writer.Header().Set("Content-Disposition", "attachment; filename="+downloadForm.FileName)
	ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
	ctx.Writer.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	ctx.Writer.Flush()
	var offset, bufSize int64 = 0, 1024 * 1024
	buf := make([]byte, bufSize)
	rangeHeader := ctx.Request.Header.Get("Range")
	if rangeHeader != "" {
		parts := strings.Split(rangeHeader, "=")
		if len(parts) == 2 && parts[0] == "bytes" {
			rangeStr := parts[1]
			ranges := strings.Split(rangeStr, "-")
			if len(ranges) == 2 {
				offset, _ = strconv.ParseInt(ranges[0], 10, 64)
				if offset >= stat.Size() {
					ctx.JSON(http.StatusRequestedRangeNotSatisfiable, gin.H{"error": "request range error"})
					logs.GetInstance().Logger.Errorf("request range error. request: %d, file size: %d", offset, stat.Size())
					return
				}
				if ranges[1] != "" {
					endOffset, _ := strconv.ParseInt(ranges[1], 10, 64)
					if endOffset >= stat.Size() {
						endOffset = stat.Size() - 1
					}
					ctx.Writer.Header().Set("Content-Range", "bytes="+ranges[0]+"-"+strconv.FormatInt(endOffset, 10)+"/"+strconv.FormatInt(stat.Size(), 10))
					ctx.Writer.Header().Set("Content-Length", strconv.FormatInt(endOffset-offset+1, 10))
					file.Seek(offset, 0)
				} else {
					ctx.Writer.Header().Set("Content-Range", "bytes="+ranges[0]+"-"+strconv.FormatInt(stat.Size()-1, 10)+"/"+strconv.FormatInt(stat.Size(), 10))
					ctx.Writer.Header().Set("Content-Length", strconv.FormatInt(stat.Size()-offset, 10))
					file.Seek(offset, 0)
				}
				ctx.Writer.WriteHeader(http.StatusPartialContent)
			}
		}
	}
	for {
		n, err := file.ReadAt(buf, offset)
		if err != nil && err != io.EOF {
			logs.GetInstance().Logger.Errorf("read file error %s", err)
			break
		}
		if n == 0 {
			break
		}
		_, err = ctx.Writer.Write(buf[:n])
		if err != nil {
			logs.GetInstance().Logger.Errorf("write file error %s", err)
			break
		}
		offset += int64(n)
		fmt.Println(n, offset)
	}
	ctx.Writer.Flush()
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

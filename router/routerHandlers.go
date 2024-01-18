package router

import (
	"context"
	_ "fmt"
	"io"
	"nas_server/gorm"
	"nas_server/logs"
	"nas_server/redis"
	"net/http"
	"os"
	"path/filepath"
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
	if !CookieVerify(ctx) {
		return
	}
	if _, err := os.Stat(folderName); err != nil && os.IsNotExist(err) {
		os.Mkdir(folderName, 0777)
	}
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
		"currentPath": folderName,
	})
}

func ClickFolderHandler(ctx *gin.Context) {
	if !CookieVerify(ctx) {
		return
	}
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
		"currentPath": folderName.FolderName,
	})
}

func FileInfoHandler(ctx *gin.Context) {
	if !CookieVerify(ctx) {
		return
	}
	var downloadForm DownloadInfo
	if err := ctx.ShouldBindJSON(&downloadForm); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind downloadForm json")
	}
	// logs.GetInstance().Logger.Infof("downloadForm: %+v", downloadForm)
	filePath := filepath.Join(downloadForm.FilePath, downloadForm.FileName)
	file, err := os.Open(filePath)
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

func DownloadHandlerV1(ctx *gin.Context) {
	var downloadInfo DownloadInfo
	if err := ctx.ShouldBindJSON(&downloadInfo); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind downloadForm json")
	}
	// logs.GetInstance().Logger.Infof("downloadForm: %+v", downloadInfo)

	// ctx.Header("Content-Disposition", "attachment; filename="+downloadForm.FileName)
    // ctx.Header("Content-Type", "application/octet-stream")
	// ctx.File(downloadForm.FilePath + downloadForm.FileName)
	filePath := filepath.Join(downloadInfo.FilePath, downloadInfo.FileName)
	file, err := os.Open(filePath)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "server open file error"})
		logs.GetInstance().Logger.Errorf("file open error %s", err)
		return
	}
	defer file.Close()
	stat, _ := file.Stat()
	ctx.Writer.Header().Set("Content-Disposition", "attachment; filename="+downloadInfo.FileName)
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
	}
	ctx.Writer.Flush()
}

func UploadHandler(ctx *gin.Context) {
	if !CookieVerify(ctx) {
		return
	}
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

	filePath := filepath.Join(uploadForm.UploadFolder, uploadForm.FileName)
	newFile, err := os.Create(filePath)
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
	if !CookieVerify(ctx) {
		return
	}
	var createFolder CreateFolder
	if err := ctx.ShouldBindJSON(&createFolder); err != nil {
		logs.GetInstance().Logger.Errorf("cannot bind createFolder: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind createFolder"})
	}
	logs.GetInstance().Logger.Infof("%+v", createFolder)

	folderPath := filepath.Join(createFolder.CurrentPath, createFolder.FolderName)
	if err := os.Mkdir(folderPath, os.ModePerm); err != nil {
		logs.GetInstance().Logger.Errorf("create folder error: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "create folder error"})
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

func RenameHandler(ctx *gin.Context) {
	if !CookieVerify(ctx) {
		return
	}
	var rename Rename
	if err := ctx.ShouldBindJSON(&rename); err != nil {
		logs.GetInstance().Logger.Errorf("cannot bind rename: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind rename"})
		return
	}
	// logs.GetInstance().Logger.Infof("old name: %s, new name: %s", rename.OldName, rename.NewName)

	rename.OldName = filepath.Join(rename.CurrentPath, rename.OldName)
	rename.NewName = filepath.Join(rename.CurrentPath, rename.NewName)
	logs.GetInstance().Logger.Infof("change %s to %s", rename.OldName, rename.NewName)

	if err := os.Rename(rename.OldName, rename.NewName); err != nil {
		logs.GetInstance().Logger.Errorf("os rename error: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "os rename error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

func DeleteHandler(ctx *gin.Context) {
	if !CookieVerify(ctx) {
		return
	}
	var delete Delete
	if err := ctx.ShouldBindJSON(&delete); err != nil {
		logs.GetInstance().Logger.Errorf("cannot bind delete: %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot bind delete"})
		return
	}

	deletePath := filepath.Join(delete.CurrentPath, delete.DeleteName)
	err := os.RemoveAll(deletePath)
	if err != nil {
		logs.GetInstance().Logger.Errorf("delete %s error: %s", deletePath, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "delete error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

func getFiles(folderName string, respFolders, respFiles *[]string) error {
	files, err := os.ReadDir(folderName)
	if err != nil {
		return err
	}

	for _, file := range files {
		if configFile(file.Name()) {
			continue
		}
		if file.IsDir() {
			*respFolders = append(*respFolders, file.Name())
		} else {
			*respFiles = append(*respFiles, file.Name())
		}
	}

	return nil
}

func configFile(name string) bool {
	if name[0] == '.' {
		return true
	}
	return false
}

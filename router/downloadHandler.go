package router

import (
	"fmt"
	"io"
	"nas_server/logs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	_ "sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sasha-s/go-deadlock"
)

var (
	downloading = make(map[string]map[string]*DownloadInfo)
	downloaded = make(map[string]map[string]*DownloadResp)
	downloadSync = make(map[string]map[string]*DownloadFileSync)
	userLongPolling = make(map[string]chan struct{})
	mutex deadlock.RWMutex
)

func DownloadHandler(ctx *gin.Context) {
	if !CookieVerify(ctx) {
		return
	}
	deadlock.Opts.DeadlockTimeout = time.Second
	var downloadInfo DownloadInfo
	if err := ctx.ShouldBindJSON(&downloadInfo); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind downloadForm json")
	}
	downloadInfo.UserName = ctx.Param("username")
	filePath := filepath.Join(downloadInfo.FilePath, downloadInfo.FileName)
	info, _ := os.Stat(filePath)

	wg := new(deadlock.WaitGroup)
	if info.Mode().IsRegular() {
		wg.Add(1)
		go func () {
			defer wg.Done()
			downloadFileHandler(ctx, downloadInfo)
		}()
	} else if info.Mode().IsDir() {
		// downloadFolderHandler(ctx, downloadInfo, wg)
	}
	wg.Wait()
}

func downloadFileHandler(ctx *gin.Context, downloadInfo DownloadInfo) {
	mutex.Lock()
	filePath := filepath.Join(downloadInfo.FilePath, downloadInfo.FileName)
	downloadInfo.FileString = filePath
	if d, ok := downloading[downloadInfo.UserName]; ok {
		if _, ok := d[downloadInfo.FileString]; ok {
			ctx.JSON(http.StatusConflict, gin.H{})
			mutex.Unlock()
			return
		}
	} else {
		downloading[downloadInfo.UserName] = make(map[string]*DownloadInfo)
		downloaded[downloadInfo.UserName] = make(map[string]*DownloadResp)
		downloadSync[downloadInfo.UserName] = make(map[string]*DownloadFileSync)
	}
	if _, ok := userLongPolling[downloadInfo.UserName]; !ok {
		userLongPolling[downloadInfo.UserName] = make(chan struct{}, 1)
	}

	file, err := os.Open(filePath)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "server open file error"})
		logs.GetInstance().Logger.Errorf("file open error %s", err)
		mutex.Unlock()
		return
	}
	stat, err := file.Stat()
	if err != nil {
		logs.GetInstance().Logger.Errorf("stat error %s", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "server open file error"})
		mutex.Unlock()
		return
	}
	downloadInfo.FileLen = stat.Size()
	var fileSync DownloadFileSync
	fileSync.Pause = make(chan struct{}, 100)
	fileSync.Resume = make(chan struct{}, 100)
	fileSync.Cancel = make(chan struct{}, 100)
	now := time.Now()
	downloadInfo.Time = fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second())
	downloading[downloadInfo.UserName][downloadInfo.FileString] = &downloadInfo
	downloadSync[downloadInfo.UserName][downloadInfo.FileString] = &fileSync
	logs.GetInstance().Logger.Infof("downloads %+v", downloading)
	mutex.Unlock()

	isCancel := false
	defer func() {
		ctx.Writer.Flush()
		file.Close()
		fileSync.Mutex.Lock()
		if form, ok := downloading[downloadInfo.UserName]; ok {
			if info, ok := form[downloadInfo.FileString]; ok && !isCancel {
				downloaded[downloadInfo.UserName][info.FileString] = &DownloadResp{
					FileName: info.FileName,
					Progress: 1,
					Speed: "0MB/s",
					Time: info.Time,
					Status: "finish",
					FileString: info.FileString,
				}
			}
			delete(form, downloadInfo.FileString)
		}
		if form, ok := downloadSync[downloadInfo.UserName]; ok {
			delete(form, downloadInfo.FileString)
		}
		fileSync.Mutex.Unlock()
		userLongPolling[downloadInfo.UserName] <- struct{}{}
	}()

	ctx.Writer.Header().Set("Content-Disposition", "attachment; filename="+downloadInfo.FileName)
	ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
	ctx.Writer.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	ctx.Writer.Flush()
	var offset, bufSize int64 = 0, 1024 * 1024
	buf := make([]byte, bufSize)
	for {
		select {
		case <-fileSync.Pause:
			fileSync.Mutex.Lock()
			downloadInfo.Status = "waiting"
			downloadInfo.Speed = "0MB/s"
			fileSync.Mutex.Unlock()
			userLongPolling[downloadInfo.UserName] <- struct{}{}
			logs.GetInstance().Logger.Infof("pause download %s", downloadInfo.FileName)
			select {
			case <-fileSync.Resume:
				logs.GetInstance().Logger.Infof("resume download %s", downloadInfo.FileName)
			case <-fileSync.Cancel:
				isCancel = true
				logs.GetInstance().Logger.Infof("cancel download %s", downloadInfo.FileName)
				return
			}
		case <-fileSync.Cancel:
			isCancel = true
			logs.GetInstance().Logger.Infof("cancel download %s", downloadInfo.FileName)
			return
		default:
			startTime := time.Now()
			n, err := file.ReadAt(buf, offset)
			if err != nil && err != io.EOF {
				logs.GetInstance().Logger.Errorf("read file error %s", err)
				return
			}
			if n == 0 {
				return
			}
			_, err = ctx.Writer.Write(buf[:n])
			if err != nil {
				logs.GetInstance().Logger.Errorf("write file error %s", err)
				return
			}
			offset += int64(n)
			fileSync.Mutex.Lock()
			duration := time.Since(startTime)
			speed := float64(n) / duration.Seconds() / float64(bufSize)
			downloadInfo.Speed = fmt.Sprintf("%.2f", speed) + "MB/s"
			downloadInfo.DownloadLen = offset
			downloadInfo.Status = "downloading"
			fileSync.Mutex.Unlock()
			if len(userLongPolling[downloadInfo.UserName]) == 0 {
				userLongPolling[downloadInfo.UserName] <- struct{}{}
			}
		}
	}
}

func downloadFolderHandler(ctx *gin.Context, downloadInfo DownloadInfo, wg *deadlock.WaitGroup) {
	filePath := filepath.Join(downloadInfo.FilePath, downloadInfo.FileName)
	folders, files := make([]string, 0), make([]string, 0)
	getFiles(filePath, &folders, &files)
	
	wg.Add(len(files))
	for _, file := range files {
		downloadFileInfo := DownloadInfo {
			FilePath: downloadInfo.FilePath,
			FileName: file,
		}
		go func (downloadFileInfo DownloadInfo) {
			defer wg.Done()
			downloadFileHandler(ctx, downloadFileInfo)
		}(downloadFileInfo)
	}
	wg.Wait()
}

func DownloadProgressHandler(ctx *gin.Context) {
	if !CookieVerify(ctx) {
		return
	}
	userName := ctx.Param("username")
	if _, ok := userLongPolling[userName]; !ok {
		userLongPolling[userName] = make(chan struct{}, 1)
	}
	<-userLongPolling[userName]
	mutex.RLock()
	if downloadInfo, ok := downloading[userName]; !ok {
		ctx.JSON(http.StatusOK, gin.H{
			"downloading": []DownloadResp{},
			"downloaded": downloaded,
		})
	} else {
		resp := make([]DownloadResp, 0, len(downloadInfo))
		for _, info := range downloadInfo {
			progress := float64(info.DownloadLen) / float64(info.FileLen)
			resp = append(resp, DownloadResp{
				FileName: info.FileName,
				Progress: progress,
				Speed: info.Speed,
				Time: info.Time,
				Status: info.Status,
				FileString: info.FileString,
			})
		}
		sort.Slice(resp, func(i, j int) bool {
			return resp[i].Time < resp[j].Time
		})
		ctx.JSON(http.StatusOK, gin.H{
			"downloading": resp,
			"downloaded": downloaded,
		})
	}
	mutex.RUnlock()
}

func PauseDownloadHandler(ctx *gin.Context) {
	if !CookieVerify(ctx) {
		return
	}
	var request RequestFolder
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind RequestFolder json")
	}
	userName := ctx.Param("username")

	mutex.RLock()
	defer mutex.RUnlock()
	fileSync := downloadSync[userName]
	if sync, ok := fileSync[request.FolderName]; ok {
		sync.Mutex.Lock()
		sync.Pause <- struct{}{}
		sync.Mutex.Unlock()
	}
	ctx.JSON(http.StatusOK, gin.H{})
}

func ResumeDownloadHandler(ctx *gin.Context) {
	if !CookieVerify(ctx) {
		return
	}
	var request RequestFolder
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind RequestFolder json")
	}
	userName := ctx.Param("username")

	mutex.RLock()
	defer mutex.RUnlock()
	fileSync := downloadSync[userName]
	if sync, ok := fileSync[request.FolderName]; ok {
		sync.Mutex.Lock()
		sync.Resume <- struct{}{}
		sync.Mutex.Unlock()
	}
	
	ctx.JSON(http.StatusOK, gin.H{})
}

func CancelDownloadHandler(ctx *gin.Context) {
	if !CookieVerify(ctx) {
		return
	}
	var request RequestFolder
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind RequestFolder json")
	}
	userName := ctx.Param("username")

	mutex.RLock()
	defer mutex.RUnlock()
	fileSync := downloadSync[userName]
	if sync, ok := fileSync[request.FolderName]; ok {
		sync.Mutex.Lock()
		sync.Cancel <- struct{}{}
		sync.Mutex.Unlock()
	}
	
	ctx.JSON(http.StatusOK, gin.H{})
}

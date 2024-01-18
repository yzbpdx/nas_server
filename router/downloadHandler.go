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
	mutex deadlock.Mutex
)

func DownloadHandler(ctx *gin.Context) {
	deadlock.Opts.DeadlockTimeout = time.Second
	var downloadInfo DownloadInfo
	if err := ctx.ShouldBindJSON(&downloadInfo); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind downloadForm json")
	}
	downloadInfo.UserName = ctx.Param("username")
	// logs.GetInstance().Logger.Infof("downloadForm: %+v", downloadInfo)

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
	downloadInfo.Pause = make(chan struct{}, 100)
	downloadInfo.Resume = make(chan struct{}, 100)
	downloadInfo.Cancel = make(chan struct{}, 100)
	now := time.Now()
	downloadInfo.Time = fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second())
	downloading[downloadInfo.UserName][downloadInfo.FileString] = &downloadInfo
	logs.GetInstance().Logger.Infof("downloads %+v", downloading)
	downloadInfo.Wg.Add(1)
	mutex.Unlock()

	go func() {
		isCancel := false
		defer func() {
			ctx.Writer.Flush()
			file.Close()
			mutex.Lock()
			// fmt.Println("lock")
			downloadInfo.Wg.Done()
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
			mutex.Unlock()
			// fmt.Println("unlock")
		}()

		ctx.Writer.Header().Set("Content-Disposition", "attachment; filename="+downloadInfo.FileName)
		ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
		ctx.Writer.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
		ctx.Writer.Flush()
		var offset, bufSize int64 = 0, 1024 * 1024
		buf := make([]byte, bufSize)
		for {
			select {
			case <-downloadInfo.Pause:
				mutex.Lock()
				downloadInfo.Status = "waiting"
				downloadInfo.Speed = "0MB/s"
				mutex.Unlock()
				logs.GetInstance().Logger.Infof("pause download %s", downloadInfo.FileName)
				select {
				case <-downloadInfo.Resume:
					logs.GetInstance().Logger.Infof("resume download %s", downloadInfo.FileName)
				case <-downloadInfo.Cancel:
					isCancel = true
				logs.GetInstance().Logger.Infof("cancel download %s", downloadInfo.FileName)
				return
				}
			case <-downloadInfo.Cancel:
				isCancel = true
				logs.GetInstance().Logger.Infof("cancel download %s", downloadInfo.FileName)
				return
			default:
				// time.Sleep(5 * time.Second)
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
				duration := time.Since(startTime)
				mutex.Lock()
				speed := float64(n) / duration.Seconds() / float64(bufSize)
				downloadInfo.Speed = fmt.Sprintf("%.2f", speed) + "MB/s"
				downloadInfo.DownloadLen = offset
				downloadInfo.Status = "downloading"
				mutex.Unlock()
			}
		}
	}()
	downloadInfo.Wg.Wait()
}

func DownloadProgressHandler(ctx *gin.Context) {
	userName := ctx.Param("username")
	mutex.Lock()
	// fmt.Println("lock")
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
	mutex.Unlock()
}

func PauseDownloadHandler(ctx *gin.Context) {
	var request RequestFolder
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind RequestFolder json")
	}
	userName := ctx.Param("username")

	mutex.Lock()
	defer mutex.Unlock()
	downloadInfo := downloading[userName]
	if info, ok := downloadInfo[request.FolderName]; ok {
		info.Pause <- struct{}{}
	}
	ctx.JSON(http.StatusOK, gin.H{})
}

func ResumeDownloadHandler(ctx *gin.Context) {
	var request RequestFolder
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind RequestFolder json")
	}
	userName := ctx.Param("username")

	mutex.Lock()
	defer mutex.Unlock()
	downloadInfo := downloading[userName]
	if info, ok := downloadInfo[request.FolderName]; ok {
		info.Resume <- struct{}{}
	}
	
	ctx.JSON(http.StatusOK, gin.H{})
}

func CancelDownloadHandler(ctx *gin.Context) {
	var request RequestFolder
	if err := ctx.ShouldBindJSON(&request); err != nil {
		logs.GetInstance().Logger.Warnf("cannot bind RequestFolder json")
	}
	userName := ctx.Param("username")

	mutex.Lock()
	defer mutex.Unlock()
	downloadInfo := downloading[userName]
	if info, ok := downloadInfo[request.FolderName]; ok {
		info.Cancel <- struct{}{}
	}
	
	ctx.JSON(http.StatusOK, gin.H{})
}

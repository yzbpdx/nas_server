package router

import (
	"mime/multipart"

	"github.com/sasha-s/go-deadlock"
)

type UserForm struct {
	Username string `json:"username"`
	PassWord string `json:"password"`
}

type ErrorResp struct {
	ErrorMsg string `json:"error"`
}

type RequestFolder struct {
	FolderName string `json:"folderName"`
}

type DownloadInfo struct {
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
	UserName string `json:"userName"`

	FileString  string
	Time        string
	FileLen     int64
	DownloadLen int64
	Speed       string
	Status      string
}

type DownloadFileSync struct {
	Pause       chan struct{}
	Resume      chan struct{}
	Cancel      chan struct{}
	Wg          deadlock.WaitGroup
	Mutex       deadlock.RWMutex
}

type DownloadResp struct {
	FileName   string  `json:"fileName"`
	Progress   float64 `json:"progress"`
	Speed      string  `json:"speed"`
	Time       string  `json:"time"`
	Status     string  `json:"status"`
	FileString string  `json:"fileString"`
}

type UploadForm struct {
	UploadFolder string                `form:"uploadFolder"`
	FileName     string                `form:"fileName"`
	File         *multipart.FileHeader `form:"file"`
}

type CreateFolder struct {
	CurrentPath string `json:"currentPath"`
	FolderName  string `json:"folderName"`
}

type Rename struct {
	CurrentPath string `json:"currentPath"`
	OldName     string `json:"oldName"`
	NewName     string `json:"newName"`
}

type Delete struct {
	CurrentPath string `json:"currentPath"`
	DeleteName  string `json:"deleteName"`
}

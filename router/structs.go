package router

import "mime/multipart"

type LoginForm struct {
	Username string `json:"username"`
	PassWord string `json:"password"`
}

type ErrorResp struct {
	ErrorMsg string `json:"error"`
}

type RequestFolder struct {
	FolderName string `json:"folderName"`
}

type DownloadForm struct {
	FilePath string `json:"filePath"`
	FileName string `json:"fileName"`
}

type UploadForm struct {
	UploadFolder string `form:"uploadFolder"`
	FileName string `form:"fileName"`
	File *multipart.FileHeader `form:"file"`
}

type CreateFolder struct {
	CurrentPath string `json:"currentPath"`
	FolderName string `json:"folderName"`
}

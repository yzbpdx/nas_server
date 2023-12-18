package router

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

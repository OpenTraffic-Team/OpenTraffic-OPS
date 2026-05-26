package dto

// RemoteFileListRequest 文件列表请求
type RemoteFileListRequest struct {
	HostIP string `json:"hostIp" binding:"required"`
	Path   string `json:"path" binding:"required"`
}

// RemoteFileReadRequest 文件读取请求
type RemoteFileReadRequest struct {
	HostIP string `json:"hostIp" binding:"required"`
	Path   string `json:"path" binding:"required"`
}

// RemoteFileWriteRequest 文件写入请求
type RemoteFileWriteRequest struct {
	HostIP  string `json:"hostIp" binding:"required"`
	Path    string `json:"path" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// RemoteFileDeleteRequest 文件删除请求
type RemoteFileDeleteRequest struct {
	HostIP string `json:"hostIp" binding:"required"`
	Path   string `json:"path" binding:"required"`
}

// RemoteFileUploadRequest 文件上传请求
type RemoteFileUploadRequest struct {
	HostIP   string `json:"hostIp" binding:"required"`
	Path     string `json:"path" binding:"required"`
	FileName string `json:"fileName" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// RemoteFileDownloadRequest 文件下载请求
type RemoteFileDownloadRequest struct {
	HostIP string `json:"hostIp" binding:"required"`
	Path   string `json:"path" binding:"required"`
}

// RemoteFileMkdirRequest 创建目录请求
type RemoteFileMkdirRequest struct {
	HostIP string `json:"hostIp" binding:"required"`
	Path   string `json:"path" binding:"required"`
}

// FileInfo 文件信息
type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"isDir"`
	ModTime int64  `json:"modTime"`
	Mode    string `json:"mode,omitempty"`
}

// FileListResponse 文件列表响应
type FileListResponse struct {
	Path  string     `json:"path"`
	Files []FileInfo `json:"files"`
}

// FileReadResponse 文件读取响应
type FileReadResponse struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Size    int64  `json:"size"`
}

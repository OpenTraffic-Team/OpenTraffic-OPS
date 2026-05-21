package handler

import (
	"context"
	"encoding/base64"

	"github.com/gin-gonic/gin"

	"rtm-server/internal/dto"
	"rtm-server/internal/service"
	"rtm-server/pkg/response"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

// handleServiceError 统一处理service层错误
// 返回true表示已处理错误，调用方应直接return
func handleServiceError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	// 前端已断开连接，无需返回响应
	if err == context.Canceled {
		return true
	}
	response.Error(c, err.Error())
	return true
}

// RemoteHandler 远程文件操作处理器
type RemoteHandler struct {
	remoteService *service.RemoteService
}

// NewRemoteHandler 创建远程处理器
func NewRemoteHandler() *RemoteHandler {
	return &RemoteHandler{
		remoteService: service.NewRemoteService(),
	}
}

// RegisterRoutes 注册远程文件操作路由
func (h *RemoteHandler) RegisterRoutes(r *gin.RouterGroup) {
	remote := r.Group("/api/remote/file")
	{
		remote.POST("/list", h.FileList)
		remote.POST("/read", h.FileRead)
		remote.POST("/write", h.FileWrite)
		remote.POST("/delete", h.FileDelete)
		remote.POST("/upload", h.FileUpload)
		remote.POST("/download", h.FileDownload)
		remote.POST("/mkdir", h.FileMkdir)
	}
}

// FileList 获取文件列表
func (h *RemoteHandler) FileList(c *gin.Context) {
	var req dto.RemoteFileListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	result, err := h.remoteService.FileList(c.Request.Context(), req.HostIP, &req)
	if handleServiceError(c, err) {
		return
	}

	response.Success(c, result)
}

// FileRead 读取文件
func (h *RemoteHandler) FileRead(c *gin.Context) {
	var req dto.RemoteFileReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	result, err := h.remoteService.FileRead(c.Request.Context(), req.HostIP, &req)
	if handleServiceError(c, err) {
		return
	}

	response.Success(c, result)
}

// checkFileSize 检查base64编码的文件内容大小
func checkFileSize(base64Content string) bool {
	// base64编码后约增大33%
	decodedLen := base64.StdEncoding.DecodedLen(len(base64Content))
	return decodedLen <= maxFileSize
}

// FileWrite 写入文件
func (h *RemoteHandler) FileWrite(c *gin.Context) {
	var req dto.RemoteFileWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	if !checkFileSize(req.Content) {
		response.Error(c, "文件超过大小限制(最大10MB)")
		return
	}

	if err := h.remoteService.FileWrite(c.Request.Context(), req.HostIP, &req); handleServiceError(c, err) {
		return
	}

	response.Success(c, nil)
}

// FileDelete 删除文件
func (h *RemoteHandler) FileDelete(c *gin.Context) {
	var req dto.RemoteFileDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	if err := h.remoteService.FileDelete(c.Request.Context(), req.HostIP, &req); handleServiceError(c, err) {
		return
	}

	response.Success(c, nil)
}

// FileUpload 上传文件
func (h *RemoteHandler) FileUpload(c *gin.Context) {
	var req dto.RemoteFileUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	if !checkFileSize(req.Content) {
		response.Error(c, "文件超过大小限制(最大10MB)")
		return
	}

	if err := h.remoteService.FileUpload(c.Request.Context(), req.HostIP, &req); handleServiceError(c, err) {
		return
	}

	response.Success(c, nil)
}

// FileDownload 下载文件
func (h *RemoteHandler) FileDownload(c *gin.Context) {
	var req dto.RemoteFileDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	result, err := h.remoteService.FileDownload(c.Request.Context(), req.HostIP, &req)
	if handleServiceError(c, err) {
		return
	}

	response.Success(c, result)
}

// FileMkdir 创建目录
func (h *RemoteHandler) FileMkdir(c *gin.Context) {
	var req dto.RemoteFileMkdirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	if err := h.remoteService.FileMkdir(c.Request.Context(), req.HostIP, &req); handleServiceError(c, err) {
		return
	}

	response.Success(c, nil)
}

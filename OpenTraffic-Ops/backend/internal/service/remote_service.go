package service

import (
	"context"
	"encoding/json"
	"fmt"

	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/ws"
)

// RemoteService 远程文件操作服务
type RemoteService struct {
	hub *ws.Hub
}

// NewRemoteService 创建远程服务
func NewRemoteService() *RemoteService {
	return &RemoteService{
		hub: ws.GetHub(),
	}
}

// FileList 获取文件列表
func (s *RemoteService) FileList(ctx context.Context, hostIP string, req *dto.RemoteFileListRequest) (*dto.FileListResponse, error) {
	if !s.hub.IsHostProxyOnline(hostIP) {
		return nil, fmt.Errorf("主机不在线")
	}

	payload, _ := json.Marshal(ws.FileListPayload{Path: req.Path})
	msg := &ws.Message{
		Type:    ws.TypeFileList,
		Payload: payload,
	}

	resp, err := s.hub.Request(ctx, hostIP, msg)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	var result ws.FileListPayload
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	files := make([]dto.FileInfo, len(result.Files))
	for i, f := range result.Files {
		files[i] = dto.FileInfo{
			Name:    f.Name,
			Path:    f.Path,
			Size:    f.Size,
			IsDir:   f.IsDir,
			ModTime: f.ModTime,
			Mode:    f.Mode,
		}
	}

	return &dto.FileListResponse{
		Path:  result.Path,
		Files: files,
	}, nil
}

// FileRead 读取文件
func (s *RemoteService) FileRead(ctx context.Context, hostIP string, req *dto.RemoteFileReadRequest) (*dto.FileReadResponse, error) {
	if !s.hub.IsHostProxyOnline(hostIP) {
		return nil, fmt.Errorf("主机不在线")
	}

	payload, _ := json.Marshal(ws.FileReadPayload{Path: req.Path})
	msg := &ws.Message{
		Type:    ws.TypeFileRead,
		Payload: payload,
	}

	resp, err := s.hub.Request(ctx, hostIP, msg)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	var result ws.FileReadPayload
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &dto.FileReadResponse{
		Path:    result.Path,
		Content: result.Content,
		Size:    result.Size,
	}, nil
}

// FileWrite 写入文件
func (s *RemoteService) FileWrite(ctx context.Context, hostIP string, req *dto.RemoteFileWriteRequest) error {
	if !s.hub.IsHostProxyOnline(hostIP) {
		return fmt.Errorf("主机不在线")
	}

	payload, _ := json.Marshal(ws.FileWritePayload{
		Path:    req.Path,
		Content: req.Content,
	})
	msg := &ws.Message{
		Type:    ws.TypeFileWrite,
		Payload: payload,
	}

	resp, err := s.hub.Request(ctx, hostIP, msg)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	return nil
}

// FileDelete 删除文件
func (s *RemoteService) FileDelete(ctx context.Context, hostIP string, req *dto.RemoteFileDeleteRequest) error {
	if !s.hub.IsHostProxyOnline(hostIP) {
		return fmt.Errorf("主机不在线")
	}

	payload, _ := json.Marshal(ws.FileDeletePayload{Path: req.Path})
	msg := &ws.Message{
		Type:    ws.TypeFileDelete,
		Payload: payload,
	}

	resp, err := s.hub.Request(ctx, hostIP, msg)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	return nil
}

// FileUpload 上传文件
func (s *RemoteService) FileUpload(ctx context.Context, hostIP string, req *dto.RemoteFileUploadRequest) error {
	if !s.hub.IsHostProxyOnline(hostIP) {
		return fmt.Errorf("主机不在线")
	}

	payload, _ := json.Marshal(ws.FileUploadPayload{
		Path:     req.Path,
		FileName: req.FileName,
		Content:  req.Content,
	})
	msg := &ws.Message{
		Type:    ws.TypeFileUpload,
		Payload: payload,
	}

	resp, err := s.hub.Request(ctx, hostIP, msg)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	return nil
}

// FileDownload 下载文件
func (s *RemoteService) FileDownload(ctx context.Context, hostIP string, req *dto.RemoteFileDownloadRequest) (*dto.FileReadResponse, error) {
	if !s.hub.IsHostProxyOnline(hostIP) {
		return nil, fmt.Errorf("主机不在线")
	}

	payload, _ := json.Marshal(ws.FileDownloadPayload{Path: req.Path})
	msg := &ws.Message{
		Type:    ws.TypeFileDownload,
		Payload: payload,
	}

	resp, err := s.hub.Request(ctx, hostIP, msg)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Error)
	}

	var result ws.FileReadPayload
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &dto.FileReadResponse{
		Path:    result.Path,
		Content: result.Content,
		Size:    result.Size,
	}, nil
}

// FileMkdir 创建目录
func (s *RemoteService) FileMkdir(ctx context.Context, hostIP string, req *dto.RemoteFileMkdirRequest) error {
	if !s.hub.IsHostProxyOnline(hostIP) {
		return fmt.Errorf("主机不在线")
	}

	payload, _ := json.Marshal(ws.FileMkdirPayload{Path: req.Path})
	msg := &ws.Message{
		Type:    ws.TypeFileMkdir,
		Payload: payload,
	}

	resp, err := s.hub.Request(ctx, hostIP, msg)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	return nil
}

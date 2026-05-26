package filemanager

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10MB
)

// Manager 文件管理器
type Manager struct{}

// New 创建文件管理器
func New() *Manager {
	return &Manager{}
}

// validatePath 验证路径安全（防止目录遍历）
func validatePath(path string) error {
	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查是否包含..
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("非法路径: 包含目录遍历字符")
	}

	// Windows下检查盘符穿越
	if filepath.VolumeName(cleanPath) != "" {
		// 允许绝对路径，但要检查是否包含..
		parts := strings.Split(cleanPath, string(filepath.Separator))
		for _, part := range parts {
			if part == ".." {
				return fmt.Errorf("非法路径: 包含目录遍历字符")
			}
		}
	}

	return nil
}

// List 列出目录内容
func (m *Manager) List(path string) ([]FileInfo, error) {
	if err := validatePath(path); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    fullPath,
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().Unix(),
			Mode:    info.Mode().String(),
		})
	}

	return files, nil
}

// Read 读取文件内容（返回base64编码）
func (m *Manager) Read(path string) (string, int64, error) {
	if err := validatePath(path); err != nil {
		return "", 0, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", 0, fmt.Errorf("获取文件信息失败: %w", err)
	}

	if info.IsDir() {
		return "", 0, fmt.Errorf("不能读取目录")
	}

	if info.Size() > maxFileSize {
		return "", 0, fmt.Errorf("文件超过大小限制(最大%dMB)", maxFileSize/1024/1024)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", 0, fmt.Errorf("读取文件失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(data), info.Size(), nil
}

// Write 写入文件（内容base64编码）
func (m *Manager) Write(path string, content string) error {
	if err := validatePath(path); err != nil {
		return err
	}

	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return fmt.Errorf("解码内容失败: %w", err)
	}

	if len(data) > maxFileSize {
		return fmt.Errorf("文件超过大小限制(最大%dMB)", maxFileSize/1024/1024)
	}

	// 确保父目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建父目录失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// Delete 删除文件或目录
func (m *Manager) Delete(path string) error {
	if err := validatePath(path); err != nil {
		return err
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}

	return nil
}

// Upload 上传文件（内容base64编码）
func (m *Manager) Upload(dirPath, fileName string, content string) error {
	if err := validatePath(dirPath); err != nil {
		return err
	}
	if err := validatePath(fileName); err != nil {
		return err
	}

	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return fmt.Errorf("解码内容失败: %w", err)
	}

	if len(data) > maxFileSize {
		return fmt.Errorf("文件超过大小限制(最大%dMB)", maxFileSize/1024/1024)
	}

	// 确保目录存在
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	fullPath := filepath.Join(dirPath, fileName)
	if err := validatePath(fullPath); err != nil {
		return err
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// Download 下载文件（返回base64编码的内容）
func (m *Manager) Download(path string) (string, int64, error) {
	// 复用Read逻辑
	return m.Read(path)
}

// Mkdir 创建目录
func (m *Manager) Mkdir(path string) error {
	if err := validatePath(path); err != nil {
		return err
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	return nil
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

// Copy 复制文件
func (m *Manager) Copy(src, dst string) error {
	if err := validatePath(src); err != nil {
		return err
	}
	if err := validatePath(dst); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	return nil
}

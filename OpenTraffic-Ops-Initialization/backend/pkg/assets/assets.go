package assets

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	assetsDir     string
	assetsDirOnce sync.Once
)

// getAssetsDir 定位 images 资源目录，优先使用 ASSETS_PATH 环境变量，
// 其次尝试源码目录、可执行文件目录和当前工作目录。
func getAssetsDir() string {
	assetsDirOnce.Do(func() {
		if p := os.Getenv("ASSETS_PATH"); p != "" {
			assetsDir = p
			return
		}

		// 1. 基于源码文件位置（开发时）
		if _, file, _, ok := runtime.Caller(0); ok {
			dir := filepath.Join(filepath.Dir(file), "images")
			if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
				assetsDir = dir
				return
			}
		}

		// 2. 基于可执行文件位置（部署时）
		if exe, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exe)
			candidates := []string{
				filepath.Join(exeDir, "pkg", "assets", "images"),
				filepath.Join(exeDir, "assets", "images"),
			}
			for _, dir := range candidates {
				if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
					assetsDir = dir
					return
				}
			}
		}

		// 3. 基于当前工作目录兜底
		if cwd, err := os.Getwd(); err == nil {
			candidates := []string{
				filepath.Join(cwd, "pkg", "assets", "images"),
				filepath.Join(cwd, "backend", "pkg", "assets", "images"),
				filepath.Join(cwd, "..", "pkg", "assets", "images"),
			}
			for _, dir := range candidates {
				if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
					assetsDir = dir
					return
				}
			}
		}
	})
	return assetsDir
}

func imagePath(name string) string {
	return filepath.Join(getAssetsDir(), name)
}

// GetImageReader returns a ReadCloser for the named image tar file located in images/.
func GetImageReader(name string) (io.ReadCloser, error) {
	path := imagePath(name)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("image not found: %s", name)
	}
	return f, nil
}

// HasImage checks whether an embedded image tar with the given name exists.
func HasImage(name string) bool {
	_, err := os.Stat(imagePath(name))
	return err == nil
}

// GetBinaryReader returns a ReadCloser for the named binary file located in images/.
func GetBinaryReader(name string) (io.ReadCloser, error) {
	path := imagePath(name)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("binary not found: %s", name)
	}
	return f, nil
}

// HasBinary checks whether an embedded binary with the given name exists.
func HasBinary(name string) bool {
	_, err := os.Stat(imagePath(name))
	return err == nil
}

// GetConfigReader returns a ReadCloser for the config file located in images/.
func GetConfigReader(name string) (io.ReadCloser, error) {
	path := imagePath(name)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config not found: %s", name)
	}
	return f, nil
}

// HasConfig checks whether an embedded config file with the given name exists.
func HasConfig(name string) bool {
	_, err := os.Stat(imagePath(name))
	return err == nil
}

// ReadEmbeddedFile reads the full content of an embedded file from images/.
func ReadEmbeddedFile(name string) ([]byte, error) {
	path := imagePath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("embedded file not found: %s", name)
	}
	return data, nil
}

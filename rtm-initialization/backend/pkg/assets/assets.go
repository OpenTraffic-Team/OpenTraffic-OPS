package assets

import (
	"embed"
	"fmt"
	"io"
)

//go:embed all:images
var imageFS embed.FS

// GetImageReader returns a ReadCloser for the named image tar file located in images/.
func GetImageReader(name string) (io.ReadCloser, error) {
	path := fmt.Sprintf("images/%s", name)
	f, err := imageFS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("embedded image not found: %s", name)
	}
	return f, nil
}

// HasImage checks whether an embedded image tar with the given name exists.
func HasImage(name string) bool {
	f, err := imageFS.Open(fmt.Sprintf("images/%s", name))
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// GetBinaryReader returns a ReadCloser for the named binary file located in images/.
func GetBinaryReader(name string) (io.ReadCloser, error) {
	path := fmt.Sprintf("images/%s", name)
	f, err := imageFS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("embedded binary not found: %s", name)
	}
	return f, nil
}

// HasBinary checks whether an embedded binary with the given name exists.
func HasBinary(name string) bool {
	f, err := imageFS.Open(fmt.Sprintf("images/%s", name))
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// GetConfigReader returns a ReadCloser for the config file located in images/.
func GetConfigReader(name string) (io.ReadCloser, error) {
	path := fmt.Sprintf("images/%s", name)
	f, err := imageFS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("embedded config not found: %s", name)
	}
	return f, nil
}

// HasConfig checks whether an embedded config file with the given name exists.
func HasConfig(name string) bool {
	f, err := imageFS.Open(fmt.Sprintf("images/%s", name))
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// ReadEmbeddedFile reads the full content of an embedded file from images/.
func ReadEmbeddedFile(name string) ([]byte, error) {
	path := fmt.Sprintf("images/%s", name)
	data, err := imageFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("embedded file not found: %s", name)
	}
	return data, nil
}

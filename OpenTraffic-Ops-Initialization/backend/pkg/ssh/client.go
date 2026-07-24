package ssh

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
)

// Config SSH连接配置
type Config struct {
	Host       string
	Port       int
	Username   string
	AuthType   string // "password" / "key"
	Password   string
	PrivateKey string
	Passphrase string
}

// Client SSH客户端
type Client struct {
	client *gossh.Client
	sftp   *sftp.Client
}

// NewClient 创建SSH客户端
func NewClient(config *Config) (*Client, error) {
	var authMethods []gossh.AuthMethod

	switch config.AuthType {
	case "password":
		if config.Password == "" {
			return nil, fmt.Errorf("password is required for password authentication")
		}
		authMethods = append(authMethods, gossh.Password(config.Password))

	case "key":
		if config.PrivateKey == "" {
			return nil, fmt.Errorf("private key is required for key authentication")
		}
		var signer gossh.Signer
		var err error
		if config.Passphrase != "" {
			signer, err = gossh.ParsePrivateKeyWithPassphrase([]byte(config.PrivateKey), []byte(config.Passphrase))
		} else {
			signer, err = gossh.ParsePrivateKey([]byte(config.PrivateKey))
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, gossh.PublicKeys(signer))

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", config.AuthType)
	}

	sshConfig := &gossh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := gossh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	// 创建SFTP客户端
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return &Client{
		client: client,
		sftp:   sftpClient,
	}, nil
}

// Close 关闭连接
func (c *Client) Close() error {
	if c.sftp != nil {
		_ = c.sftp.Close()
	}
	if c.client != nil {
		_ = c.client.Close()
	}
	return nil
}

// TestConnection 测试连接
func (c *Client) TestConnection() error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var out bytes.Buffer
	session.Stdout = &out
	if err := session.Run("echo ok"); err != nil {
		return fmt.Errorf("failed to run test command: %w", err)
	}

	if out.String() != "ok\n" {
		return fmt.Errorf("unexpected response: %s", out.String())
	}

	return nil
}

// Execute 执行远程命令并返回输出（带30秒超时）
func (c *Client) Execute(command string) (string, error) {
	return c.ExecuteWithTimeout(command, 30*time.Second)
}

// ExecuteWithTimeout 执行远程命令并返回输出（可自定义超时）
func (c *Client) ExecuteWithTimeout(command string, timeout time.Duration) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var out bytes.Buffer
	var errBuf bytes.Buffer
	session.Stdout = &out
	session.Stderr = &errBuf

	done := make(chan error, 1)
	go func() {
		done <- session.Run(command)
	}()

	select {
	case err := <-done:
		if err != nil {
			stderr := errBuf.String()
			if stderr != "" {
				return out.String(), fmt.Errorf("command failed: %w, stderr: %s", err, stderr)
			}
			return out.String(), fmt.Errorf("command failed: %w", err)
		}
		return out.String(), nil
	case <-time.After(timeout):
		return "", fmt.Errorf("command execution timed out after %v", timeout)
	}
}

// UploadFile 通过SFTP上传文件
func (c *Client) UploadFile(reader io.Reader, remotePath string, size int64) error {
	remotePath = filepath.ToSlash(remotePath)
	// 确保目标目录存在
	dir := filepath.ToSlash(filepath.Dir(remotePath))
	if err := c.sftp.MkdirAll(dir); err != nil {
		return fmt.Errorf("failed to create remote directory %s: %w", dir, err)
	}

	// 创建远程文件
	remoteFile, err := c.sftp.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	// 写入文件内容
	written, err := io.Copy(remoteFile, reader)
	if err != nil {
		return fmt.Errorf("failed to write remote file: %w", err)
	}

	if size > 0 && written != size {
		return fmt.Errorf("incomplete upload: expected %d bytes, wrote %d bytes", size, written)
	}

	return nil
}

// ExecuteWithLog 执行命令并收集stdout和stderr
func (c *Client) ExecuteWithLog(command string) (stdout, stderr string, err error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	session.Stdout = &outBuf
	session.Stderr = &errBuf

	execErr := session.Run(command)
	return outBuf.String(), errBuf.String(), execErr
}

// ReadFile 通过SFTP读取远程文件
func (c *Client) ReadFile(remotePath string) ([]byte, error) {
	remotePath = filepath.ToSlash(remotePath)
	remoteFile, err := c.sftp.Open(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	data, err := io.ReadAll(remoteFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote file %s: %w", remotePath, err)
	}
	return data, nil
}

// WriteFile 通过SFTP写入远程文件
func (c *Client) WriteFile(data []byte, remotePath string) error {
	remotePath = filepath.ToSlash(remotePath)
	// 确保目标目录存在
	dir := filepath.ToSlash(filepath.Dir(remotePath))
	if err := c.sftp.MkdirAll(dir); err != nil {
		return fmt.Errorf("failed to create remote directory %s: %w", dir, err)
	}

	remoteFile, err := c.sftp.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
	}
	defer remoteFile.Close()

	written, err := remoteFile.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write remote file: %w", err)
	}
	if written != len(data) {
		return fmt.Errorf("incomplete write: expected %d bytes, wrote %d bytes", len(data), written)
	}
	return nil
}

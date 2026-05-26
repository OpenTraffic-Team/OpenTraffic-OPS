package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
)

const (
	shellTimeout = 5 * time.Minute
)

// ShellExecutor 持久Shell执行器（基于PTY）
type ShellExecutor struct {
	ptmx   *os.File
	cmd    *exec.Cmd
	mu     sync.Mutex
	closed bool
}

// NewShellExecutor 创建新的Shell执行器（使用PTY）
func NewShellExecutor() (*ShellExecutor, error) {
	shell := getDefaultShell()

	cmd := exec.Command(shell)

	// 设置工作目录为 home
	if home, err := os.UserHomeDir(); err == nil {
		cmd.Dir = home
	}

	// 设置环境变量（保留 TERM 以支持颜色）
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	// 使用PTY启动shell
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("启动PTY shell失败: %w", err)
	}

	se := &ShellExecutor{
		ptmx:   ptmx,
		cmd:    cmd,
		closed: false,
	}

	// 启动超时监控
	go se.timeoutMonitor()

	return se, nil
}

// Write 向shell写入数据
func (se *ShellExecutor) Write(data []byte) error {
	se.mu.Lock()
	defer se.mu.Unlock()

	if se.closed {
		return fmt.Errorf("shell已关闭")
	}

	_, err := se.ptmx.Write(data)
	return err
}

// Read 读取shell输出（PTY合并了stdout和stderr）
func (se *ShellExecutor) Read(p []byte) (int, error) {
	return se.ptmx.Read(p)
}

// ReadStdout 读取stdout（兼容旧接口，实际与Read相同）
func (se *ShellExecutor) ReadStdout(p []byte) (int, error) {
	return se.ptmx.Read(p)
}

// ReadStderr 读取stderr（PTY已合并，返回EOF）
func (se *ShellExecutor) ReadStderr(p []byte) (int, error) {
	return 0, io.EOF
}

// Resize 调整终端大小
func (se *ShellExecutor) Resize(cols, rows int) error {
	se.mu.Lock()
	defer se.mu.Unlock()

	if se.closed {
		return fmt.Errorf("shell已关闭")
	}

	return pty.Setsize(se.ptmx, &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	})
}

// Close 关闭shell
func (se *ShellExecutor) Close() error {
	se.mu.Lock()
	defer se.mu.Unlock()

	if se.closed {
		return nil
	}
	se.closed = true

	// 关闭PTY
	se.ptmx.Close()

	// 结束进程
	if se.cmd != nil && se.cmd.Process != nil {
		killProcessGroup(se.cmd.Process.Pid)
		// 等待进程结束
		go se.cmd.Wait()
	}

	return nil
}

// IsRunning 检查shell是否还在运行
func (se *ShellExecutor) IsRunning() bool {
	se.mu.Lock()
	defer se.mu.Unlock()

	if se.closed || se.cmd == nil || se.cmd.Process == nil {
		return false
	}

	return isProcessRunning(se.cmd.Process)
}

// timeoutMonitor 超时监控，超过5分钟自动关闭
func (se *ShellExecutor) timeoutMonitor() {
	timer := time.NewTimer(shellTimeout)
	defer timer.Stop()

	<-timer.C

	se.mu.Lock()
	closed := se.closed
	se.mu.Unlock()

	if !closed {
		se.Close()
	}
}

// getDefaultShell 获取默认shell
func getDefaultShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/sh"
}

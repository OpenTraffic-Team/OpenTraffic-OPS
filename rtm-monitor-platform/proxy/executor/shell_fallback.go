//go:build !linux

package executor

import (
	"os"
	"os/exec"
)

// setupProcessGroup fallback实现（非Linux平台）
func setupProcessGroup(cmd *exec.Cmd) {
	// Linux-only: 无操作
}

// killProcessGroup fallback实现（非Linux平台）
func killProcessGroup(pid int) error {
	// Linux-only: 无操作
	return nil
}

// isProcessRunning fallback实现（非Linux平台）
func isProcessRunning(proc *os.Process) bool {
	// Linux-only: 假设进程仍在运行
	return proc != nil
}

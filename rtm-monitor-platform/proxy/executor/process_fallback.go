//go:build !linux

package executor

import "os/exec"

func setupCmdSysProcAttr(cmd *exec.Cmd) {
	// Linux-only: 无操作
}

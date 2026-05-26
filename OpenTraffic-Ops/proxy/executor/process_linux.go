//go:build linux

package executor

import (
	"os/exec"
	"syscall"
)

func setupCmdSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

//go:build linux

package executor

import (
	"os"
	"os/exec"
	"syscall"
)

func setupProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func killProcessGroup(pid int) error {
	return syscall.Kill(-pid, syscall.SIGKILL)
}

func isProcessRunning(proc *os.Process) bool {
	err := proc.Signal(syscall.Signal(0))
	return err == nil
}

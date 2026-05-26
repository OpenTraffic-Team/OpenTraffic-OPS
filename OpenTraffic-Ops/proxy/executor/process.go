package executor

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// Executor 进程指令执行器
type Executor struct {
	// 进程名到启动命令的映射
	processCmds map[string]string
}

// New 创建执行器
func New(cmds map[string]string) *Executor {
	return &Executor{processCmds: cmds}
}

// Execute 执行指令
func (e *Executor) Execute(cmdType, processName, params string) (bool, string) {
	switch cmdType {
	case "startProcess":
		return e.startProcess(processName)
	case "stopProcess":
		return e.stopProcess(processName)
	case "restartProcess":
		ok, msg := e.stopProcess(processName)
		if !ok {
			return ok, msg
		}
		time.Sleep(1 * time.Second)
		return e.startProcess(processName)
	default:
		return false, fmt.Sprintf("未知指令类型: %s", cmdType)
	}
}

// startProcess 启动进程
func (e *Executor) startProcess(name string) (bool, string) {
	// 如果进程已在运行，直接返回
	if pid, err := findProcessPID(name); err == nil {
		return true, fmt.Sprintf("进程 %s 已在运行 (PID=%d)", name, pid)
	}

	// 获取启动命令
	cmdStr, ok := e.processCmds[name]
	if !ok || cmdStr == "" {
		return false, fmt.Sprintf("未配置进程 %s 的启动命令", name)
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	setupCmdSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return false, fmt.Sprintf("启动进程 %s 失败: %v", name, err)
	}

	// 非阻塞，立即返回
	go func() {
		_ = cmd.Wait()
	}()

	return true, fmt.Sprintf("进程 %s 已启动 (PID=%d)", name, cmd.Process.Pid)
}

// stopProcess 停止进程
func (e *Executor) stopProcess(name string) (bool, string) {
	pid, err := findProcessPID(name)
	if err != nil {
		return false, fmt.Sprintf("进程 %s 未找到: %v", name, err)
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		return false, fmt.Sprintf("获取进程 %s (PID=%d) 失败: %v", name, pid, err)
	}

	// 先尝试优雅终止
	if err := p.Terminate(); err != nil {
		// 优雅终止失败则强制 Kill
		if err := p.Kill(); err != nil {
			return false, fmt.Sprintf("终止进程 %s (PID=%d) 失败: %v", name, pid, err)
		}
	}

	return true, fmt.Sprintf("进程 %s (PID=%d) 已停止", name, pid)
}

// findProcessPID 根据进程名查找 PID
func findProcessPID(name string) (int32, error) {
	procs, err := process.Processes()
	if err != nil {
		return 0, err
	}
	for _, p := range procs {
		pname, err := p.Name()
		if err != nil {
			continue
		}
		if strings.Contains(pname, name) {
			return p.Pid, nil
		}
		// 也尝试匹配命令行
		cmdline, _ := p.Cmdline()
		if strings.Contains(cmdline, name) {
			return p.Pid, nil
		}
	}
	return 0, fmt.Errorf("未找到进程 %s", name)
}

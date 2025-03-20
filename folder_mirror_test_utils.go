package main

import (
	"os"
	"os/exec"
	"testing"
)

// execCommand 是标准库exec.Command函数的可测试包装器
var execCommand = exec.Command

// 用于测试的模拟命令
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// 设置fake命令环境
func setupFakeExecCommand(t *testing.T) func() {
	// 保存原来的函数
	oldExecCommand := execCommand
	
	// 替换为fake版本
	execCommand = fakeExecCommand
	
	// 返回清理函数
	return func() {
		execCommand = oldExecCommand
	}
}

// 模拟成功的命令
func mockSuccess() {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(0)
}

// 模拟失败的命令
func mockFailure() {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(1)
}

// 测试辅助进程
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	
	// 获取命令名称（通常是第三个参数，因为第一个是可执行文件，第二个是--）
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	
	if len(args) == 0 {
		// 没有找到命令
		os.Exit(1)
	}
	
	// 根据命令类型返回不同的退出码
	cmd := args[0]
	switch cmd {
	case "rsync":
		// rsync成功
		os.Exit(0)
	case "echo":
		// echo成功
		os.Exit(0)
	case "vim":
		// vim成功
		os.Exit(0)
	case "fail-rsync":
		// rsync失败
		os.Exit(1)
	default:
		// 未知命令
		os.Exit(127)
	}
} 
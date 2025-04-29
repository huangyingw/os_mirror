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

// 测试模拟成功的命令
func mockSuccess() {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	osExit(0)
}

// 测试模拟失败的命令
func mockFailure() {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	osExit(1)
}

// 测试辅助进程 - 这个函数必须保持此名称以被mock命令调用
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
		osExit(1)
	}
	
	// 根据命令类型返回不同的退出码
	cmd := args[0]
	switch cmd {
	case "rsync":
		// rsync成功
		osExit(0)
	case "echo":
		// echo成功
		osExit(0)
	case "fail-rsync":
		// rsync失败
		osExit(1)
	default:
		// 未知命令
		osExit(127)
	}
}

// 测试模拟命令功能
func TestMockFunctions(t *testing.T) {
	// 保存原始的osExit函数
	oldOsExit := osExit
	exitCode := 0
	osExit = func(code int) {
		exitCode = code
		// 不真正退出
	}
	defer func() { osExit = oldOsExit }()
	
	// 测试mockSuccess
	oldEnv := os.Getenv("GO_WANT_HELPER_PROCESS")
	os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	mockSuccess()
	if exitCode != 0 {
		t.Errorf("mockSuccess应当返回退出码0，但得到: %d", exitCode)
	}
	
	// 测试mockFailure
	exitCode = 0
	mockFailure()
	if exitCode != 1 {
		t.Errorf("mockFailure应当返回退出码1，但得到: %d", exitCode)
	}
	
	// 恢复环境变量
	os.Setenv("GO_WANT_HELPER_PROCESS", oldEnv)
	
	// 测试当GO_WANT_HELPER_PROCESS不为1时的行为
	os.Setenv("GO_WANT_HELPER_PROCESS", "0")
	exitCode = -1 // 设置一个不会被函数改变的值
	mockSuccess()
	if exitCode != -1 {
		t.Errorf("当GO_WANT_HELPER_PROCESS不为1时，mockSuccess不应当改变退出码")
	}
	
	mockFailure()
	if exitCode != -1 {
		t.Errorf("当GO_WANT_HELPER_PROCESS不为1时，mockFailure不应当改变退出码")
	}
}

// 测试TestHelperProcess本身的行为
func TestHelperProcessFunction(t *testing.T) {
	// 保存原始的osExit函数
	oldOsExit := osExit
	exitCode := 0
	osExit = func(code int) {
		exitCode = code
		// 不真正退出
	}
	defer func() { osExit = oldOsExit }()
	
	// 保存和设置环境变量
	oldEnv := os.Getenv("GO_WANT_HELPER_PROCESS")
	defer os.Setenv("GO_WANT_HELPER_PROCESS", oldEnv)
	
	// 测试当GO_WANT_HELPER_PROCESS不为1时的行为
	os.Setenv("GO_WANT_HELPER_PROCESS", "0")
	TestHelperProcess(t)
	// 这种情况下函数应该提前返回，不会改变退出码
	if exitCode != 0 {
		t.Errorf("当GO_WANT_HELPER_PROCESS不为1时，TestHelperProcess不应当改变退出码")
	}
	
	// 测试GO_WANT_HELPER_PROCESS为1时，但没有命令参数的情况
	os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	oldArgs := os.Args
	os.Args = []string{"test"} // 没有-- 后面的命令参数
	defer func() { os.Args = oldArgs }()
	
	TestHelperProcess(t)
	if exitCode != 1 {
		t.Errorf("当没有命令参数时，TestHelperProcess应当返回退出码1，但得到: %d", exitCode)
	}
	
	// 测试带有有效命令的情况
	os.Args = []string{"test", "--", "rsync"}
	exitCode = 0
	TestHelperProcess(t)
	if exitCode != 0 {
		t.Errorf("对于rsync命令，TestHelperProcess应当返回退出码0，但得到: %d", exitCode)
	}
	
	// 测试带有失败命令的情况
	os.Args = []string{"test", "--", "fail-rsync"}
	exitCode = 0
	TestHelperProcess(t)
	if exitCode != 1 {
		t.Errorf("对于fail-rsync命令，TestHelperProcess应当返回退出码1，但得到: %d", exitCode)
	}
	
	// 测试带有未知命令的情况
	os.Args = []string{"test", "--", "unknown-command"}
	exitCode = 0
	TestHelperProcess(t)
	if exitCode != 127 {
		t.Errorf("对于未知命令，TestHelperProcess应当返回退出码127，但得到: %d", exitCode)
	}
} 
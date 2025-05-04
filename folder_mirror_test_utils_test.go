package main

import (
	"os"
	"testing"
)

// 测试 mockSuccess 函数
func TestMockSuccess(t *testing.T) {
	// 设置环境变量
	const testStdout = "mock success stdout"
	const testStderr = "mock success stderr"
	
	// 保存原始环境并在测试结束时恢复
	oldMockStdout := os.Getenv("MOCK_STDOUT")
	oldMockStderr := os.Getenv("MOCK_STDERR")
	oldGoWant := os.Getenv("GO_WANT_HELPER_PROCESS")
	defer func() {
		os.Setenv("MOCK_STDOUT", oldMockStdout)
		os.Setenv("MOCK_STDERR", oldMockStderr)
		os.Setenv("GO_WANT_HELPER_PROCESS", oldGoWant)
	}()
	
	// 保存原始的osExit
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	
	// 测试 GO_WANT_HELPER_PROCESS 不等于 "1" 的情况
	os.Setenv("GO_WANT_HELPER_PROCESS", "0")
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
	}
	
	mockSuccess()
	
	if exitCalled {
		t.Error("当 GO_WANT_HELPER_PROCESS 不等于 '1' 时，mockSuccess 不应调用 osExit")
	}
	
	// 测试 GO_WANT_HELPER_PROCESS 等于 "1" 的情况
	os.Setenv("MOCK_STDOUT", testStdout)
	os.Setenv("MOCK_STDERR", testStderr)
	os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	
	exitCode := -1
	osExit = func(code int) {
		exitCode = code
	}
	
	mockSuccess()
	
	if exitCode != 0 {
		t.Errorf("mockSuccess 返回了非零退出码: %d", exitCode)
	}
}

// 测试 mockFailure 函数
func TestMockFailure(t *testing.T) {
	// 设置环境变量
	const testStdout = "mock failure stdout"
	const testStderr = "mock failure stderr"
	
	// 保存原始环境并在测试结束时恢复
	oldMockStdout := os.Getenv("MOCK_STDOUT")
	oldMockStderr := os.Getenv("MOCK_STDERR")
	oldGoWant := os.Getenv("GO_WANT_HELPER_PROCESS")
	defer func() {
		os.Setenv("MOCK_STDOUT", oldMockStdout)
		os.Setenv("MOCK_STDERR", oldMockStderr)
		os.Setenv("GO_WANT_HELPER_PROCESS", oldGoWant)
	}()
	
	// 保存原始的osExit
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	
	// 测试 GO_WANT_HELPER_PROCESS 不等于 "1" 的情况
	os.Setenv("GO_WANT_HELPER_PROCESS", "0")
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
	}
	
	mockFailure()
	
	if exitCalled {
		t.Error("当 GO_WANT_HELPER_PROCESS 不等于 '1' 时，mockFailure 不应调用 osExit")
	}
	
	// 测试 GO_WANT_HELPER_PROCESS 等于 "1" 的情况
	os.Setenv("MOCK_STDOUT", testStdout)
	os.Setenv("MOCK_STDERR", testStderr)
	os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	
	exitCode := -1
	osExit = func(code int) {
		exitCode = code
	}
	
	mockFailure()
	
	if exitCode != 1 {
		t.Errorf("mockFailure 返回了非1退出码: %d", exitCode)
	}
}

// 测试TestMockFunctions函数
func TestMockFunctionsOnly(t *testing.T) {
	// 直接运行已有的TestMockFunctions测试
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
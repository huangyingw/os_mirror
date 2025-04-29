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
	
	// 设置测试环境变量
	os.Setenv("MOCK_STDOUT", testStdout)
	os.Setenv("MOCK_STDERR", testStderr)
	os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	
	// 测试 mockSuccess
	exitCode := 0
	oldOsExit := osExit
	osExit = func(code int) {
		exitCode = code
	}
	defer func() { osExit = oldOsExit }()
	
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
	
	// 设置测试环境变量
	os.Setenv("MOCK_STDOUT", testStdout)
	os.Setenv("MOCK_STDERR", testStderr)
	os.Setenv("GO_WANT_HELPER_PROCESS", "1")
	
	// 测试 mockFailure
	exitCode := 0
	oldOsExit := osExit
	osExit = func(code int) {
		exitCode = code
	}
	defer func() { osExit = oldOsExit }()
	
	mockFailure()
	
	if exitCode != 1 {
		t.Errorf("mockFailure 返回了非1退出码: %d", exitCode)
	}
} 
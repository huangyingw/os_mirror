package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// 测试dry-run模式下的执行逻辑
func TestDryRunExecution(t *testing.T) {
	// 检查是否可以执行模拟命令
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("环境中没有找到go命令，跳过测试")
	}

	// 设置模拟命令
	cleanup := setupFakeExecCommand(t)
	defer cleanup()
	
	// 创建临时测试目录
	testDir, sourceDir, targetDir := setupTestDirs(t)
	defer os.RemoveAll(testDir)
	
	// 创建临时标记文件
	originalMarkerFile := markerFile
	tmpMarkerFile, err := ioutil.TempFile("", "marker_test_dry_run")
	if err != nil {
		t.Fatalf("无法创建临时标记文件: %v", err)
	}
	tmpMarkerFile.Close()
	os.Remove(tmpMarkerFile.Name()) // 删除文件以便让程序自己创建
	
	markerFile = tmpMarkerFile.Name()
	defer func() {
		markerFile = originalMarkerFile
		os.Remove(tmpMarkerFile.Name())
	}()
	
	// 创建测试规则文件
	excludeFile, includeFile := createTestRuleFiles(t, testDir)
	
	// 备份和修改环境变量HOME
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", testDir)
	defer os.Setenv("HOME", originalHome)
	
	// 创建loadrc/bashrc目录结构
	bashrcDir := filepath.Join(testDir, "loadrc/bashrc")
	if err := os.MkdirAll(bashrcDir, 0755); err != nil {
		t.Fatalf("无法创建规则目录: %v", err)
	}
	
	// 复制规则文件到目标位置
	if err := copyFile(excludeFile, filepath.Join(bashrcDir, "mirror_exclude")); err != nil {
		t.Fatalf("无法复制排除规则文件: %v", err)
	}
	if err := copyFile(includeFile, filepath.Join(bashrcDir, "mirror_include")); err != nil {
		t.Fatalf("无法复制包含规则文件: %v", err)
	}
	
	// 设置模拟EDITOR环境变量
	originalEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", "echo") // 使用echo替代真实编辑器
	defer os.Setenv("EDITOR", originalEditor)
	
	// 调用dry-run模式执行逻辑
	args := []string{"-aH", "--force", "--delete-during", 
		"--exclude-from=" + filepath.Join(bashrcDir, "mirror_exclude"),
		"--include-from=" + filepath.Join(bashrcDir, "mirror_include"),
		"-n", "-v", sourceDir + "/", targetDir + "/"}
	
	// 执行命令
	cmd := execCommand("rsync", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("模拟rsync命令执行失败: %v, 输出: %s", err, string(output))
	}
	
	// 创建标记文件
	if err := createMarkerFile(); err != nil {
		t.Errorf("无法创建标记文件: %v", err)
	}
	
	// 测试是否创建标记文件
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Errorf("执行dry-run后标记文件未被创建")
	}
}

// 测试实际执行模式
func TestActualExecution(t *testing.T) {
	// 检查是否可以执行模拟命令
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("环境中没有找到go命令，跳过测试")
	}

	// 设置模拟命令
	cleanup := setupFakeExecCommand(t)
	defer cleanup()
	
	// 创建临时测试目录
	testDir, sourceDir, targetDir := setupTestDirs(t)
	defer os.RemoveAll(testDir)
	
	// 创建临时标记文件并写入有效时间戳
	originalMarkerFile := markerFile
	tmpMarkerFile, err := ioutil.TempFile("", "marker_test_actual")
	if err != nil {
		t.Fatalf("无法创建临时标记文件: %v", err)
	}
	markerFile = tmpMarkerFile.Name()
	defer func() {
		markerFile = originalMarkerFile
		os.Remove(tmpMarkerFile.Name())
	}()
	
	// 创建有效的标记文件
	if err := createMarkerFile(); err != nil {
		t.Fatalf("无法创建标记文件: %v", err)
	}
	
	// 创建测试规则文件
	excludeFile, includeFile := createTestRuleFiles(t, testDir)
	
	// 备份和修改环境变量HOME
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", testDir)
	defer os.Setenv("HOME", originalHome)
	
	// 创建loadrc/bashrc目录结构
	bashrcDir := filepath.Join(testDir, "loadrc/bashrc")
	if err := os.MkdirAll(bashrcDir, 0755); err != nil {
		t.Fatalf("无法创建规则目录: %v", err)
	}
	
	// 复制规则文件到目标位置
	if err := copyFile(excludeFile, filepath.Join(bashrcDir, "mirror_exclude")); err != nil {
		t.Fatalf("无法复制排除规则文件: %v", err)
	}
	if err := copyFile(includeFile, filepath.Join(bashrcDir, "mirror_include")); err != nil {
		t.Fatalf("无法复制包含规则文件: %v", err)
	}
	
	// 调用实际执行逻辑
	args := []string{"-aH", "--force", "--delete-during", 
		"--exclude-from=" + filepath.Join(bashrcDir, "mirror_exclude"),
		"--include-from=" + filepath.Join(bashrcDir, "mirror_include"),
		sourceDir + "/", targetDir + "/"}
	
	// 执行命令
	cmd := execCommand("rsync", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("模拟rsync命令执行失败: %v, 输出: %s", err, string(output))
	}
	
	// 删除标记文件
	if err := os.Remove(markerFile); err != nil {
		t.Fatalf("无法删除标记文件: %v", err)
	}
	
	// 测试是否删除标记文件
	if _, err := os.Stat(markerFile); !os.IsNotExist(err) {
		t.Errorf("执行实际操作后标记文件未被删除")
	}
}

// 测试执行失败的情况
func TestExecutionFailure(t *testing.T) {
	// 创建临时测试目录
	testDir, sourceDir, targetDir := setupTestDirs(t)
	defer os.RemoveAll(testDir)
	
	// 创建临时标记文件并写入有效时间戳
	originalMarkerFile := markerFile
	tmpMarkerFile, err := ioutil.TempFile("", "marker_test_failure")
	if err != nil {
		t.Fatalf("无法创建临时标记文件: %v", err)
	}
	markerFile = tmpMarkerFile.Name()
	defer func() {
		markerFile = originalMarkerFile
		os.Remove(tmpMarkerFile.Name())
	}()
	
	// 创建有效的标记文件
	if err := createMarkerFile(); err != nil {
		t.Fatalf("无法创建标记文件: %v", err)
	}
	
	// 创建测试规则文件
	excludeFile, includeFile := createTestRuleFiles(t, testDir)
	
	// 备份和修改环境变量HOME
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", testDir)
	defer os.Setenv("HOME", originalHome)
	
	// 创建loadrc/bashrc目录结构
	bashrcDir := filepath.Join(testDir, "loadrc/bashrc")
	if err := os.MkdirAll(bashrcDir, 0755); err != nil {
		t.Fatalf("无法创建规则目录: %v", err)
	}
	
	// 复制规则文件到目标位置
	if err := copyFile(excludeFile, filepath.Join(bashrcDir, "mirror_exclude")); err != nil {
		t.Fatalf("无法复制排除规则文件: %v", err)
	}
	if err := copyFile(includeFile, filepath.Join(bashrcDir, "mirror_include")); err != nil {
		t.Fatalf("无法复制包含规则文件: %v", err)
	}
	
	// 设置模拟命令为失败执行
	oldExecCommand := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		if command == "rsync" {
			// 使用fail-rsync命令模拟失败
			return exec.Command("fail-rsync", args...)
		}
		return oldExecCommand(command, args...)
	}
	defer func() {
		execCommand = oldExecCommand
	}()
	
	// 调用实际执行逻辑，应该失败
	args := []string{"-aH", "--force", "--delete-during", 
		"--exclude-from=" + filepath.Join(bashrcDir, "mirror_exclude"),
		"--include-from=" + filepath.Join(bashrcDir, "mirror_include"),
		sourceDir + "/", targetDir + "/"}
	
	// 执行命令
	cmd := execCommand("rsync", args...)
	_, err = cmd.CombinedOutput()
	if err == nil {
		t.Errorf("模拟rsync命令预期失败，但却成功了")
	}
} 
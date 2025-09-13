package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mockExecCommand 模拟rsync命令
func mockExecCommand(command string, args ...string) *exec.Cmd {
	// 返回一个成功的echo命令
	return exec.Command("echo", "mock rsync output")
}

// TestSafetyMechanisms 测试所有安全机制
func TestSafetyMechanisms(t *testing.T) {
	t.Run("标记文件不存在时拒绝执行", testNoMarkerFileRefusal)
	t.Run("无效标记文件被拒绝", testInvalidMarkerFileRefusal)
	t.Run("过期标记文件被拒绝", testExpiredMarkerFileRefusal)
	t.Run("空源目录被拒绝", testEmptySourceRefusal)
	t.Run("标记文件执行后自动删除", testMarkerFileAutoDelete)
	t.Run("DryRun不删除文件", testDryRunNoDelete)
	t.Run("标记文件在项目根目录", testMarkerFileLocation)
}

// 测试没有标记文件时拒绝执行
func testNoMarkerFileRefusal(t *testing.T) {
	// 创建测试目录
	testDir, err := ioutil.TempDir("", "safety_test_no_marker")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	sourceDir := filepath.Join(testDir, "source")
	targetDir := filepath.Join(testDir, "target")
	os.MkdirAll(sourceDir, 0755)
	os.MkdirAll(targetDir, 0755)
	
	// 添加测试文件
	ioutil.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644)

	// 确保没有标记文件
	markerPath := ".folder_mirror_marker"
	os.Remove(markerPath)

	// 检查标记文件应该返回错误
	valid, err := checkMarkerFile(sourceDir)
	if valid {
		t.Error("没有标记文件时不应该返回valid=true")
	}
	if err == nil {
		t.Error("没有标记文件时应该返回错误")
	}
	if !strings.Contains(err.Error(), "找不到标记文件") {
		t.Errorf("错误信息不正确: %v", err)
	}
}

// 测试无效标记文件被拒绝
func testInvalidMarkerFileRefusal(t *testing.T) {
	// 创建测试目录
	testDir, err := ioutil.TempDir("", "safety_test_invalid")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	sourceDir := filepath.Join(testDir, "source")
	os.MkdirAll(sourceDir, 0755)
	ioutil.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644)

	// 创建无效的标记文件
	markerPath := ".folder_mirror_marker"
	ioutil.WriteFile(markerPath, []byte("invalid_content"), 0644)

	// 检查应该拒绝
	valid, err := checkMarkerFile(sourceDir)
	if valid {
		t.Error("无效标记文件不应该返回valid=true")
	}
	if err == nil {
		t.Error("无效标记文件应该返回错误")
	}
	if !strings.Contains(err.Error(), "无法解析标记文件") && !strings.Contains(err.Error(), "时间戳") {
		t.Errorf("错误信息不正确，期望包含'无法解析标记文件'和'时间戳': %v", err)
	}
}

// 测试过期标记文件被拒绝
func testExpiredMarkerFileRefusal(t *testing.T) {
	// 创建测试目录
	testDir, err := ioutil.TempDir("", "safety_test_expired")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	sourceDir := filepath.Join(testDir, "source")
	os.MkdirAll(sourceDir, 0755)
	ioutil.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644)

	// 创建过期的标记文件（2小时前）
	markerPath := ".folder_mirror_marker"
	expiredTime := time.Now().Unix() - 7200
	ioutil.WriteFile(markerPath, []byte(fmt.Sprintf("%d", expiredTime)), 0644)

	// 检查应该拒绝
	valid, err := checkMarkerFile(sourceDir)
	if valid {
		t.Error("过期标记文件不应该返回valid=true")
	}
	if err == nil {
		t.Error("过期标记文件应该返回错误")
	}
	if !strings.Contains(err.Error(), "太旧") {
		t.Errorf("错误信息不正确，期望包含'太旧': %v", err)
	}
}

// 测试空源目录被拒绝
func testEmptySourceRefusal(t *testing.T) {
	// 创建空目录
	emptyDir, err := ioutil.TempDir("", "safety_test_empty")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(emptyDir)

	// 检查空目录
	empty, err := isDirEmpty(emptyDir)
	if err != nil {
		t.Fatalf("检查空目录失败: %v", err)
	}
	if !empty {
		t.Error("空目录应该返回empty=true")
	}
}

// 测试标记文件执行后自动删除
func testMarkerFileAutoDelete(t *testing.T) {
	// 创建测试目录
	testDir, err := ioutil.TempDir("", "safety_test_auto_delete")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	sourceDir := filepath.Join(testDir, "source")
	targetDir := filepath.Join(testDir, "target")
	os.MkdirAll(sourceDir, 0755)
	os.MkdirAll(targetDir, 0755)
	ioutil.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644)

	// 保存原始函数
	oldOsExit := osExit
	oldExecCommand := execCommand
	defer func() {
		osExit = oldOsExit
		execCommand = oldExecCommand
	}()

	// Mock osExit
	osExit = func(code int) {
		panic("expected exit")
	}

	// Mock execCommand
	execCommand = mockExecCommand

	// 创建有效的标记文件
	err = createMarkerFile(sourceDir)
	if err != nil {
		t.Fatalf("创建标记文件失败: %v", err)
	}

	// 验证标记文件存在
	markerPath := ".folder_mirror_marker"
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Fatal("标记文件应该存在")
	}

	// 执行实际运行（会触发删除）
	args := []string{"-aH", "--force", "--delete-during"}
	
	func() {
		defer func() {
			recover() // 捕获预期的panic
		}()
		handleActualRun(args, sourceDir+"/", targetDir+"/")
	}()

	// 验证标记文件被删除
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Error("标记文件应该被删除")
	}
}

// 测试DryRun不删除文件
func testDryRunNoDelete(t *testing.T) {
	// 创建测试目录
	testDir, err := ioutil.TempDir("", "safety_test_dryrun")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	sourceDir := filepath.Join(testDir, "source")
	targetDir := filepath.Join(testDir, "target")
	os.MkdirAll(sourceDir, 0755)
	os.MkdirAll(targetDir, 0755)
	
	// 在源和目标创建文件
	sourceFile := filepath.Join(sourceDir, "keep.txt")
	targetFile := filepath.Join(targetDir, "delete.txt")
	ioutil.WriteFile(sourceFile, []byte("source"), 0644)
	ioutil.WriteFile(targetFile, []byte("target"), 0644)

	// 保存原始函数
	oldOsExit := osExit
	oldExecCommand := execCommand
	defer func() {
		osExit = oldOsExit
		execCommand = oldExecCommand
	}()

	// Mock osExit
	osExit = func(code int) {
		panic("expected exit")
	}

	// Mock execCommand 返回成功
	execCommand = mockExecCommand

	// 执行dry run
	args := []string{"-aH", "--force", "--delete-during"}
	
	func() {
		defer func() {
			recover() // 捕获预期的panic
		}()
		handleDryRun(args, sourceDir+"/", targetDir+"/")
	}()

	// 验证目标文件没有被删除（dry-run不应该真的删除）
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Error("DryRun不应该删除目标文件")
	}

	// 验证标记文件被创建
	markerPath := ".folder_mirror_marker"
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("DryRun应该创建标记文件")
	}
}

// 测试标记文件位置正确性
func testMarkerFileLocation(t *testing.T) {
	// 创建测试目录
	testDir, err := ioutil.TempDir("", "safety_test_location")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	sourceDir := filepath.Join(testDir, "source")
	os.MkdirAll(sourceDir, 0755)
	
	// 在源目录添加文件
	ioutil.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644)

	// 清理可能存在的标记文件
	markerPath := ".folder_mirror_marker"
	defer os.Remove(markerPath)

	// 创建标记文件（应该在项目根目录）
	err = createMarkerFile(sourceDir)
	if err != nil {
		t.Fatalf("创建标记文件失败: %v", err)
	}

	// 验证标记文件在项目根目录而不是源目录
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("标记文件应该在项目根目录")
	}
	
	sourceMarkerPath := filepath.Join(sourceDir, ".folder_mirror_marker")
	if _, err := os.Stat(sourceMarkerPath); !os.IsNotExist(err) {
		t.Error("标记文件不应该在源目录")
	}

	// 检查标记文件（应该有效）
	valid, err := checkMarkerFile(sourceDir)
	if err != nil {
		t.Errorf("有效的标记文件不应该返回错误: %v", err)
	}
	if !valid {
		t.Error("标记文件应该有效")
	}
}


// TestCompleteWorkflow 测试完整的工作流程
func TestCompleteWorkflow(t *testing.T) {
	// 创建测试环境
	testDir, err := ioutil.TempDir("", "workflow_test")
	if err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)

	sourceDir := filepath.Join(testDir, "source")
	targetDir := filepath.Join(testDir, "target")
	os.MkdirAll(sourceDir, 0755)
	os.MkdirAll(targetDir, 0755)
	
	// 添加测试文件
	ioutil.WriteFile(filepath.Join(sourceDir, "keep.txt"), []byte("keep"), 0644)
	ioutil.WriteFile(filepath.Join(targetDir, "delete.txt"), []byte("delete"), 0644)

	// 保存原始函数
	oldOsExit := osExit
	oldExecCommand := execCommand
	defer func() {
		osExit = oldOsExit
		execCommand = oldExecCommand
	}()

	// Mock函数
	osExit = func(code int) {
		panic(fmt.Sprintf("exit:%d", code))
	}
	execCommand = mockExecCommand

	// 步骤1: 尝试直接运行（应该失败）
	t.Run("步骤1-无标记文件拒绝", func(t *testing.T) {
		valid, err := checkMarkerFile(sourceDir)
		if valid || err == nil {
			t.Error("无标记文件应该拒绝执行")
		}
	})

	// 步骤2: 执行dry-run
	t.Run("步骤2-执行DryRun", func(t *testing.T) {
		args := []string{"-aH", "--force", "--delete-during"}
		func() {
			defer func() {
				if r := recover(); r != nil {
					if r != "exit:0" {
						t.Errorf("DryRun应该以exit:0结束，实际: %v", r)
					}
				}
			}()
			handleDryRun(args, sourceDir+"/", targetDir+"/")
		}()
		
		// 验证标记文件创建
		markerPath := ".folder_mirror_marker"
		if _, err := os.Stat(markerPath); os.IsNotExist(err) {
			t.Error("DryRun应该创建标记文件")
		}
	})

	// 步骤3: 执行实际运行
	t.Run("步骤3-执行实际运行", func(t *testing.T) {
		// 先确保有标记文件
		err := createMarkerFile(sourceDir)
		if err != nil {
			t.Fatalf("创建标记文件失败: %v", err)
		}
		
		args := []string{"-aH", "--force", "--delete-during"}
		func() {
			defer func() {
				if r := recover(); r != nil {
					if r != "exit:0" {
						t.Errorf("实际运行应该以exit:0结束，实际: %v", r)
					}
				}
			}()
			handleActualRun(args, sourceDir+"/", targetDir+"/")
		}()
		
		// 验证标记文件被删除
		markerPath := ".folder_mirror_marker"
		if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
			t.Error("实际运行后标记文件应该被删除")
		}
	})

	// 步骤4: 再次尝试运行（应该失败）
	t.Run("步骤4-第二次运行拒绝", func(t *testing.T) {
		valid, err := checkMarkerFile(sourceDir)
		if valid || err == nil {
			t.Error("第二次运行应该被拒绝（无标记文件）")
		}
	})
}
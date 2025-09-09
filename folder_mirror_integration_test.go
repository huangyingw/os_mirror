package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// 集成测试：完整的dry run流程
func TestIntegrationDryRunFlow(t *testing.T) {
	// 保存原始设置
	oldOsExit := osExit
	oldPrintHook := printHook
	oldDisablePrint := disablePrint
	defer func() {
		osExit = oldOsExit
		printHook = oldPrintHook
		disablePrint = oldDisablePrint
	}()

	// 禁用打印
	disablePrint = true
	
	// 创建测试目录
	testDir := "/tmp/integration_test_" + strings.ReplaceAll(t.Name(), "/", "_")
	sourceDir := filepath.Join(testDir, "source")
	targetDir := filepath.Join(testDir, "target")
	
	// 清理并创建目录
	os.RemoveAll(testDir)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("创建源目录失败: %v", err)
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("创建目标目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)
	
	// 创建测试文件
	testFiles := []string{"file1.txt", "file2.go", "subdir/file3.md"}
	for _, file := range testFiles {
		fullPath := filepath.Join(sourceDir, file)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("创建子目录失败: %v", err)
		}
		if err := ioutil.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
	}
	
	// 防止osExit
	exitCalled := false
	exitCode := -1
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
		panic("expected exit") // 使用panic停止执行
	}
	
	// 验证路径并准备
	src, tgt := validateAndPreparePaths(sourceDir, targetDir)
	if !strings.HasSuffix(src, "/") {
		t.Error("源路径应以/结尾")
	}
	if !strings.HasSuffix(tgt, "/") {
		t.Error("目标路径应以/结尾")
	}
	
	// 准备rsync参数
	args := prepareRsyncArgs(strings.TrimSuffix(src, "/"))
	if len(args) < 4 {
		t.Error("rsync参数不足")
	}
	
	// 执行dry run
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望panic但未发生")
			}
		}()
		handleDryRun(args, src, tgt)
	}()
	
	// 验证结果
	if !exitCalled || exitCode != 0 {
		t.Errorf("期望osExit(0)，但得到exitCode=%d", exitCode)
	}
	
	// 验证标记文件被创建
	markerPath := filepath.Join(strings.TrimSuffix(src, "/"), ".folder_mirror_marker")
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("标记文件应被创建")
	}
	
	// 验证日志文件被创建
	logPath := filepath.Join(strings.TrimSuffix(src, "/"), ".folder_mirror.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("日志文件应被创建")
	}
}

// 集成测试：完整的实际执行流程
func TestIntegrationActualRunFlow(t *testing.T) {
	// 保存原始设置
	oldOsExit := osExit
	oldMarkerTimeout := markerTimeout
	oldDisablePrint := disablePrint
	defer func() {
		osExit = oldOsExit
		markerTimeout = oldMarkerTimeout
		disablePrint = oldDisablePrint
	}()

	// 禁用打印
	disablePrint = true
	
	// 创建测试目录
	testDir := "/tmp/integration_actual_" + strings.ReplaceAll(t.Name(), "/", "_")
	sourceDir := filepath.Join(testDir, "source")
	targetDir := filepath.Join(testDir, "target")
	
	// 清理并创建目录
	os.RemoveAll(testDir)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("创建源目录失败: %v", err)
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("创建目标目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)
	
	// 创建测试文件
	if err := ioutil.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	
	// 创建有效的标记文件
	if err := createMarkerFile(sourceDir); err != nil {
		t.Fatalf("创建标记文件失败: %v", err)
	}
	
	// 防止osExit
	exitCalled := false
	exitCode := -1
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
	}
	
	// 准备参数
	src := sourceDir + "/"
	tgt := targetDir + "/"
	args := []string{"-aH", "--force", "--delete-during", "--progress"}
	
	// 执行实际运行
	handleActualRun(args, src, tgt)
	
	// 验证结果
	if !exitCalled || exitCode != 0 {
		t.Errorf("期望osExit(0)，但得到exitCode=%d", exitCode)
	}
	
	// 验证标记文件被删除
	markerPath := filepath.Join(strings.TrimSuffix(src, "/"), ".folder_mirror_marker")
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Error("标记文件应被删除")
	}
}

// 测试标记文件超时
func TestMarkerFileTimeout(t *testing.T) {
	// 保存原始设置
	oldMarkerTimeout := markerTimeout
	defer func() {
		markerTimeout = oldMarkerTimeout
	}()
	
	// 创建测试目录
	testDir := "/tmp/marker_timeout_test_" + strings.ReplaceAll(t.Name(), "/", "_")
	sourceDir := filepath.Join(testDir, "source")
	
	// 清理并创建目录
	os.RemoveAll(testDir)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("创建源目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)
	
	// 创建过期的标记文件（2小时前）
	markerPath := filepath.Join(sourceDir, ".folder_mirror_marker")
	oldTimestamp := time.Now().Unix() - 7200
	timestampStr := strings.TrimSpace(strconv.FormatInt(oldTimestamp, 10))
	if err := ioutil.WriteFile(markerPath, []byte(timestampStr), 0644); err != nil {
		t.Fatalf("创建标记文件失败: %v", err)
	}
	
	// 检查标记文件
	valid, err := checkMarkerFile(sourceDir)
	if valid {
		t.Error("过期的标记文件不应该有效")
	}
	if err == nil || !strings.Contains(err.Error(), "太旧") {
		t.Errorf("应返回标记文件太旧的错误，但得到: %v", err)
	}
}

// 测试错误处理：源目录与目标目录嵌套
func TestNestedDirectoryError(t *testing.T) {
	// 保存原始设置
	oldOsExit := osExit
	oldDisablePrint := disablePrint
	defer func() {
		osExit = oldOsExit
		disablePrint = oldDisablePrint
	}()
	
	// 禁用打印
	disablePrint = true
	
	// 创建测试目录
	testDir := "/tmp/nested_test_" + strings.ReplaceAll(t.Name(), "/", "_")
	sourceDir := filepath.Join(testDir, "source")
	targetDir := filepath.Join(sourceDir, "target") // 目标是源的子目录
	
	// 清理并创建目录
	os.RemoveAll(testDir)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("创建源目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)
	
	// 创建测试文件
	if err := ioutil.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	
	// 防止osExit
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		panic("expected exit") // 使用panic停止执行
	}
	
	// 尝试验证嵌套路径
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望panic但未发生")
			}
		}()
		
		validateAndPreparePaths(sourceDir, targetDir)
	}()
	
	if !exitCalled {
		t.Error("期望调用osExit")
	}
}

// 测试readRuleFile错误处理
func TestReadRuleFileError(t *testing.T) {
	// 尝试读取不存在的文件
	rules, err := readRuleFile("/nonexistent/file")
	if err == nil {
		t.Error("应返回错误")
	}
	if rules != nil {
		t.Error("规则列表应为nil")
	}
}

// 测试符号链接检查
func TestSymlinkValidation(t *testing.T) {
	// 创建测试目录
	testDir := "/tmp/symlink_test_" + strings.ReplaceAll(t.Name(), "/", "_")
	realDir := filepath.Join(testDir, "real")
	linkDir := filepath.Join(testDir, "link")
	otherDir := filepath.Join(testDir, "other")
	
	// 清理并创建目录
	os.RemoveAll(testDir)
	if err := os.MkdirAll(realDir, 0755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	defer os.RemoveAll(testDir)
	
	// 创建符号链接
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatalf("创建符号链接失败: %v", err)
	}
	
	// 测试真实目录和符号链接
	same, err := checkDirSameOrNested(realDir, linkDir)
	if err != nil {
		t.Errorf("不期望的错误: %v", err)
	}
	if !same {
		t.Error("真实目录和其符号链接应被识别为相同")
	}
	
	// 测试符号链接和其他目录
	same, err = checkDirSameOrNested(linkDir, otherDir)
	if err != nil {
		t.Errorf("不期望的错误: %v", err)
	}
	if same {
		t.Error("符号链接和其他目录不应被识别为相同")
	}
}
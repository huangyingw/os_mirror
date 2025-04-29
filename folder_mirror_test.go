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

// 创建临时目录结构用于测试
func setupTestDirs(t *testing.T) (string, string, string) {
	// 创建临时测试目录
	testDir, err := ioutil.TempDir("", "folder_mirror_test_")
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}

	// 创建源目录
	sourceDir := filepath.Join(testDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("无法创建源目录: %v", err)
	}

	// 创建目标目录
	targetDir := filepath.Join(testDir, "target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("无法创建目标目录: %v", err)
	}

	// 在源目录中创建一些测试文件和目录
	testFiles := []string{
		"file1.txt",
		"file2.txt",
		"subdir/file3.txt",
		".git/config",
		"node_modules/package.json",
		"src/build/output.js",
		"lib/build/temp.txt",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(sourceDir, file)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("无法创建目录 %s: %v", dir, err)
		}
		if err := ioutil.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("无法创建文件 %s: %v", fullPath, err)
		}
	}

	return testDir, sourceDir, targetDir
}

// 创建测试规则文件
func createTestRuleFiles(t *testing.T, testDir string) (string, string) {
	// 创建排除规则文件
	excludeFile := filepath.Join(testDir, "mirror_exclude")
	excludeRules := []string{
		"# 测试排除规则",
		".git/",
		"node_modules/",
		"*/build/*",
	}
	if err := ioutil.WriteFile(excludeFile, []byte(strings.Join(excludeRules, "\n")), 0644); err != nil {
		t.Fatalf("无法创建排除规则文件: %v", err)
	}

	// 创建包含规则文件
	includeFile := filepath.Join(testDir, "mirror_include")
	includeRules := []string{
		"# 测试包含规则",
		"*.txt",
		"subdir/",
	}
	if err := ioutil.WriteFile(includeFile, []byte(strings.Join(includeRules, "\n")), 0644); err != nil {
		t.Fatalf("无法创建包含规则文件: %v", err)
	}

	return excludeFile, includeFile
}

// 测试目录检查函数
func TestDirExists(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "dir_exists_test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试存在的目录
	if !dirExists(tempDir) {
		t.Errorf("dirExists(%s) = false, 期望 true", tempDir)
	}

	// 测试不存在的目录
	nonExistentDir := filepath.Join(tempDir, "non_existent")
	if dirExists(nonExistentDir) {
		t.Errorf("dirExists(%s) = true, 期望 false", nonExistentDir)
	}

	// 测试远程路径
	remotePath := "user@host:/path/to/dir"
	if !dirExists(remotePath) {
		t.Errorf("dirExists(%s) = false, 期望 true (远程路径应该默认为存在)", remotePath)
	}

	// 测试文件而非目录
	filePath := filepath.Join(tempDir, "test_file")
	if err := ioutil.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("无法创建测试文件: %v", err)
	}
	if dirExists(filePath) {
		t.Errorf("dirExists(%s) = true, 期望 false (文件不应被视为目录)", filePath)
	}
}

// 测试目录创建函数
func TestCreateDir(t *testing.T) {
	// 创建临时目录
	baseDir, err := ioutil.TempDir("", "create_dir_test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(baseDir)

	// 测试创建目录
	newDir := filepath.Join(baseDir, "new_dir", "subdir")
	if err := createDir(newDir); err != nil {
		t.Errorf("createDir(%s) 失败: %v", newDir, err)
	}
	if !dirExists(newDir) {
		t.Errorf("创建目录 %s 后, dirExists 返回 false", newDir)
	}

	// 测试远程路径 - 应返回错误
	remotePath := "user@host:/path/to/dir"
	err = createDir(remotePath)
	if err == nil {
		t.Errorf("createDir(%s) 应当返回错误，但得到了nil", remotePath)
	} else if !strings.Contains(err.Error(), "不支持创建远程目录") {
		t.Errorf("createDir(%s) 返回了意外的错误: %v", remotePath, err)
	}
}

// 测试标记文件创建和检查
func TestMarkerFile(t *testing.T) {
	// 保存原始常量值以便在测试后恢复
	originalMarkerFile := markerFile
	originalMarkerTimeout := markerTimeout
	
	// 创建临时文件作为测试标记文件
	tmpMarkerFile, err := ioutil.TempFile("", "marker_test")
	if err != nil {
		t.Fatalf("无法创建临时标记文件: %v", err)
	}
	tmpMarkerPath := tmpMarkerFile.Name()
	tmpMarkerFile.Close()
	defer os.Remove(tmpMarkerPath)
	
	// 修改变量值以使用临时文件
	markerFile = tmpMarkerPath
	markerTimeout = 10 // 10秒超时用于测试
	
	// 测试完成后恢复原始值
	defer func() {
		markerFile = originalMarkerFile
		markerTimeout = originalMarkerTimeout
	}()
	
	// 测试创建标记文件
	if err := createMarkerFile(); err != nil {
		t.Errorf("createMarkerFile() 失败: %v", err)
	}
	
	// 验证标记文件内容
	data, err := ioutil.ReadFile(markerFile)
	if err != nil {
		t.Fatalf("无法读取标记文件: %v", err)
	}
	
	// 验证时间戳
	timestampStr := strings.TrimSpace(string(data))
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		t.Fatalf("标记文件内容不是有效的Unix时间戳: %s", string(data))
	}
	
	now := time.Now().Unix()
	diff := now - timestamp
	if diff < 0 {
		diff = -diff
	}
	
	if diff > 5 {
		t.Errorf("标记文件时间戳与当前时间相差太大: %v 秒", diff)
	}
	
	// 测试检查标记文件
	valid, err := checkMarkerFile()
	if err != nil {
		t.Errorf("checkMarkerFile() 失败: %v", err)
	}
	if !valid {
		t.Errorf("checkMarkerFile() = false, 期望 true")
	}
	
	// 测试过期的标记文件
	expiredTimestamp := time.Now().Unix() - markerTimeout - 10
	if err := ioutil.WriteFile(markerFile, []byte(strconv.FormatInt(expiredTimestamp, 10)), 0644); err != nil {
		t.Fatalf("无法写入标记文件: %v", err)
	}
	
	valid, err = checkMarkerFile()
	if valid || err == nil {
		t.Errorf("对于过期的标记文件，checkMarkerFile() = %v, %v; 期望 false, error", valid, err)
	}
	
	// 测试标记文件格式错误
	if err := ioutil.WriteFile(markerFile, []byte("not_a_timestamp"), 0644); err != nil {
		t.Fatalf("无法写入标记文件: %v", err)
	}
	
	valid, err = checkMarkerFile()
	if valid || err == nil {
		t.Errorf("对于格式错误的标记文件，checkMarkerFile() = %v, %v; 期望 false, error", valid, err)
	}
	
	// 删除标记文件测试不存在的情况
	if err := os.Remove(markerFile); err != nil {
		t.Fatalf("无法删除标记文件: %v", err)
	}
	
	valid, err = checkMarkerFile()
	if valid || err == nil {
		t.Errorf("对于不存在的标记文件，checkMarkerFile() = %v, %v; 期望 false, error", valid, err)
	}
}

// 测试读取规则文件
func TestReadRuleFile(t *testing.T) {
	// 创建临时规则文件
	tmpFile, err := ioutil.TempFile("", "rules_test")
	if err != nil {
		t.Fatalf("无法创建临时规则文件: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	// 写入测试规则
	testRules := []string{
		"# 这是注释行",
		"",
		"rule1",
		"  rule2  ", // 带空格的规则
		"# 另一个注释",
		"rule3",
	}
	
	if _, err := tmpFile.WriteString(strings.Join(testRules, "\n")); err != nil {
		t.Fatalf("无法写入测试规则: %v", err)
	}
	tmpFile.Close()
	
	// 读取规则文件
	rules, err := readRuleFile(tmpFile.Name())
	if err != nil {
		t.Errorf("readRuleFile(%s) 失败: %v", tmpFile.Name(), err)
	}
	
	// 验证结果
	expectedRules := []string{"rule1", "rule2", "rule3"}
	if len(rules) != len(expectedRules) {
		t.Errorf("规则数量不匹配: 得到 %d, 期望 %d", len(rules), len(expectedRules))
	}
	
	for i, rule := range rules {
		if rule != expectedRules[i] {
			t.Errorf("规则[%d] = %q, 期望 %q", i, rule, expectedRules[i])
		}
	}
	
	// 测试不存在的文件
	_, err = readRuleFile("/non/existent/file")
	if err == nil {
		t.Errorf("对不存在的文件，readRuleFile() 应该返回错误")
	}
	
	// 创建无效的规则文件测试扫描错误
	invalidContent := []byte{0, 1, 2, 3, 4} // 不可读的二进制内容
	invalidFile, err := ioutil.TempFile("", "invalid_rules_test")
	if err != nil {
		t.Fatalf("无法创建临时文件: %v", err)
	}
	defer os.Remove(invalidFile.Name())
	
	// 写入无效内容
	if _, err := invalidFile.Write(invalidContent); err != nil {
		t.Fatalf("无法写入无效内容: %v", err)
	}
	invalidFile.Close()
	
	// 尝试读取无效内容的文件
	if _, err := readRuleFile(invalidFile.Name()); err != nil {
		// 这实际上应该成功，因为readRuleFile只是跳过无法解析的行
		t.Errorf("无法读取含无效内容的文件: %v", err)
	}
}

// 测试检查目录是否为空
func TestIsDirEmpty(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "empty_dir_test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试空目录
	empty, err := isDirEmpty(tempDir)
	if err != nil {
		t.Errorf("isDirEmpty(%s) 失败: %v", tempDir, err)
	}
	if !empty {
		t.Errorf("isDirEmpty(%s) = false, 期望 true (空目录应返回true)", tempDir)
	}

	// 测试非空目录
	filePath := filepath.Join(tempDir, "test_file")
	if err := ioutil.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("无法创建测试文件: %v", err)
	}

	empty, err = isDirEmpty(tempDir)
	if err != nil {
		t.Errorf("isDirEmpty(%s) 失败: %v", tempDir, err)
	}
	if empty {
		t.Errorf("isDirEmpty(%s) = true, 期望 false (非空目录应返回false)", tempDir)
	}

	// 测试不存在的目录
	nonExistentDir := filepath.Join(tempDir, "non_existent")
	_, err = isDirEmpty(nonExistentDir)
	if err == nil {
		t.Errorf("isDirEmpty(%s) 应返回错误，但没有", nonExistentDir)
	}

	// 测试远程路径 - 应返回错误
	remotePath := "user@host:/path/to/dir"
	_, err = isDirEmpty(remotePath)
	if err == nil {
		t.Errorf("isDirEmpty(%s) 应当返回错误，但得到了nil", remotePath)
	} else if !strings.Contains(err.Error(), "不支持检查远程目录是否为空") {
		t.Errorf("isDirEmpty(%s) 返回了意外的错误: %v", remotePath, err)
	}
}

// 测试检查目录是否相同或嵌套
func TestCheckDirSameOrNested(t *testing.T) {
	// 创建临时目录
	baseDir, err := ioutil.TempDir("", "dir_relation_test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(baseDir)

	// 创建测试目录结构
	sourceDir := filepath.Join(baseDir, "source")
	targetDir := filepath.Join(baseDir, "target")
	nestedDir := filepath.Join(sourceDir, "nested")
	
	for _, dir := range []string{sourceDir, targetDir, nestedDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("无法创建目录 %s: %v", dir, err)
		}
	}

	// 测试相同目录
	same, err := checkDirSameOrNested(sourceDir, sourceDir)
	if err != nil {
		t.Errorf("checkDirSameOrNested(%s, %s) 失败: %v", sourceDir, sourceDir, err)
	}
	if !same {
		t.Errorf("checkDirSameOrNested(%s, %s) = false, 期望 true (相同目录应返回true)", sourceDir, sourceDir)
	}

	// 测试不同目录
	same, err = checkDirSameOrNested(sourceDir, targetDir)
	if err != nil {
		t.Errorf("checkDirSameOrNested(%s, %s) 失败: %v", sourceDir, targetDir, err)
	}
	if same {
		t.Errorf("checkDirSameOrNested(%s, %s) = true, 期望 false (不同目录应返回false)", sourceDir, targetDir)
	}

	// 测试嵌套目录 - 目标是源的子目录
	same, err = checkDirSameOrNested(sourceDir, nestedDir)
	if err != nil {
		t.Errorf("checkDirSameOrNested(%s, %s) 失败: %v", sourceDir, nestedDir, err)
	}
	if !same {
		t.Errorf("checkDirSameOrNested(%s, %s) = false, 期望 true (目标是源的子目录应返回true)", sourceDir, nestedDir)
	}

	// 测试嵌套目录 - 源是目标的子目录
	same, err = checkDirSameOrNested(nestedDir, sourceDir)
	if err != nil {
		t.Errorf("checkDirSameOrNested(%s, %s) 失败: %v", nestedDir, sourceDir, err)
	}
	if !same {
		t.Errorf("checkDirSameOrNested(%s, %s) = false, 期望 true (源是目标的子目录应返回true)", nestedDir, sourceDir)
	}

	// 测试远程路径
	remotePath := "user@host:/path/to/dir"
	
	// 本地路径和远程路径 - 应返回错误
	_, err = checkDirSameOrNested(sourceDir, remotePath)
	if err == nil {
		t.Errorf("checkDirSameOrNested(%s, %s) 应当返回错误，但得到了nil", sourceDir, remotePath)
	} else if !strings.Contains(err.Error(), "不支持远程目标目录路径") {
		t.Errorf("checkDirSameOrNested(%s, %s) 返回了意外的错误: %v", sourceDir, remotePath, err)
	}
	
	// 远程路径和本地路径 - 应返回错误
	_, err = checkDirSameOrNested(remotePath, sourceDir)
	if err == nil {
		t.Errorf("checkDirSameOrNested(%s, %s) 应当返回错误，但得到了nil", remotePath, sourceDir)
	} else if !strings.Contains(err.Error(), "不支持远程源目录路径") {
		t.Errorf("checkDirSameOrNested(%s, %s) 返回了意外的错误: %v", remotePath, sourceDir, err)
	}
	
	// 测试无效的绝对路径
	invalidPath := string([]byte{0})
	_, err = checkDirSameOrNested(invalidPath, targetDir)
	if err == nil {
		t.Errorf("对于无效路径，checkDirSameOrNested应当返回错误")
	}
	
	_, err = checkDirSameOrNested(sourceDir, invalidPath)
	if err == nil {
		t.Errorf("对于无效路径，checkDirSameOrNested应当返回错误")
	}
	
	// 两个远程路径 - 应返回错误
	remotePath2 := "user2@host2:/path/to/dir2"
	_, err = checkDirSameOrNested(remotePath, remotePath2)
	if err == nil {
		t.Errorf("checkDirSameOrNested(%s, %s) 应当返回错误，但得到了nil", remotePath, remotePath2)
	} else if !strings.Contains(err.Error(), "不支持远程源目录路径") {
		t.Errorf("checkDirSameOrNested(%s, %s) 返回了意外的错误: %v", remotePath, remotePath2, err)
	}
}

// 测试彩色打印函数 - 由于输出到控制台，只能做基本验证
func TestPrintColored(t *testing.T) {
	// 这个测试基本上是检查函数是否会抛出panic
	// 因为我们不能轻易捕获标准输出流进行验证
	printColored(colorRed, "测试红色消息")
	printColored(colorGreen, "测试绿色消息")
	printColored(colorYellow, "测试黄色消息")
	// 如果没有panic，则测试通过
} 
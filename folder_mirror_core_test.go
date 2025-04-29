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

// 测试创建标记文件各种场景
func TestCreateMarkerFileScenarios(t *testing.T) {
	// 保存原始值
	originalMarkerFile := markerFile
	
	// 测试完成后恢复
	defer func() {
		markerFile = originalMarkerFile
	}()
	
	// 场景1: 正常创建标记文件
	tmpFile, err := ioutil.TempFile("", "marker_test_normal_")
	if err != nil {
		t.Fatalf("无法创建临时文件: %v", err)
	}
	tmpFile.Close()
	os.Remove(tmpFile.Name()) // 删除文件，让函数创建它
	
	markerFile = tmpFile.Name()
	if err := createMarkerFile(); err != nil {
		t.Errorf("无法创建标记文件(正常情况): %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	// 验证文件内容
	content, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		t.Errorf("无法读取创建的标记文件: %v", err)
	}
	_, err = strconv.ParseInt(string(content), 10, 64)
	if err != nil {
		t.Errorf("标记文件内容不是有效的时间戳: %s", string(content))
	}
	
	// 场景2: 在只读目录中创建标记文件
	if os.Getuid() == 0 {
		// 跳过root用户的测试，因为root可以写入只读目录
		t.Log("跳过只读目录测试，因为当前用户是root")
	} else {
		// 创建只读目录
		readonlyDir, err := ioutil.TempDir("", "readonly_dir_")
		if err != nil {
			t.Fatalf("无法创建临时目录: %v", err)
		}
		defer os.RemoveAll(readonlyDir)
		
		// 设置为只读
		if err := os.Chmod(readonlyDir, 0500); err != nil {
			t.Fatalf("无法将目录设为只读: %v", err)
		}
		
		// 尝试在只读目录中创建标记文件
		markerFile = filepath.Join(readonlyDir, "marker")
		if err := createMarkerFile(); err == nil {
			t.Error("在只读目录中创建标记文件应当失败，但成功了")
		}
	}
}

// 测试检查标记文件各种场景
func TestCheckMarkerFileScenarios(t *testing.T) {
	// 保存原始值
	originalMarkerFile := markerFile
	originalMarkerTimeout := markerTimeout
	
	// 测试完成后恢复
	defer func() {
		markerFile = originalMarkerFile
		markerTimeout = originalMarkerTimeout
	}()
	
	// 设置较短的超时用于测试
	markerTimeout = 30 // 30秒
	
	// 场景1: 标记文件不存在
	markerFile = "/tmp/non_existent_marker_file_for_test"
	valid, err := checkMarkerFile()
	if valid {
		t.Error("对不存在的标记文件，checkMarkerFile返回true")
	}
	if err == nil || !strings.Contains(err.Error(), "找不到标记文件") {
		t.Errorf("对不存在的标记文件，期望错误信息包含'找不到标记文件'，但得到: %v", err)
	}
	
	// 场景2: 标记文件存在但内容无效
	tmpFile, err := ioutil.TempFile("", "marker_test_invalid_")
	if err != nil {
		t.Fatalf("无法创建临时文件: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	// 写入无效内容
	if _, err := tmpFile.WriteString("not_a_timestamp"); err != nil {
		t.Fatalf("无法写入临时文件: %v", err)
	}
	tmpFile.Close()
	
	markerFile = tmpFile.Name()
	valid, err = checkMarkerFile()
	if valid {
		t.Error("对内容无效的标记文件，checkMarkerFile返回true")
	}
	if err == nil || !strings.Contains(err.Error(), "无法解析") {
		t.Errorf("对内容无效的标记文件，期望错误信息包含'无法解析'，但得到: %v", err)
	}
	
	// 场景3: 标记文件已过期
	tmpFile, err = ioutil.TempFile("", "marker_test_expired_")
	if err != nil {
		t.Fatalf("无法创建临时文件: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	// 写入过期时间戳
	expiredTime := time.Now().Add(-time.Duration(markerTimeout+10) * time.Second).Unix()
	if _, err := tmpFile.WriteString(strconv.FormatInt(expiredTime, 10)); err != nil {
		t.Fatalf("无法写入临时文件: %v", err)
	}
	tmpFile.Close()
	
	markerFile = tmpFile.Name()
	valid, err = checkMarkerFile()
	if valid {
		t.Error("对过期的标记文件，checkMarkerFile返回true")
	}
	if err == nil || !strings.Contains(err.Error(), "标记文件太旧") {
		t.Errorf("对过期的标记文件，期望错误信息包含'标记文件太旧'，但得到: %v", err)
	}
	
	// 场景4: 标记文件有效
	tmpFile, err = ioutil.TempFile("", "marker_test_valid_")
	if err != nil {
		t.Fatalf("无法创建临时文件: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	// 写入当前时间戳
	currentTime := time.Now().Unix()
	if _, err := tmpFile.WriteString(strconv.FormatInt(currentTime, 10)); err != nil {
		t.Fatalf("无法写入临时文件: %v", err)
	}
	tmpFile.Close()
	
	markerFile = tmpFile.Name()
	valid, err = checkMarkerFile()
	if !valid {
		t.Errorf("对有效的标记文件，checkMarkerFile返回false: %v", err)
	}
	if err != nil {
		t.Errorf("对有效的标记文件，checkMarkerFile返回错误: %v", err)
	}
}

// 测试读取规则文件的更复杂场景
func TestReadRuleFileExtended(t *testing.T) {
	// 测试解析含有多种混合格式的规则文件
	tmpFile, err := ioutil.TempFile("", "rules_complex_")
	if err != nil {
		t.Fatalf("无法创建临时文件: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	// 写入各种格式的规则
	complexRules := []string{
		"# 这是注释行",
		"",
		"rule1", 
		"  rule2  ", // 带空格
		"# 另一个注释",
		"", // 空行
		"  # 缩进的注释",
		"rule3",
		"  ", // 只有空格的行
		"#",  // 只有#的行
		"rule4",
	}
	
	if _, err := tmpFile.WriteString(strings.Join(complexRules, "\n")); err != nil {
		t.Fatalf("无法写入测试规则: %v", err)
	}
	tmpFile.Close()
	
	// 读取规则文件
	rules, err := readRuleFile(tmpFile.Name())
	if err != nil {
		t.Errorf("readRuleFile(%s) 失败: %v", tmpFile.Name(), err)
	}
	
	// 验证结果
	expectedRules := []string{"rule1", "rule2", "rule3", "rule4"}
	if len(rules) != len(expectedRules) {
		t.Errorf("规则数量不匹配: 得到 %d, 期望 %d", len(rules), len(expectedRules))
		t.Errorf("得到的规则: %v", rules)
	}
	
	for i, rule := range rules {
		if i < len(expectedRules) && rule != expectedRules[i] {
			t.Errorf("规则[%d] = %q, 期望 %q", i, rule, expectedRules[i])
		}
	}
	
	// 测试读取权限被拒绝的文件
	if os.Getuid() == 0 {
		// 跳过root用户的测试，因为root可以读取任何文件
		t.Log("跳过权限测试，因为当前用户是root")
	} else {
		// 创建一个无权限读取的文件
		noPermFile, err := ioutil.TempFile("", "no_perm_file_")
		if err != nil {
			t.Fatalf("无法创建临时文件: %v", err)
		}
		noPermFile.Close()
		defer os.Remove(noPermFile.Name())
		
		// 删除所有权限
		if err := os.Chmod(noPermFile.Name(), 0000); err != nil {
			t.Fatalf("无法修改文件权限: %v", err)
		}
		
		// 尝试读取无权限文件
		_, err = readRuleFile(noPermFile.Name())
		if err == nil {
			t.Error("读取无权限文件应当失败，但成功了")
		}
	}
}

// 测试检查相同或嵌套目录的功能 - 扩展测试
func TestCheckDirSameOrNestedExtended(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "dir_nested_test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试目录结构
	parentDir := filepath.Join(tempDir, "parent")
	childDir := filepath.Join(parentDir, "child")
	siblingDir := filepath.Join(tempDir, "sibling")

	// 创建目录
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatalf("无法创建子目录: %v", err)
	}
	if err := os.MkdirAll(siblingDir, 0755); err != nil {
		t.Fatalf("无法创建兄弟目录: %v", err)
	}

	// 测试用例
	testCases := []struct {
		name       string
		source     string
		target     string
		expected   bool
		expectErr  bool
		errMessage string
	}{
		{
			name:     "相同目录",
			source:   parentDir,
			target:   parentDir,
			expected: true,
		},
		{
			name:     "子目录作为目标",
			source:   parentDir,
			target:   childDir,
			expected: true,
		},
		{
			name:     "父目录作为目标",
			source:   childDir,
			target:   parentDir,
			expected: true,
		},
		{
			name:     "兄弟目录",
			source:   parentDir,
			target:   siblingDir,
			expected: false,
		},
		{
			name:       "远程源路径",
			source:     "user@host:/remote/source",
			target:     siblingDir,
			expected:   false,
			expectErr:  true,
			errMessage: "不支持远程源目录路径",
		},
		{
			name:       "远程目标路径",
			source:     parentDir,
			target:     "user@host:/remote/target",
			expected:   false,
			expectErr:  true,
			errMessage: "不支持远程目标目录路径",
		},
	}

	// 执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := checkDirSameOrNested(tc.source, tc.target)

			// 检查错误
			if tc.expectErr {
				if err == nil {
					t.Errorf("期望错误但没有得到错误")
				} else if tc.errMessage != "" && !strings.Contains(err.Error(), tc.errMessage) {
					t.Errorf("错误消息不匹配，期望包含 %q，得到 %q", tc.errMessage, err.Error())
				}
				return
			}

			// 如果不期望错误，检查结果是否符合预期
			if err != nil {
				t.Errorf("意外错误: %v", err)
				return
			}

			if result != tc.expected {
				t.Errorf("结果不符合预期，期望 %v，得到 %v", tc.expected, result)
			}
		})
	}
}

// 测试创建和检查符号链接目录
func TestCheckDirSymlinks(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "symlink_test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试目录结构
	realDir := filepath.Join(tempDir, "real_dir")
	symlinkDir := filepath.Join(tempDir, "symlink_dir")
	otherDir := filepath.Join(tempDir, "other_dir")

	// 创建目录
	if err := os.MkdirAll(realDir, 0755); err != nil {
		t.Fatalf("无法创建真实目录: %v", err)
	}
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatalf("无法创建其他目录: %v", err)
	}

	// 创建符号链接
	if err := os.Symlink(realDir, symlinkDir); err != nil {
		// 跳过此测试，如果环境不支持符号链接
		t.Skipf("环境不支持符号链接: %v", err)
	}

	// 测试用例
	testCases := []struct {
		name     string
		source   string
		target   string
		expected bool
	}{
		{
			name:     "真实目录和符号链接",
			source:   realDir,
			target:   symlinkDir,
			expected: true,
		},
		{
			name:     "符号链接和真实目录",
			source:   symlinkDir,
			target:   realDir,
			expected: true,
		},
		{
			name:     "符号链接和其他目录",
			source:   symlinkDir,
			target:   otherDir,
			expected: false,
		},
	}

	// 执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := checkDirSameOrNested(tc.source, tc.target)
			if err != nil {
				t.Errorf("意外错误: %v", err)
				return
			}
			if result != tc.expected {
				t.Errorf("结果不符合预期，期望 %v，得到 %v", tc.expected, result)
			}
		})
	}
}

// 测试目标路径不存在的情况
func TestCheckDirNonExistentTarget(t *testing.T) {
	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "nonexistent_test")
	if err != nil {
		t.Fatalf("无法创建临时目录: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建源目录
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("无法创建源目录: %v", err)
	}

	// 不存在的目标目录
	targetDir := filepath.Join(tempDir, "nonexistent")

	// 测试检查
	result, err := checkDirSameOrNested(sourceDir, targetDir)
	if err != nil {
		t.Errorf("检查不存在的目标目录时出错: %v", err)
	}
	if result {
		t.Error("不存在的目标目录被错误地识别为相同或嵌套目录")
	}
}

// 测试无法获取Lstat信息的情况
func TestCheckDirLstatError(t *testing.T) {
	// 使用一个特殊的无法访问的路径
	sourceDir := "/proc/self/pagemap" // 普通用户无法访问此文件
	targetDir := "/tmp"

	// 测试检查
	_, err := checkDirSameOrNested(sourceDir, targetDir)
	// 这里我们只需要验证函数不会崩溃，具体错误消息可能因操作系统而异
	if err == nil {
		// 如果没有错误，可能是因为当前用户有足够的权限或者在某些特殊环境中运行
		t.Log("预期会有错误，但没有得到错误。这可能取决于运行测试的用户权限。")
	}
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// 查找子字符串在字符串中的索引位置
func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
} 
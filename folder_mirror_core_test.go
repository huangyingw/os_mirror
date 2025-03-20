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
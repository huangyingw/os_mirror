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

// 测试创建标记文件的各种场景
func TestCreateMarkerFileScenarios(t *testing.T) {
	// 场景1: 正常创建标记文件
	testSourceDir, err := ioutil.TempDir("", "marker_test_source_")
	if err != nil {
		t.Fatalf("无法创建临时源目录: %v", err)
	}
	defer os.RemoveAll(testSourceDir)
	
	if err := createMarkerFile(testSourceDir); err != nil {
		t.Errorf("无法创建标记文件(正常情况): %v", err)
	}
	
	// 验证文件内容
	markerPath := filepath.Join(testSourceDir, ".folder_mirror_marker")
	content, err := ioutil.ReadFile(markerPath)
	if err != nil {
		t.Errorf("无法读取创建的标记文件: %v", err)
	}
	_, err = strconv.ParseInt(strings.TrimSpace(string(content)), 10, 64)
	if err != nil {
		t.Errorf("标记文件内容不是有效的时间戳: %s", string(content))
	}
	
	// 场景2: 在只读源目录中创建标记文件（应该失败）
	if os.Getuid() != 0 { // 跳过root用户，root可以写入只读目录
		readonlyDir, err := ioutil.TempDir("", "readonly_source_")
		if err != nil {
			t.Fatalf("无法创建只读测试目录: %v", err)
		}
		defer os.RemoveAll(readonlyDir)
		
		// 设置为只读
		if err := os.Chmod(readonlyDir, 0500); err != nil {
			t.Fatalf("无法将目录设为只读: %v", err)
		}
		
		// 尝试在只读目录中创建标记文件
		if err := createMarkerFile(readonlyDir); err == nil {
			t.Error("在只读源目录中创建标记文件应当失败，但成功了")
		}
	}
}

// 测试检查标记文件各种场景
func TestCheckMarkerFileScenarios(t *testing.T) {
	// 保存原始超时设置
	originalMarkerTimeout := markerTimeout
	defer func() {
		markerTimeout = originalMarkerTimeout
	}()
	
	// 设置较短的超时用于测试
	markerTimeout = 30 // 30秒
	
	testSourceDir, err := ioutil.TempDir("", "check_marker_test_")
	if err != nil {
		t.Fatalf("无法创建临时源目录: %v", err)
	}
	defer os.RemoveAll(testSourceDir)
	
	// 场景1: 标记文件不存在
	valid, err := checkMarkerFile(testSourceDir)
	if valid {
		t.Error("对不存在的标记文件，checkMarkerFile返回true")
	}
	if err == nil || !strings.Contains(err.Error(), "找不到标记文件") {
		t.Errorf("对不存在的标记文件，期望错误信息包含'找不到标记文件'，但得到: %v", err)
	}
	
	// 场景2: 标记文件存在但内容无效
	markerPath := filepath.Join(testSourceDir, ".folder_mirror_marker")
	if err := ioutil.WriteFile(markerPath, []byte("not_a_timestamp"), 0644); err != nil {
		t.Fatalf("无法写入无效标记文件: %v", err)
	}
	
	valid, err = checkMarkerFile(testSourceDir)
	if valid {
		t.Error("对内容无效的标记文件，checkMarkerFile返回true")
	}
	if err == nil || !strings.Contains(err.Error(), "无法解析") {
		t.Errorf("对内容无效的标记文件，期望错误信息包含'无法解析'，但得到: %v", err)
	}
	
	// 场景3: 标记文件已过期
	expiredTime := time.Now().Add(-time.Duration(markerTimeout+10) * time.Second).Unix()
	if err := ioutil.WriteFile(markerPath, []byte(strconv.FormatInt(expiredTime, 10)), 0644); err != nil {
		t.Fatalf("无法写入过期标记文件: %v", err)
	}
	
	valid, err = checkMarkerFile(testSourceDir)
	if valid {
		t.Error("对过期的标记文件，checkMarkerFile返回true")
	}
	if err == nil || !strings.Contains(err.Error(), "太旧") {
		t.Errorf("对过期的标记文件，期望错误信息包含'太旧'，但得到: %v", err)
	}
	
	// 场景4: 标记文件有效
	currentTime := time.Now().Unix()
	if err := ioutil.WriteFile(markerPath, []byte(strconv.FormatInt(currentTime, 10)), 0644); err != nil {
		t.Fatalf("无法写入有效标记文件: %v", err)
	}
	
	valid, err = checkMarkerFile(testSourceDir)
	if !valid {
		t.Errorf("对有效的标记文件，checkMarkerFile返回false: %v", err)
	}
	if err != nil {
		t.Errorf("对有效的标记文件，checkMarkerFile返回错误: %v", err)
	}
}
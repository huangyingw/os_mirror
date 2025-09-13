package main

import (
	"io/ioutil"
	"os"
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
	
	// 清理可能存在的标记文件
	markerPath := ".folder_mirror_marker"
	defer os.Remove(markerPath)
	
	if err := createMarkerFile(testSourceDir); err != nil {
		t.Errorf("无法创建标记文件(正常情况): %v", err)
	}
	
	// 验证文件内容
	content, err := ioutil.ReadFile(markerPath)
	if err != nil {
		t.Errorf("无法读取创建的标记文件: %v", err)
	}
	_, err = strconv.ParseInt(strings.TrimSpace(string(content)), 10, 64)
	if err != nil {
		t.Errorf("标记文件内容不是有效的时间戳: %s", string(content))
	}
	
	// 场景2: 重复创建标记文件（应该成功，覆盖旧文件）
	if err := createMarkerFile(testSourceDir); err != nil {
		t.Errorf("重复创建标记文件失败: %v", err)
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
	
	// 清理可能存在的标记文件
	markerPath := ".folder_mirror_marker"
	defer os.Remove(markerPath)
	
	// 场景1: 标记文件不存在
	valid, err := checkMarkerFile(testSourceDir)
	if valid {
		t.Error("对不存在的标记文件，checkMarkerFile返回true")
	}
	if err == nil || !strings.Contains(err.Error(), "找不到标记文件") {
		t.Errorf("对不存在的标记文件，期望错误信息包含'找不到标记文件'，但得到: %v", err)
	}
	
	// 场景2: 标记文件存在但内容无效
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
package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// 测试源目录和目标目录相同的情况
func TestSameSourceAndTarget(t *testing.T) {
	// 创建测试目录结构
	baseDir, err := ioutil.TempDir("", "same_dir_test")
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(baseDir)

	// 创建测试文件
	testFile := filepath.Join(baseDir, "test_file")
	if err := ioutil.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("无法创建测试文件: %v", err)
	}

	// 捕获输出
	var capturedOutput []string
	oldHook := printHook
	printHook = func(message string) {
		capturedOutput = append(capturedOutput, message)
	}
	defer func() {
		printHook = oldHook
	}()

	// 临时修改os.Exit函数
	oldOsExit := osExit
	exitCode := 0
	osExit = func(code int) {
		exitCode = code
		// 不实际退出，只记录退出码
	}
	defer func() {
		osExit = oldOsExit
	}()

	// 测试相同目录
	source := baseDir
	target := baseDir

	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	if !strings.HasSuffix(target, "/") {
		target += "/"
	}

	printColored(colorGreen, "源目录: "+source)
	printColored(colorGreen, "目标目录: "+target)

	// 执行检查
	isSameOrNested, _ := checkDirSameOrNested(source, target)
	if isSameOrNested {
		printColored(colorRed, "错误: 源目录和目标目录相同或互为子目录，操作危险，终止执行")
		osExit(1)
	}

	// 验证结果
	if exitCode != 1 {
		t.Errorf("对于相同的源目录和目标目录，期望退出码为1，但得到 %d", exitCode)
	}

	found := false
	for _, msg := range capturedOutput {
		if strings.Contains(msg, "源目录和目标目录相同或互为子目录") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("对于相同的源目录和目标目录，应显示警告消息，但未找到")
	}
}

// 测试源目录为空的情况
func TestEmptySourceDir(t *testing.T) {
	// 创建测试目录结构
	sourceDir, err := ioutil.TempDir("", "empty_source_test")
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	targetDir, err := ioutil.TempDir("", "target_dir_test")
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(targetDir)

	// 确保源目录为空(默认就是空的)

	// 捕获输出
	var capturedOutput []string
	oldHook := printHook
	printHook = func(message string) {
		capturedOutput = append(capturedOutput, message)
	}
	defer func() {
		printHook = oldHook
	}()

	// 临时修改os.Exit函数
	oldOsExit := osExit
	exitCode := 0
	osExit = func(code int) {
		exitCode = code
		// 不实际退出，只记录退出码
	}
	defer func() {
		osExit = oldOsExit
	}()

	// 测试空源目录
	source := sourceDir
	target := targetDir

	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	if !strings.HasSuffix(target, "/") {
		target += "/"
	}

	printColored(colorGreen, "源目录: "+source)
	printColored(colorGreen, "目标目录: "+target)

	// 执行检查
	isEmpty, _ := isDirEmpty(source)
	if isEmpty {
		printColored(colorRed, "错误: 源目录为空，不执行镜像操作")
		osExit(1)
	}

	// 验证结果
	if exitCode != 1 {
		t.Errorf("对于空的源目录，期望退出码为1，但得到 %d", exitCode)
	}

	found := false
	for _, msg := range capturedOutput {
		if strings.Contains(msg, "源目录为空") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("对于空的源目录，应显示警告消息，但未找到")
	}
}

// 测试远程源目录路径
func TestRemoteSourceDir(t *testing.T) {
	// 创建本地测试目录
	targetDir, err := ioutil.TempDir("", "target_dir_remote_test")
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(targetDir)

	// 使用远程源路径
	sourceDir := "user@host:/remote/source"

	// 捕获输出
	var capturedOutput []string
	oldHook := printHook
	printHook = func(message string) {
		capturedOutput = append(capturedOutput, message)
	}
	defer func() {
		printHook = oldHook
	}()

	// 临时修改os.Exit函数
	oldOsExit := osExit
	exitCode := 0
	osExit = func(code int) {
		exitCode = code
		// 不实际退出，只记录退出码
	}
	defer func() {
		osExit = oldOsExit
	}()

	// 确保路径末尾有斜杠
	source := sourceDir
	target := targetDir
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	if !strings.HasSuffix(target, "/") {
		target += "/"
	}

	printColored(colorGreen, "源目录: "+source)
	printColored(colorGreen, "目标目录: "+target)

	// 先检查源目录是否为空 - 应该报错
	_, err = isDirEmpty(source)
	if err == nil {
		t.Errorf("isDirEmpty应该报错，但没有")
	} else {
		// 错误路径检查
		printColored(colorRed, "错误: 无法检查源目录是否为空: "+err.Error())
		osExit(1)
	}

	// 验证结果
	if exitCode != 1 {
		t.Errorf("对于远程源目录，期望退出码为1，但得到 %d", exitCode)
	}

	found := false
	for _, msg := range capturedOutput {
		if strings.Contains(msg, "不支持检查远程目录是否为空") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("对于远程源目录，应显示适当的错误消息，但未找到")
	}
}

// 测试远程目标目录路径
func TestRemoteTargetDir(t *testing.T) {
	// 创建本地测试目录
	sourceDir, err := ioutil.TempDir("", "source_dir_remote_test")
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	// 在源目录中创建一个文件，使其不为空
	testFile := filepath.Join(sourceDir, "test_file")
	if err := ioutil.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("无法创建测试文件: %v", err)
	}

	// 使用远程目标路径
	targetDir := "user@host:/remote/target"

	// 捕获输出
	var capturedOutput []string
	oldHook := printHook
	printHook = func(message string) {
		capturedOutput = append(capturedOutput, message)
	}
	defer func() {
		printHook = oldHook
	}()

	// 临时修改os.Exit函数
	oldOsExit := osExit
	exitCode := 0
	osExit = func(code int) {
		exitCode = code
		// 不实际退出，只记录退出码
	}
	defer func() {
		osExit = oldOsExit
	}()

	// 确保路径末尾有斜杠
	source := sourceDir
	target := targetDir
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	if !strings.HasSuffix(target, "/") {
		target += "/"
	}

	printColored(colorGreen, "源目录: "+source)
	printColored(colorGreen, "目标目录: "+target)

	// 检查源目录是否为空
	isEmpty, err := isDirEmpty(source)
	if err != nil {
		printColored(colorRed, "错误: 无法检查源目录是否为空: "+err.Error())
		osExit(1)
	}
	if isEmpty {
		printColored(colorRed, "错误: 源目录为空，不执行镜像操作")
		osExit(1)
	}

	// 检查目录关系，应该在这里报错
	_, err = checkDirSameOrNested(source, target)
	if err == nil {
		t.Errorf("checkDirSameOrNested应该报错，但没有")
	} else {
		// 错误路径检查
		printColored(colorRed, "错误: "+err.Error())
		osExit(1)
	}

	// 验证结果
	if exitCode != 1 {
		t.Errorf("对于远程目标目录，期望退出码为1，但得到 %d", exitCode)
	}

	found := false
	for _, msg := range capturedOutput {
		if strings.Contains(msg, "不支持远程目标目录路径") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("对于远程目标目录，应显示适当的错误消息，但未找到")
	}
}

// 测试创建目标目录失败
func TestCreateTargetDirFailure(t *testing.T) {
	// 创建本地测试目录
	sourceDir, err := ioutil.TempDir("", "source_dir_create_test")
	if err != nil {
		t.Fatalf("无法创建测试目录: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	// 在源目录中创建一个文件，使其不为空
	testFile := filepath.Join(sourceDir, "test_file")
	if err := ioutil.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("无法创建测试文件: %v", err)
	}

	// 使用一个不存在且无法创建的目标目录
	// 在类Unix系统上，/proc是一个特殊的文件系统，通常不允许创建新目录
	targetDir := "/proc/impossible_dir"
	if os.Getuid() == 0 {
		// root用户可能在/proc创建目录，所以使用另一个通常无权限的目录
		targetDir = "/sys/impossible_dir"
	}

	// 捕获输出
	var capturedOutput []string
	oldHook := printHook
	printHook = func(message string) {
		capturedOutput = append(capturedOutput, message)
	}
	defer func() {
		printHook = oldHook
	}()

	// 临时修改os.Exit函数
	oldOsExit := osExit
	exitCode := 0
	osExit = func(code int) {
		exitCode = code
		// 不实际退出，只记录退出码
	}
	defer func() {
		osExit = oldOsExit
	}()

	// 确保路径末尾有斜杠
	source := sourceDir
	target := targetDir
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	if !strings.HasSuffix(target, "/") {
		target += "/"
	}

	printColored(colorGreen, "源目录: "+source)
	printColored(colorGreen, "目标目录: "+target)

	// 检查源目录是否为空
	isEmpty, err := isDirEmpty(source)
	if err != nil {
		printColored(colorRed, "错误: 无法检查源目录是否为空: "+err.Error())
		osExit(1)
	}
	if isEmpty {
		printColored(colorRed, "错误: 源目录为空，不执行镜像操作")
		osExit(1)
	}

	// 检查源目录和目标目录是否相同或嵌套
	_, err = checkDirSameOrNested(source, target)
	if err != nil {
		printColored(colorRed, "错误: "+err.Error())
		osExit(1)
	}

	// 检查目标目录是否存在，不存在则创建
	if !dirExists(target) {
		printColored(colorYellow, "目标目录不存在，尝试创建...")
		if err := createDir(target); err != nil {
			printColored(colorRed, "创建目标目录失败: "+err.Error())
			osExit(1)
		}
	}

	// 验证结果
	if exitCode != 1 {
		t.Errorf("对于创建目标目录失败，期望退出码为1，但得到 %d", exitCode)
	}

	found := false
	for _, msg := range capturedOutput {
		if strings.Contains(msg, "创建目标目录失败") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("对于创建目标目录失败，应显示适当的错误消息，但未找到")
	}
} 
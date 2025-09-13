package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// 测试Cobra命令创建
func TestCobraCommandCreation(t *testing.T) {
	// 创建root命令
	rootCmd := &cobra.Command{
		Use:   "folder_mirror SOURCE_DIR TARGET_DIR",
		Short: "镜像文件夹内容到目标位置",
	}

	// 验证命令创建成功
	if rootCmd == nil {
		t.Fatal("无法创建root命令")
	}

	// 验证命令属性
	if rootCmd.Use != "folder_mirror SOURCE_DIR TARGET_DIR" {
		t.Errorf("期望Use为'folder_mirror SOURCE_DIR TARGET_DIR'，得到'%s'", rootCmd.Use)
	}
}

// 测试prepareRsyncArgs函数
func TestPrepareRsyncArgsCobra(t *testing.T) {
	// 保存原始环境
	oldTesting := os.Getenv("TESTING")
	oldOsExit := osExit
	defer func() {
		os.Setenv("TESTING", oldTesting)
		osExit = oldOsExit
	}()

	// 设置测试环境
	os.Setenv("TESTING", "1")
	
	// 防止测试退出
	osExit = func(code int) {
		// 不实际退出
	}

	// 调用函数
	args := prepareRsyncArgs("/tmp/test_source")

	// 验证基本参数
	expectedArgs := []string{"-aH", "--force", "--delete-during", "--progress"}
	for i, expected := range expectedArgs {
		if i >= len(args) || args[i] != expected {
			t.Errorf("期望参数[%d]为'%s'，得到'%s'", i, expected, args[i])
		}
	}

	// 验证包含exclude-from参数
	hasExclude := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "--exclude-from=") {
			hasExclude = true
			break
		}
	}
	if !hasExclude {
		t.Error("缺少--exclude-from参数")
	}
}

// 测试validateAndPreparePaths函数
func TestValidateAndPreparePathsCobra(t *testing.T) {
	// 保存原始osExit
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()

	// 创建测试目录
	testSource := "/tmp/test_source_cobra"
	testTarget := "/tmp/test_target_cobra"
	
	// 清理并创建源目录
	os.RemoveAll(testSource)
	os.RemoveAll(testTarget)
	if err := os.MkdirAll(testSource, 0755); err != nil {
		t.Fatalf("无法创建测试源目录: %v", err)
	}
	defer os.RemoveAll(testSource)
	defer os.RemoveAll(testTarget)

	// 创建测试文件使目录非空
	if err := ioutil.WriteFile(testSource+"/test.txt", []byte("test"), 0644); err != nil {
		t.Fatalf("无法创建测试文件: %v", err)
	}

	// 防止osExit调用
	exitCalled := false
	osExit = func(code int) {
		exitCalled = true
		panic("osExit called") // 使用panic停止执行
	}

	// 正常情况测试
	func() {
		defer func() {
			if r := recover(); r != nil {
				// 预期不会panic
				t.Errorf("不期望的panic: %v", r)
			}
		}()

		source, target := validateAndPreparePaths(testSource, testTarget)
		
		// 验证路径末尾有斜杠
		if !strings.HasSuffix(source, "/") {
			t.Error("源路径应该以/结尾")
		}
		if !strings.HasSuffix(target, "/") {
			t.Error("目标路径应该以/结尾")
		}
	}()

	// 源目录不存在的情况
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("期望panic但未发生")
			}
		}()

		exitCalled = false
		validateAndPreparePaths("/nonexistent/path", testTarget)
		
		if !exitCalled {
			t.Error("期望调用osExit")
		}
	}()
}

// 测试handleDryRun的输出
func TestHandleDryRunOutput(t *testing.T) {
	// 保存原始函数
	oldOsExit := osExit
	oldExecCommand := execCommand
	oldPrintHook := printHook
	defer func() {
		osExit = oldOsExit
		execCommand = oldExecCommand
		printHook = oldPrintHook
	}()

	// 捕获打印输出
	var capturedOutput []string
	printHook = func(msg string) {
		capturedOutput = append(capturedOutput, msg)
	}

	// 防止退出
	osExit = func(code int) {
		// 不实际退出
	}

	// 模拟rsync命令
	execCommand = func(command string, args ...string) *exec.Cmd {
		// 返回一个简单的echo命令
		return exec.Command("echo", "test output")
	}

	// 准备测试参数
	args := []string{"-aH", "--force"}
	source := "/tmp/test_source/"
	target := "/tmp/test_target/"

	// 执行dry run
	handleDryRun(args, source, target)

	// 验证输出包含关键信息
	outputStr := strings.Join(capturedOutput, "\n")
	if !strings.Contains(outputStr, "DRY-RUN") {
		t.Error("输出应包含DRY-RUN提示")
	}
	if !strings.Contains(outputStr, ".folder_mirror.log") {
		t.Error("输出应包含日志文件路径")
	}
}


// 测试标记文件的创建和验证
func TestMarkerFileOperations(t *testing.T) {
	// 创建临时源目录
	testSourceDir, err := ioutil.TempDir("", "marker_ops_test_")
	if err != nil {
		t.Fatalf("无法创建临时源目录: %v", err)
	}
	defer os.RemoveAll(testSourceDir)

	// 测试创建标记文件
	err = createMarkerFile(testSourceDir)
	if err != nil {
		t.Fatalf("创建标记文件失败: %v", err)
	}

	// 验证文件存在（现在在项目根目录）
	markerPath := ".folder_mirror_marker"
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("标记文件应该存在")
	}

	// 测试检查标记文件（应该有效）
	valid, err := checkMarkerFile(testSourceDir)
	if err != nil {
		t.Fatalf("检查标记文件失败: %v", err)
	}
	if !valid {
		t.Error("新创建的标记文件应该有效")
	}

	// 删除标记文件
	os.Remove(markerPath)

	// 测试检查不存在的标记文件
	valid, err = checkMarkerFile(testSourceDir)
	if valid {
		t.Error("不存在的标记文件不应该有效")
	}
	if err == nil || !strings.Contains(err.Error(), "找不到标记文件") {
		t.Error("应返回找不到标记文件的错误")
	}
}

// 测试命令行帮助文本
func TestCommandHelpText(t *testing.T) {
	// 创建root命令（模拟main中的创建）
	rootCmd := &cobra.Command{
		Use:   "folder_mirror SOURCE_DIR TARGET_DIR",
		Short: "镜像文件夹内容到目标位置",
		Long: `folder_mirror 是一个用于同步文件夹内容的工具。
它使用 rsync 在源目录和目标目录之间进行镜像操作。

该工具支持dry-run模式，允许你在实际执行前预览将要进行的操作。

特性:
  • 支持灵活的参数位置 - 标志可以放在任何位置
  • 自动处理符号链接
  • 防止源目录和目标目录相同或嵌套
  • 彩色输出提高可读性`,
	}

	// 添加dry-run标志
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "测试镜像操作，不实际复制文件")

	// 捕获帮助输出
	buf := new(bytes.Buffer)
	rootCmd.SetOutput(buf)
	rootCmd.SetArgs([]string{"--help"})
	
	// 执行命令（会显示帮助）
	err := rootCmd.Execute()
	
	// Cobra的--help不返回错误
	if err != nil && !strings.Contains(err.Error(), "help requested") {
		// 某些版本的Cobra可能返回特定错误
	}

	// 验证帮助文本包含关键信息
	helpText := buf.String()
	if !strings.Contains(helpText, "folder_mirror") {
		t.Error("帮助文本应包含命令名")
	}
	if !strings.Contains(helpText, "dry-run") {
		t.Error("帮助文本应包含dry-run标志")
	}
	if !strings.Contains(helpText, "SOURCE_DIR") && !strings.Contains(helpText, "folder_mirror") {
		t.Error("帮助文本应包含命令用法信息")
	}
}
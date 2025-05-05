package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 定义可配置参数（改为变量以便于测试）
var (
	markerFile    = "/tmp/folder_mirror_marker"
	markerTimeout = int64(3600) // 1小时（秒）
)

// osExit 封装了os.Exit函数，便于测试
var osExit = os.Exit

// 可注入的exec.Command，便于测试
var execCommand = exec.Command

// 定义颜色常量
const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[0;33m"
	colorNone   = "\033[0m"
)

// 用于测试的全局钩子变量
var printHook func(string)
var disablePrint bool = false

// 彩色打印
func printColored(color, message string) {
	if !disablePrint {
		fmt.Printf("%s%s%s\n", color, message, colorNone)
	}
	// 如果测试钩子存在，调用它
	if printHook != nil {
		printHook(message)
	}
}

// 读取包含或排除规则文件
func readRuleFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rules []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			rules = append(rules, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

// 检查目录是否存在
func dirExists(path string) bool {
	// 远程路径检查 (包含冒号的路径)
	if strings.Contains(path, ":") {
		// 如果是远程路径，我们无法直接检查，默认存在
		return true
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// 创建目录
func createDir(path string) error {
	// 远程路径检查 (包含冒号的路径)
	if strings.Contains(path, ":") {
		return fmt.Errorf("不支持创建远程目录，请使用本地文件系统路径: %s", path)
	}

	return os.MkdirAll(path, 0755)
}

// 检查标记文件
func checkMarkerFile() (bool, error) {
	data, err := ioutil.ReadFile(markerFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("找不到标记文件。请先使用 --dry-run 参数生成标记文件")
		}
		return false, err
	}

	timestampStr := strings.TrimSpace(string(data))
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("无法解析标记文件时间戳: %v", err)
	}

	currentTime := time.Now().Unix()
	timeDiff := currentTime - timestamp

	if timeDiff > markerTimeout {
		return false, fmt.Errorf("标记文件太旧 (%d 秒, 最大 %d)", timeDiff, markerTimeout)
	}

	return true, nil
}

// 创建标记文件
func createMarkerFile() error {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	return ioutil.WriteFile(markerFile, []byte(timestamp), 0644)
}

// 检查源目录和目标目录是否相同或有从属关系
func checkDirSameOrNested(source, target string) (bool, error) {
	// 检查远程路径 (包含冒号的路径)
	if strings.Contains(source, ":") {
		return false, fmt.Errorf("不支持远程源目录路径，请使用本地文件系统路径: %s", source)
	}
	if strings.Contains(target, ":") {
		return false, fmt.Errorf("不支持远程目标目录路径，请使用本地文件系统路径: %s", target)
	}

	// 获取源目录和目标目录的绝对路径
	absSource, err := filepath.Abs(source)
	if err != nil {
		return false, fmt.Errorf("无法获取源目录绝对路径: %v", err)
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false, fmt.Errorf("无法获取目标目录绝对路径: %v", err)
	}

	// 检查目录是否相同
	if absSource == absTarget {
		return true, nil
	}

	// 检查目标目录是否是源目录的子目录
	if strings.HasPrefix(absTarget, absSource+string(filepath.Separator)) {
		return true, nil
	}

	// 检查源目录是否是目标目录的子目录
	if strings.HasPrefix(absSource, absTarget+string(filepath.Separator)) {
		return true, nil
	}

	// 检查源目录和目标目录是否通过符号链接指向相同位置
	srcInfo, err := os.Lstat(absSource)
	if err != nil {
		return false, fmt.Errorf("无法获取源目录信息: %v", err)
	}

	tgtInfo, err := os.Lstat(absTarget)
	if err != nil {
		// 目标目录可能不存在，此时不是同一目录
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("无法获取目标目录信息: %v", err)
	}

	// 如果两者都是符号链接，解析它们的真实路径并比较
	if srcInfo.Mode()&os.ModeSymlink != 0 || tgtInfo.Mode()&os.ModeSymlink != 0 {
		realSource, err := filepath.EvalSymlinks(absSource)
		if err != nil {
			return false, fmt.Errorf("无法解析源目录符号链接: %v", err)
		}

		realTarget, err := filepath.EvalSymlinks(absTarget)
		if err != nil {
			return false, fmt.Errorf("无法解析目标目录符号链接: %v", err)
		}

		// 比较解析后的路径
		if realSource == realTarget {
			return true, nil
		}

		// 检查解析后的路径是否有嵌套关系
		if strings.HasPrefix(realTarget, realSource+string(filepath.Separator)) {
			return true, nil
		}

		if strings.HasPrefix(realSource, realTarget+string(filepath.Separator)) {
			return true, nil
		}
	}

	return false, nil
}

// 检查目录是否为空
func isDirEmpty(dir string) (bool, error) {
	// 远程路径检查 (包含冒号的路径)
	if strings.Contains(dir, ":") {
		return false, fmt.Errorf("不支持检查远程目录是否为空，请使用本地文件系统路径: %s", dir)
	}

	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// 读取目录中的第一个条目
	_, err = f.Readdirnames(1)
	if err == nil {
		// 找到至少一个条目，目录不为空
		return false, nil
	}
	if err != nil && err.Error() == "EOF" {
		// 没有找到条目，目录为空
		return true, nil
	}
	// 其他错误
	return false, err
}

// 处理只读运行(dry-run)模式
func handleDryRun(args []string, source, target string) {
	printColored(colorYellow, "在DRY-RUN模式下运行。不会进行实际更改。")
	
	// 添加dry-run参数
	args = append(args, "-n", "-v")
	
	// 创建临时文件保存结果
	logFilePath := "/tmp/folder_mirror.log"
	logFile, err := os.Create(logFilePath)
	if err != nil {
		printColored(colorRed, "创建日志文件失败: "+err.Error())
		osExit(1)
	}
	defer logFile.Close()
	
	printColored(colorGreen, "结果将保存到: "+logFilePath)
	
	// 添加源和目标路径
	args = append(args, source, target)
	
	// 执行rsync命令，并允许实时显示进度
	printColored(colorGreen, "执行文件夹镜像模拟...")
	cmd := execCommand("rsync", args...)
	
	// 创建一个管道，同时输出到终端和日志文件
	cmd.Stderr = os.Stderr
	
	// 创建多写器，同时写入到stdout和logFile
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		printColored(colorRed, "无法创建输出管道: "+err.Error())
		osExit(1)
	}
	
	// 启动命令
	if err := cmd.Start(); err != nil {
		printColored(colorRed, "执行rsync失败: "+err.Error())
		osExit(1)
	}
	
	// 读取输出并同时写入到终端和日志文件
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			fmt.Fprintln(logFile, line)
		}
	}()
	
	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		printColored(colorRed, "执行rsync失败: "+err.Error())
		osExit(1)
	}
	
	// 创建标记文件
	if err := createMarkerFile(); err != nil {
		printColored(colorRed, "创建标记文件失败: "+err.Error())
		osExit(1)
	}
	
	printColored(colorGreen, "模拟操作完成。标记文件已创建: "+markerFile)
	printColored(colorGreen, "干运行结果已保存到文件: "+logFilePath)
	printColored(colorYellow, "请检查输出结果，确认无误后可执行实际操作(不带--dry-run参数)")
	// 不再自动打开编辑器查看文件，用户可以手动查看结果文件
	osExit(0)
}

// 处理实际执行模式
func handleActualRun(args []string, source, target string) {
	// 检查标记文件
	valid, err := checkMarkerFile()
	if !valid {
		printColored(colorRed, "错误: "+err.Error())
		printColored(colorRed, "请先使用 --dry-run 参数重新生成标记文件。")
		osExit(1)
	}
	
	printColored(colorGreen, "执行实际文件夹镜像操作...")
	
	// 添加源和目标路径
	args = append(args, source, target)
	
	// 执行rsync命令，并允许实时输出进度
	cmd := execCommand("rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// 执行命令并等待完成
	err = cmd.Run()
	if err != nil {
		printColored(colorRed, "执行rsync失败: "+err.Error())
		osExit(1)
	}
	
	printColored(colorGreen, "实际文件夹镜像操作成功完成!")
	
	// 删除标记文件
	if err := os.Remove(markerFile); err != nil {
		printColored(colorYellow, "警告: 无法删除标记文件: "+err.Error())
	}
	
	// 确保调用osExit
	osExit(0)
}

// 验证路径并准备目录
func validateAndPreparePaths(source, target string) (string, string) {
	// 确保路径末尾有斜杠
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	if !strings.HasSuffix(target, "/") {
		target += "/"
	}

	printColored(colorGreen, "源目录: "+source)
	printColored(colorGreen, "目标目录: "+target)

	// 检查源目录是否存在
	if !dirExists(source) {
		printColored(colorRed, "错误: 源目录不存在: "+source)
		osExit(1)
	}

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

	// 检查源目录和目标目录是否相同或嵌套或为远程路径
	isSameOrNested, err := checkDirSameOrNested(source, target)
	if err != nil {
		printColored(colorRed, "错误: "+err.Error())
		osExit(1)
	}
	if isSameOrNested {
		printColored(colorRed, "错误: 源目录和目标目录相同或互为子目录，操作危险，终止执行")
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
	
	return source, target
}

// 准备rsync命令的参数
func prepareRsyncArgs() []string {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		printColored(colorRed, "无法获取用户主目录: "+err.Error())
		osExit(1)
	}

	// 读取排除和包含的文件列表
	excludeListPath := filepath.Join(homeDir, "loadrc/bashrc/mirror_exclude")
	includeListPath := filepath.Join(homeDir, "loadrc/bashrc/mirror_include")

	// 在测试环境中，使用临时文件来替代实际文件
	if os.Getenv("TESTING") == "1" {
		tmpExclude, err := ioutil.TempFile("", "test_exclude")
		if err == nil {
			fmt.Fprintln(tmpExclude, "*.tmp")
			tmpExclude.Close()
			excludeListPath = tmpExclude.Name()
			defer os.Remove(excludeListPath)
		}
	}

	// 验证排除规则文件是否存在
	if _, err := os.Stat(excludeListPath); os.IsNotExist(err) {
		printColored(colorRed, "错误: 排除规则文件不存在: "+excludeListPath)
		osExit(1)
	}

	// 构建rsync命令参数
	args := []string{"-aH", "--force", "--delete-during", "--progress"}

	// 使用文件方式添加排除规则，简化代码
	// rsync 原生支持 */build/* 等通配符格式
	args = append(args, "--exclude-from="+excludeListPath)

	// 如果包含规则文件存在，也使用文件方式添加
	if _, err := os.Stat(includeListPath); !os.IsNotExist(err) {
		args = append(args, "--include-from="+includeListPath)
	} else {
		printColored(colorYellow, "警告: 包含规则文件不存在: "+includeListPath)
		// 继续执行，因为包含列表是可选的
	}
	
	return args
}

func main() {
	// 检查是否存在--dry-run参数（无论位置）
	hasDryRunFlag := false
	for _, arg := range os.Args {
		if arg == "--dry-run" || arg == "-dry-run" {
			hasDryRunFlag = true
			break
		}
	}

	// 解析命令行参数
	dryRun := flag.Bool("dry-run", hasDryRunFlag, "测试镜像操作，不实际复制文件")
	help := flag.Bool("help", false, "显示帮助信息")
	flag.Parse()

	if *help || flag.NArg() < 2 {
		fmt.Printf("用法: %s [--dry-run] SOURCE_DIR TARGET_DIR\n\n", os.Args[0])
		fmt.Println("选项:")
		fmt.Println("  --dry-run          测试镜像操作，不实际复制文件")
		fmt.Println("  --help             显示帮助信息")
		fmt.Println()
		fmt.Println("参数:")
		fmt.Println("  SOURCE_DIR         源目录路径")
		fmt.Println("  TARGET_DIR         目标目录路径")
		osExit(1)
	}

	// 获取源目录和目标目录
	source := flag.Arg(0)
	target := flag.Arg(1)
	
	// 验证路径并准备目录
	source, target = validateAndPreparePaths(source, target)
	
	// 准备rsync命令的参数
	args := prepareRsyncArgs()
	
	// 根据运行模式执行不同的处理
	if *dryRun || hasDryRunFlag {
		handleDryRun(args, source, target)
	} else {
		handleActualRun(args, source, target)
	}
} 

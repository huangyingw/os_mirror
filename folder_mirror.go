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
	// 跳过远程路径检查 (包含冒号的路径)
	if strings.Contains(path, ":") {
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
	// 跳过远程路径 (包含冒号的路径)
	if strings.Contains(path, ":") {
		return nil
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

func main() {
	// 解析命令行参数
	dryRun := flag.Bool("dry-run", false, "测试镜像操作，不实际复制文件")
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
		os.Exit(1)
	}

	// 获取源目录和目标目录
	source := flag.Arg(0)
	target := flag.Arg(1)

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
		os.Exit(1)
	}

	// 检查目标目录是否存在，不存在则创建
	if !dirExists(target) {
		printColored(colorYellow, "目标目录不存在，尝试创建...")
		if err := createDir(target); err != nil {
			printColored(colorRed, "创建目标目录失败: "+err.Error())
			os.Exit(1)
		}
	}

	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		printColored(colorRed, "无法获取用户主目录: "+err.Error())
		os.Exit(1)
	}

	// 读取排除和包含的文件列表
	excludeListPath := filepath.Join(homeDir, "loadrc/bashrc/mirror_exclude")
	includeListPath := filepath.Join(homeDir, "loadrc/bashrc/mirror_include")

	// 验证排除规则文件是否存在
	if _, err := os.Stat(excludeListPath); os.IsNotExist(err) {
		printColored(colorRed, "错误: 排除规则文件不存在: "+excludeListPath)
		os.Exit(1)
	}

	// 构建rsync命令参数
	args := []string{"-aH", "--force", "--delete-during"}

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

	if *dryRun {
		printColored(colorYellow, "在DRY-RUN模式下运行。不会进行实际更改。")
		
		// 添加dry-run参数
		args = append(args, "-n", "-v")
		
		// 创建临时文件保存结果
		tmpFile, err := ioutil.TempFile("", "folder_mirror_*.log")
		if err != nil {
			printColored(colorRed, "创建临时文件失败: "+err.Error())
			os.Exit(1)
		}
		defer tmpFile.Close()
		
		printColored(colorGreen, "结果将保存到: "+tmpFile.Name())
		
		// 添加源和目标路径
		args = append(args, source, target)
		
		// 执行rsync命令
		cmd := exec.Command("rsync", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			printColored(colorRed, "执行rsync失败: "+err.Error())
			os.Exit(1)
		}
		
		// 保存输出到临时文件
		if _, err := tmpFile.Write(output); err != nil {
			printColored(colorRed, "写入临时文件失败: "+err.Error())
			os.Exit(1)
		}
		
		// 创建标记文件
		if err := createMarkerFile(); err != nil {
			printColored(colorRed, "创建标记文件失败: "+err.Error())
			os.Exit(1)
		}
		
		printColored(colorGreen, "预览完成。标记文件已创建: "+markerFile)
		
		// 在编辑器中打开结果文件
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		
		cmd = exec.Command(editor, tmpFile.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			printColored(colorRed, "打开编辑器失败: "+err.Error())
			os.Exit(1)
		}
	} else {
		// 检查标记文件
		valid, err := checkMarkerFile()
		if !valid {
			printColored(colorRed, "错误: "+err.Error())
			printColored(colorRed, "请先使用 --dry-run 参数重新生成标记文件。")
			os.Exit(1)
		}
		
		printColored(colorGreen, "执行实际文件夹镜像...")
		
		// 添加源和目标路径
		args = append(args, source, target)
		
		// 执行rsync命令
		cmd := exec.Command("rsync", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			printColored(colorRed, "执行rsync失败: "+err.Error())
			fmt.Println(string(output))
			os.Exit(1)
		}
		
		printColored(colorGreen, "文件夹镜像成功完成!")
		
		// 删除标记文件
		if err := os.Remove(markerFile); err != nil {
			printColored(colorYellow, "警告: 无法删除标记文件: "+err.Error())
		}
	}
} 
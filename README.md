# folder_mirror - 文件夹镜像工具

`folder_mirror` 是一个用 Go 编写的工具，用于有选择地将一个文件夹镜像到另一个文件夹。它使用 `rsync` 作为底层复制工具，支持包含和排除规则，以及预览模式。

## 功能特点

- 基于 rsync 进行高效的文件复制
- 支持预览模式 (dry-run)，可以查看哪些文件将被复制
- 使用标记文件确保预览后再执行实际操作
- 支持通过配置文件定义包含和排除规则
- 支持本地和远程路径
- 彩色输出，提供更好的用户体验

## 安装

```bash
go build -o folder_mirror folder_mirror.go
```

## 使用方法

```
folder_mirror [--dry-run] SOURCE_DIR TARGET_DIR

选项:
  --dry-run          测试镜像操作，不实际复制文件
  --help             显示帮助信息

参数:
  SOURCE_DIR         源目录路径
  TARGET_DIR         目标目录路径
```

## 工作流程

1. 使用 `--dry-run` 预览将要进行的操作
2. 检查预览结果，确认无误
3. 运行命令（不带 `--dry-run` 参数）执行实际操作

## 配置文件

该工具使用两个配置文件来定义包含和排除规则：

- `$HOME/loadrc/bashrc/mirror_exclude` - 包含要排除的文件和目录模式
- `$HOME/loadrc/bashrc/mirror_include` - 包含要明确包含的文件和目录模式

### 排除文件格式

```
# 这是注释
/path/to/exclude/
*.tmp
```

### 包含文件格式

```
# 这是注释
/path/to/include/
*.important
```

## 示例

预览模式：

```bash
folder_mirror --dry-run /home/user/source/ /backup/target/
```

执行实际复制：

```bash
folder_mirror /home/user/source/ /backup/target/
```

## 运行测试

```bash
go test -v
``` 
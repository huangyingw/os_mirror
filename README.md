# folder_mirror - 文件夹镜像工具

`folder_mirror` 是一个用 Go 编写的工具，用于有选择地将一个文件夹镜像到另一个文件夹。它使用 `rsync` 作为底层复制工具，支持包含和排除规则，以及预览模式。

## 功能特点

- 基于 rsync 进行高效的文件复制
- 支持预览模式 (dry-run)，可以查看哪些文件将被复制，预览结果会保存到文件
- 使用标记文件确保预览后再执行实际操作
- 支持通过配置文件定义包含和排除规则
- 支持本地和远程路径
- 彩色输出，提供更好的用户体验
- 防止相同或嵌套目录之间的操作，避免潜在的文件损失
- 防止空源目录的镜像，避免清空目标目录
- 防止对远程路径执行危险操作

## 构建和安装

### 使用 Makefile 构建

推荐使用提供的 Makefile 进行构建和安装：

```bash
# 构建应用程序
make build

# 运行测试
make test

# 生成测试覆盖率报告
make coverage

# 安装到系统目录
sudo make install

# 清理构建文件
make clean

# 查看所有可用命令
make help
```

### 手动构建

如果不使用 Makefile，也可以手动构建：

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

## 安全特性

该工具包含多项安全检查，以防止意外的数据丢失：

1. 防止在相同或嵌套目录之间执行镜像操作
2. 防止从空源目录镜像（这可能会清空目标目录）
3. 对远程路径执行额外的安全检查

## 工作流程

1. 使用 `--dry-run` 预览将要进行的操作，结果会保存到临时文件
2. 查看生成的预览结果文件，确认无误
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

## 开发和测试

### 运行测试

```bash
# 运行所有测试
make test

# 生成测试覆盖率报告
make coverage
```

测试覆盖率报告将生成在 `coverage.html` 文件中，可以在浏览器中查看详细的覆盖情况。

### 代码结构

- `folder_mirror.go` - 主程序代码
- `folder_mirror_test.go` - 测试文件
- `folder_mirror_test_utils.go` - 测试辅助函数

## 依赖项

- Go 1.16 或更高版本
- 系统中已安装 rsync 
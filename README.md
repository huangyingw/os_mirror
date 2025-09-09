# Folder Mirror Tool

一个基于rsync的文件夹镜像工具，使用Go语言和Cobra框架实现。该工具可以安全地将源目录同步到目标目录，支持dry-run模式确保操作安全。

## 功能特性

- **安全的文件夹镜像**：使用rsync进行高效的文件同步
- **Dry-run模式**：在实际执行前预览将要进行的操作
- **标记文件保护机制**：防止误操作导致数据丢失
- **自定义规则支持**：通过配置文件控制包含/排除规则
- **详细的操作日志**：记录所有操作到日志文件
- **目录安全检查**：防止相同或互为子目录的操作
- **空目录保护**：防止空源目录导致目标目录被清空

## 安装

```bash
go build -o folder_mirror folder_mirror.go
```

## 使用方法

### 基本用法

```bash
# Dry-run模式（推荐先执行）
./folder_mirror --dry-run /path/to/source/ /path/to/target/

# 实际执行（需要先运行dry-run）
./folder_mirror /path/to/source/ /path/to/target/
```

### 命令行参数

- `--dry-run, -n`: 只模拟运行，不进行实际文件操作
- `--help, -h`: 显示帮助信息

## 文件位置策略

### 配置文件位置

配置文件必须放在**项目根目录**（即folder_mirror可执行文件所在的目录）：

- `mirror_exclude`: 排除规则文件
- `mirror_include`: 包含规则文件

**注意**：配置文件只能放在项目根目录，其他位置的配置文件不会生效。

### 运行时文件位置

运行时产生的文件保存在**源目录**中：

- `.folder_mirror_marker`: 标记文件（包含时间戳）
- `.folder_mirror.log`: 操作日志文件

### 文件结构示例

```
项目根目录/
├── folder_mirror           # 可执行文件
├── mirror_exclude          # 排除规则配置（可选）
├── mirror_include          # 包含规则配置（可选）
└── mirror_exclude.template # 排除规则模板（参考）

源目录/
├── .folder_mirror_marker   # 运行时生成的标记文件
├── .folder_mirror.log      # 运行时生成的日志文件
└── ... （其他源文件）
```

## 配置文件说明

### mirror_exclude（排除规则）

指定要排除的文件和目录模式，每行一个规则：

```
# 示例 mirror_exclude
.git/
.svn/
*.tmp
*.swp
.DS_Store
node_modules/
__pycache__/
```

### mirror_include（包含规则）

指定要包含的文件模式（通常与排除规则配合使用）：

```
# 示例 mirror_include
*.go
*.md
```

### 默认排除规则

如果没有提供`mirror_exclude`文件，工具会使用以下默认排除规则：

- `.git/`
- `.svn/`
- `*.tmp`
- `*.swp`
- `.folder_mirror_marker`
- `.folder_mirror.log`

## 安全机制

### 标记文件保护

1. **Dry-run生成标记**：dry-run模式成功后会在源目录生成`.folder_mirror_marker`文件
2. **实际执行验证**：实际执行前检查标记文件是否存在且有效（5分钟内）
3. **自动清理**：实际执行成功后自动删除标记文件

### 目录安全检查

- 源目录和目标目录不能相同
- 源目录和目标目录不能互为子目录
- 源目录不能为空
- 不支持远程路径操作

## 测试

运行单元测试：

```bash
go test -v
```

运行测试覆盖率：

```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## 注意事项

1. **先执行dry-run**：强烈建议在实际执行前先运行dry-run模式查看将要进行的操作
2. **检查日志文件**：dry-run后查看`.folder_mirror.log`文件确认操作内容
3. **备份重要数据**：在对重要数据进行镜像操作前，请确保有备份
4. **配置文件位置**：确保配置文件放在正确的位置（项目根目录）

## 依赖项

- Go 1.16 或更高版本
- 系统中已安装 rsync
- Cobra CLI 框架

## 许可证

MIT License 
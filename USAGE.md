# Folder Mirror 使用说明

## 概述
Folder Mirror 是一个基于 rsync 的文件夹同步工具，支持 dry-run 模式预览操作，确保安全可靠的文件同步。

## 主要特性
- 🔍 **Dry-run 模式**：执行前预览所有操作
- 📝 **操作日志**：详细记录同步过程
- 🛡️ **安全检查**：防止源目录和目标目录相同或嵌套
- 🎨 **彩色输出**：提高可读性
- 📁 **灵活配置**：支持包含/排除规则

## 安装

```bash
# 编译程序
go build -o folder_mirror folder_mirror.go

# 添加执行权限
chmod +x folder_mirror

# （可选）移动到 PATH 目录
sudo mv folder_mirror /usr/local/bin/
```

## 基本用法

### 1. Dry-run 模式（推荐先执行）
```bash
# 预览操作，不实际执行
folder_mirror --dry-run /source/dir /target/dir

# 或使用短格式
folder_mirror -n /source/dir /target/dir
```

### 2. 实际执行
```bash
# 必须先执行 dry-run 生成标记文件
folder_mirror /source/dir /target/dir
```

### 3. 查看帮助
```bash
folder_mirror --help
```

## 文件位置说明

### 生成的文件
执行后会在**源目录**中生成以下文件：

1. **`.folder_mirror.log`** - 操作日志文件
   - 记录所有同步的文件和操作
   - dry-run 模式下显示将要执行的操作

2. **`.folder_mirror_marker`** - 标记文件
   - 用于确保先执行 dry-run 再执行实际操作
   - 有效期 1 小时
   - 实际执行后自动删除

### 配置文件位置

配置文件位于**源目录**中：

1. **`源目录/mirror_exclude`** - 排除规则文件（推荐）
   - 指定不需要同步的文件和目录
   - 每行一个规则，支持通配符
   - 如果不存在，使用默认排除规则

2. **`源目录/mirror_include`** - 包含规则文件（可选）
   - 强制包含的文件（优先级高于 exclude）
   - 通常不需要，除非要覆盖排除规则

这种设计让每个项目可以有自己的同步规则，便于版本管理。

## 配置文件设置

### 创建排除规则文件

```bash
# 复制模板文件到源目录
cp mirror_exclude.template /path/to/source/mirror_exclude

# 编辑配置
vim /path/to/source/mirror_exclude
```

### 排除规则示例
```bash
# 系统目录
/media/
/tmp/

# 版本控制
.git/
.svn/

# 构建输出
*/build/*
*/dist/*
node_modules/

# 临时文件
*.tmp
*.log

# 特定项目
*/myproject/*
```

### 包含规则示例（可选）
```bash
# 在源目录创建包含规则（如果需要）
cp mirror_include.template /path/to/source/mirror_include

# 示例内容：
# 即使在 node_modules 中也要包含
node_modules/my-local-package/

# 在排除目录中保留特定文件
build/README.md
```

## 工作流程

### 推荐的安全流程

1. **第一步：Dry-run 预览**
   ```bash
   folder_mirror --dry-run /source /target
   ```
   - 查看 `.folder_mirror.log` 确认操作
   - 生成标记文件

2. **第二步：检查日志**
   ```bash
   # 查看生成的日志
   cat /source/.folder_mirror.log
   ```

3. **第三步：执行同步**
   ```bash
   # 确认无误后执行
   folder_mirror /source /target
   ```

## 高级用法

### 参数位置灵活性
Cobra 框架支持参数在任意位置：

```bash
# 标志在前
folder_mirror --dry-run /source /target

# 标志在中间
folder_mirror /source --dry-run /target

# 标志在后
folder_mirror /source /target --dry-run
```

### 远程同步
支持 rsync 的远程语法：
```bash
# 从远程到本地
folder_mirror user@host:/remote/source /local/target

# 从本地到远程
folder_mirror /local/source user@host:/remote/target
```

## 注意事项

1. **标记文件有效期**：1小时，超时需重新执行 dry-run
2. **源目录必须非空**：空目录会报错
3. **自动创建目标目录**：如果不存在会自动创建
4. **删除同步**：会删除目标中源不存在的文件（`--delete-during`）
5. **保留权限**：保留文件权限和时间戳（`-aH`）

## 错误处理

### 常见错误及解决方案

1. **"找不到标记文件"**
   - 先执行 `--dry-run` 生成标记文件

2. **"标记文件太旧"**
   - 重新执行 `--dry-run`

3. **"源目录为空"**
   - 确保源目录包含文件

4. **"源目录和目标目录相同"**
   - 使用不同的目录路径

5. **"源目录中未找到 mirror_exclude 文件"**
   - 程序会使用默认排除规则
   - 建议在源目录创建 `mirror_exclude` 文件

## 测试建议

在生产环境使用前，建议：

1. 使用测试目录验证配置
2. 仔细检查 dry-run 输出
3. 备份重要数据
4. 逐步测试排除/包含规则

## 技术细节

- 基于 Go 语言和 Cobra 框架
- 底层使用 rsync 命令
- 支持符号链接
- 彩色终端输出
- 单元测试覆盖率 > 60%

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可

[添加你的许可信息]
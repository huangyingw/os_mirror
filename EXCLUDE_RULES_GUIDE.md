# mirror_exclude 文件配置指南

## 排除 boot/grub/ 的不同方式

### 1. 排除所有位置的 boot/grub/ 目录
```
boot/grub/
```
效果：排除源目录中任何位置的 `boot/grub/` 目录及其所有内容
- ✗ /boot/grub/grub.cfg
- ✗ /data/boot/grub/menu.lst
- ✓ /boot/vmlinuz

### 2. 只排除根目录下的 boot/grub/
```
/boot/grub/
```
效果：只排除源根目录下的 `boot/grub/`
- ✗ /boot/grub/grub.cfg
- ✓ /data/boot/grub/menu.lst (不会被排除)
- ✓ /boot/vmlinuz

### 3. 排除所有 boot 目录
```
boot/
```
效果：排除任何位置的 `boot/` 目录及其所有内容
- ✗ /boot/grub/grub.cfg
- ✗ /boot/vmlinuz
- ✗ /data/boot/grub/menu.lst

### 4. 只排除根目录下的 boot
```
/boot/
```
效果：只排除源根目录下的 `boot/` 目录
- ✗ /boot/grub/grub.cfg
- ✗ /boot/vmlinuz
- ✓ /data/boot/grub/menu.lst (不会被排除)

## 推荐配置

如果你想排除 boot/grub/ 目录，最常用的配置是：

```
# 排除所有位置的 boot/grub/ 目录
boot/grub/
```

如果你的源目录结构是：
```
/source/
├── boot/
│   ├── grub/     <- 会被排除
│   │   └── grub.cfg
│   └── vmlinuz   <- 不会被排除
├── data/
│   └── boot/
│       └── grub/ <- 也会被排除
└── other/
    └── file.txt  <- 不会被排除
```

## 注意事项

1. **路径分隔符**：始终使用正斜杠 `/`，即使在 Windows 上
2. **目录标识**：目录名后加 `/` 表示排除整个目录
3. **相对vs绝对**：
   - `boot/grub/` - 相对路径，匹配任何位置
   - `/boot/grub/` - 绝对路径，只匹配源根目录下的路径
4. **通配符**：支持 `*` 和 `**` 等通配符
   - `*.log` - 排除所有 .log 文件
   - `**/temp/` - 排除任何位置的 temp 目录

## 测试你的规则

使用 dry-run 模式测试排除规则是否正确：
```bash
./folder_mirror --dry-run /source/ /target/
```

查看生成的日志文件确认哪些文件被包含/排除：
```bash
cat /source/.folder_mirror.log
```
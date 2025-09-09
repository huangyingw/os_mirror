#!/bin/bash
# 测试 folder_mirror 的强大功能

echo "=== Folder Mirror 功能展示 ==="
echo

# 创建测试目录
TEST_DIR="/tmp/cobra_test_$$"
SOURCE_DIR="$TEST_DIR/source"
TARGET_DIR="$TEST_DIR/target"
mkdir -p "$SOURCE_DIR" "$TARGET_DIR"
echo "test file" > "$SOURCE_DIR/test.txt"

echo "1. 🎯 查看自动生成的帮助"
echo "================================"
./folder_mirror --help
echo

echo "2. ✨ Cobra 自动支持参数任意位置"
echo "=================================="
echo

echo "✅ 标志在前:"
echo "./folder_mirror --dry-run $SOURCE_DIR $TARGET_DIR"
./folder_mirror --dry-run "$SOURCE_DIR" "$TARGET_DIR" 2>&1 | head -3
echo

echo "✅ 标志在中间:"
echo "./folder_mirror $SOURCE_DIR --dry-run $TARGET_DIR"
./folder_mirror "$SOURCE_DIR" --dry-run "$TARGET_DIR" 2>&1 | head -3
echo

echo "✅ 标志在后:"
echo "./folder_mirror $SOURCE_DIR $TARGET_DIR --dry-run"
./folder_mirror "$SOURCE_DIR" "$TARGET_DIR" --dry-run 2>&1 | head -3
echo

echo "✅ 短格式 -n 也支持任意位置:"
echo "./folder_mirror $SOURCE_DIR $TARGET_DIR -n"
./folder_mirror "$SOURCE_DIR" "$TARGET_DIR" -n 2>&1 | head -3
echo

echo "3. 🚨 Cobra 的错误处理"
echo "======================="
echo

echo "❌ 参数不足:"
./folder_mirror "$SOURCE_DIR" 2>&1 | head -2
echo

echo "❌ 参数过多:"
./folder_mirror "$SOURCE_DIR" "$TARGET_DIR" "extra" 2>&1 | head -2
echo

echo "❌ 未知标志:"
./folder_mirror --unknown-flag "$SOURCE_DIR" "$TARGET_DIR" 2>&1 | head -2
echo

# 清理
rm -rf "$TEST_DIR"

echo "=== Folder Mirror 优势总结 ==="
echo "✅ 自动支持标志在任意位置"
echo "✅ 自动生成美观的帮助文档"
echo "✅ 自动验证参数数量"
echo "✅ 自动处理未知标志"
echo "✅ 支持短格式和长格式"
echo "✅ 无需手工编写参数解析代码"
echo "✅ 代码更清晰、可维护性更好"
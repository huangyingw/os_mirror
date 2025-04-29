# Makefile for folder_mirror

# 变量
GO = go
BIN_NAME = folder_mirror
BUILD_DIR = build
COVERAGE_PROFILE = coverage.out
COVERAGE_HTML = coverage.html

# 默认目标
.PHONY: all
all: build

# 构建应用程序
.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BIN_NAME) folder_mirror.go

# 运行测试
.PHONY: test
test:
	$(GO) test -v

# 生成测试覆盖率报告
.PHONY: coverage
coverage:
	$(GO) test -coverprofile=$(COVERAGE_PROFILE)
	$(GO) tool cover -func=$(COVERAGE_PROFILE)
	$(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated in $(COVERAGE_HTML)"

# 安装应用程序
.PHONY: install
install: build
	@cp $(BUILD_DIR)/$(BIN_NAME) /usr/local/bin/
	@echo "Installed $(BIN_NAME) to /usr/local/bin/"

# 清理生成的文件
.PHONY: clean
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_PROFILE)
	@rm -f $(COVERAGE_HTML)
	@echo "Cleaned up build files"

# 显示帮助信息
.PHONY: help
help:
	@echo "适用于 folder_mirror 的 Makefile 命令:"
	@echo "  make            - 构建应用程序"
	@echo "  make build      - 构建应用程序"
	@echo "  make test       - 运行单元测试"
	@echo "  make coverage   - 生成测试覆盖率报告"
	@echo "  make install    - 安装应用程序到 /usr/local/bin/"
	@echo "  make clean      - 清理构建文件"
	@echo "  make help       - 显示此帮助信息" 
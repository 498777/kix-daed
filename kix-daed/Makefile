# 基础变量
NAME := daed
BIN_DIR := bin
VERSION ?= $(shell git describe --tags --always || echo "dev")

# 你的自定义仓库配置
CUSTOM_DAE_REPO := https://github.com/olicesx/dae.git
CUSTOM_DAE_BRANCH := optimize/code-quality-fixes

# 编译指令 (强制关闭 CGO 以实现静态编译，增强兼容性)
LDFLAGS := -s -w -X github.com/daeuniverse/daed/internal/version.Version=$(VERSION)
GOBUILD := CGO_ENABLED=0 go build -v -trimpath -ldflags "$(LDFLAGS)"

.PHONY: all prepare patch build-web build-backend clean

# 一键全流程编译
all: prepare patch build-web build-backend

# 1. 环境准备：初始化官方子模块
prepare:
	@echo "==> 正在初始化官方子模块..."
	git submodule update --init --recursive wing

# 2. 核心补丁：替换 dae-core 为你的自定义仓库
patch:
	@echo "==> 正在替换 dae-core 为自定义仓库..."
	cd wing && \
	git submodule set-url dae-core $(CUSTOM_DAE_REPO) && \
	git submodule update --init --remote dae-core && \
	cd dae-core && \
	git fetch origin $(CUSTOM_DAE_BRANCH) && \
	git checkout $(CUSTOM_DAE_BRANCH)
	@echo "==> 子模块替换完成。"

# 3. 前端编译：使用 pnpm 编译 UI
build-web:
	@echo "==> 正在编译前端界面..."
	pnpm install
	pnpm build --filter daed

# 4. 后端编译：根据环境变量 GOOS/GOARCH 生成二进制
build-backend:
	@echo "==> 正在编译后端二进制 ($(GOOS)-$(GOARCH))..."
	mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(NAME)-$(GOOS)-$(GOARCH)

clean:
	rm -rf $(BIN_DIR)
	rm -rf apps/web/dist

NAME := daed
BIN_DIR := bin

# 接收从 Action 传入的各个仓库短哈希
D_SHA ?= $(shell git rev-parse --short HEAD)
W_SHA ?= "unknown"
C_SHA ?= "unknown"

# 注入版本号格式：daed_HASH-wing_HASH-core_HASH
FULL_VERSION := daed_$(D_SHA)-wing_$(W_SHA)-core_$(C_SHA)

LDFLAGS := -s -w -X github.com/daeuniverse/daed/internal/version.Version=$(FULL_VERSION)
GOBUILD := CGO_ENABLED=0 go build -v -trimpath -ldflags "$(LDFLAGS)"

.PHONY: all prepare patch build-web build-backend

all: prepare patch build-web build-backend

prepare:
	@echo "==> 初始化官方子模块..."
	git submodule update --init --recursive wing

patch:
	@echo "==> 替换并追踪自定义内核..."
	# 使用 cd 确保在子模块路径内操作
	cd wing && \
	git submodule set-url dae-core https://github.com/olicesx/dae.git && \
	git submodule update --init --remote dae-core
	cd wing/dae-core && \
	git fetch origin optimize/code-quality-fixes && \
	git checkout optimize/code-quality-fixes

build-web:
	@echo "==> 正在安装依赖并配置环境..."
	pnpm install --no-frozen-lockfile
	@echo "==> 正在按照依赖链编译所有前端包..."
	pnpm build

build-backend:
	@echo "==> 编译后端二进制 ($(GOOS)-$(GOARCH))..."
	mkdir -p $(BIN_DIR)
	# 关键修复：进入 wing 目录后再执行 go build
	# 因为 main 模块（go.mod）在 wing 文件夹里
	cd wing && \
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -o ../$(BIN_DIR)/$(NAME)-$(GOOS)-$(GOARCH) .

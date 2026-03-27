# 路径：kix-daed/Makefile

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
	cd wing && \
	git submodule set-url dae-core https://github.com/olicesx/dae.git && \
	git submodule update --init --remote dae-core && \
	cd dae-core && \
	git fetch origin optimize/code-quality-fixes && \
	git checkout optimize/code-quality-fixes

build-web:
	@echo "==> 编译前端..."
	pnpm install
	pnpm build --filter daed

build-backend:
	@echo "==> 编译后端二进制 ($(GOOS)-$(GOARCH))..."
	mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(NAME)-$(GOOS)-$(GOARCH)

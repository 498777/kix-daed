NAME := daed
BIN_DIR := bin

# 强制指定对齐官方环境的工具链
CLANG_VER := 15
export CLANG := clang-$(CLANG_VER)
export STRIP := llvm-strip-$(CLANG_VER)
export GOEXPERIMENT := newinliner,simd

# 版本号计算
D_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
W_SHA ?= "unknown"
C_SHA ?= "unknown"
FULL_VERSION := daed_$(D_SHA)-wing_$(W_SHA)-core_$(C_SHA)

.PHONY: all prepare build-web build-backend

all: prepare build-web build-backend

prepare:
	@echo "==> 1. 初始化子模块并切换自定义内核..."
	git submodule update --init --recursive wing
	cd wing && \
		git submodule set-url dae-core https://github.com/olicesx/dae.git && \
		git submodule update --init --remote --recursive dae-core
	cd wing/dae-core && \
		git fetch origin optimize/code-quality-fixes && \
		git checkout optimize/code-quality-fixes

build-web:
	@echo "==> 2. 正在编译前端..."
	pnpm install --no-frozen-lockfile
	pnpm build

build-backend:
	@echo "==> 3. 准备自定义 outbound 优化..."
	rm -rf ./outbound
	git clone --depth=1 -b perf/complete-optimizations https://github.com/olicesx/outbound.git ./outbound
	
	@echo "==> 4. 注入 Go Mod 替换规则..."
	cd wing && go mod edit -replace github.com/daeuniverse/outbound=../outbound
	cd wing && CGO_ENABLED=0 go mod tidy

	@echo "==> 5. 源码级路径强力注入..."
	grep -rl "/usr/local/share/dae-wing" wing | xargs -r sed -i 's|/usr/local/share/dae-wing|/usr/share/daed|g'
	grep -rl "/usr/share/v2ray" wing | xargs -r sed -i 's|/usr/share/v2ray|/usr/share/daed|g'
	find wing -name "*.go" -exec sed -i 's|Join(.*, "wing.db")|"/etc/daed/wing.db"|g' {} +
	find wing -name "*.go" -exec sed -i 's|"wing.db"|"/etc/daed/wing.db"|g' {} +

	@echo "==> 6. 编译 eBPF..."
	$(MAKE) -C wing/dae-core ebpf CLANG=$(CLANG) STRIP=$(STRIP) TARGET=bpfel,bpfeb

	@echo "==> 7. 准备前端资源嵌入..."
	rm -rf wing/webrender/web
	mkdir -p wing/webrender/web
	cp -r dist/* wing/webrender/web/
	
	cd wing && unset GOOS && unset GOARCH && go generate -v ./...

	@echo "==> 8. 最终编译..."
	mkdir -p $(BIN_DIR)
	$(eval FINAL_LDFLAGS := -s -w -extldflags "-static" \
		-X github.com/daeuniverse/dae-wing/db.AppVersion=$(FULL_VERSION) \
		-X github.com/daeuniverse/dae-wing/db.AppName=daed)

	cd wing && CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -trimpath \
		-ldflags "$(FINAL_LDFLAGS)" \
		-tags "kernelext embedallowed $(shell cat wing/dae-core/.build_tags 2>/dev/null)" \
		-o ../$(BIN_DIR)/$(NAME)-$(GOARCH) .
	
	chmod +x $(BIN_DIR)/$(NAME)-$(GOARCH)
	@echo "==> 编译完成: $(BIN_DIR)/$(NAME)-$(GOARCH)"

NAME := daed
BIN_DIR := bin

D_SHA ?= $(shell git rev-parse --short HEAD)
W_SHA ?= "unknown"
C_SHA ?= "unknown"

FULL_VERSION := daed_$(D_SHA)-wing_$(W_SHA)-core_$(C_SHA)

LDFLAGS := -s -w -extldflags "-static" -X github.com/daeuniverse/daed/internal/version.Version=$(FULL_VERSION)
GOBUILD := CGO_ENABLED=0 go build -v -trimpath -ldflags "$(LDFLAGS)"

.PHONY: all prepare build-web build-backend

all: prepare build-web build-backend

prepare:
	@echo "==> 初始化子模块并切换自定义内核..."
	git submodule update --init --recursive wing
	cd wing && \
		git submodule set-url dae-core https://github.com/olicesx/dae.git && \
		git submodule update --init --remote dae-core
	cd wing/dae-core && \
		git fetch origin optimize/code-quality-fixes && \
		git checkout optimize/code-quality-fixes

build-web:
	@echo "==> 正在编译前端..."
	pnpm install --no-frozen-lockfile
	pnpm build

build-backend:
	@echo "==> 准备自定义优化依赖..."
	rm -rf ./outbound
	git clone --depth=1 -b perf/complete-optimizations https://github.com/olicesx/outbound.git ./outbound
	
	@echo "==> 注入本地替换规则..."
	cd wing && go mod edit -replace github.com/daeuniverse/outbound=../outbound
	cd wing && CGO_ENABLED=0 go mod tidy
	
	@echo "==> 暴力修复 dae-core 源码中的 go:generate (Heredoc 稳定版)..."
	@cat << 'EOF' > patch_ebpf.go
package main
import ("os"; "strings"; "path/filepath")
func main() {
	err := filepath.Walk("wing/dae-core", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".go") {
			b, _ := os.ReadFile(path)
			s := string(b)
			if strings.Contains(s, "bpf2go") {
				s = strings.ReplaceAll(s, "$$BPF_CLANG", "clang")
				s = strings.ReplaceAll(s, "$$BPF_STRIP_FLAG", "strip")
				s = strings.ReplaceAll(s, "$$BPF_CFLAGS", "-O2 -g -Wall")
				s = strings.ReplaceAll(s, "$$BPF_TARGET", "bpf")
				os.WriteFile(path, []byte(s), 0644)
			}
		}
		return nil
	})
	if err != nil { os.Exit(1) }
}
EOF
	@GOOS= GOARCH= go run patch_ebpf.go
	@rm patch_ebpf.go

	@echo "==> 验证修改结果 (Debug)..."
	@grep "go:generate" wing/dae-core/control/control.go || true

	@echo "==> 强制生成 eBPF 绑定..."
	cd wing/dae-core && GOOS= GOARCH= go generate -v ./...

	@echo "==> 生成 wing 辅助代码..."
	cd wing && GOOS= GOARCH= go generate -v ./...

	@echo "==> 最终编译 (GOARCH=$(GOARCH))..."
	mkdir -p $(BIN_DIR)
	cd wing && CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) \
		-tags kernelext \
		-o ../$(BIN_DIR)/$(NAME)-$(GOARCH) .
	chmod +x $(BIN_DIR)/$(NAME)-$(GOARCH)

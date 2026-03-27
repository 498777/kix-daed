package main

import (
	"os"
	"strings"
	"path/filepath"
)

func main() {
	// 遍历 wing/dae-core 目录下的所有 Go 文件
	err := filepath.Walk("wing/dae-core", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".go") {
			b, _ := os.ReadFile(path)
			s := string(b)
			
			// 只处理包含 bpf2go 的 generate 行
			if strings.Contains(s, "go:generate") && strings.Contains(s, "bpf2go") {
				
				// 1. 替换基础变量
				s = strings.ReplaceAll(s, "$BPF_CLANG", "clang")
				s = strings.ReplaceAll(s, "$BPF_STRIP_FLAG", "strip")
				s = strings.ReplaceAll(s, "$BPF_TARGET", "bpf")

				// 2. 彻底移除原有的 -cflags "$BPF_CFLAGS" 部分，避免位移报错
				// 我们需要匹配不同的引用方式（带引号或不带引号）
				s = strings.ReplaceAll(s, "-cflags \"$BPF_CFLAGS\"", "")
				s = strings.ReplaceAll(s, "-cflags $BPF_CFLAGS", "")

				// 3. 规范化参数：确保 C 编译参数放在 -- 后面（这是 bpf2go 的救命稻草）
				// 寻找源文件的结尾（通常是 .c），并在后面加上优化参数
				if !strings.Contains(s, " -- ") {
					// 兼容 control.go 和 trace.go 等不同源文件
					s = strings.ReplaceAll(s, ".c", ".c -- -O2 -g -Wall")
				}

				os.WriteFile(path, []byte(s), 0644)
			}
		}
		return nil
	})
	if err != nil {
		os.Exit(1)
	}
}

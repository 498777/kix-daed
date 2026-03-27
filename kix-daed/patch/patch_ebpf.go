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
			if strings.Contains(s, "bpf2go") {
				// 焊死变量，解决 bpf2go 参数解析位移问题
				s = strings.ReplaceAll(s, "$BPF_CLANG", "clang")
				s = strings.ReplaceAll(s, "$BPF_STRIP_FLAG", "strip")
				s = strings.ReplaceAll(s, "$BPF_CFLAGS", "-O2 -g -Wall")
				s = strings.ReplaceAll(s, "$BPF_TARGET", "bpf")
				os.WriteFile(path, []byte(s), 0644)
			}
		}
		return nil
	})
	if err != nil {
		os.Exit(1)
	}
}

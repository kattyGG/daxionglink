package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法:")
		fmt.Println("  mytool.exe c    # 测速所有服务器，生成 servertested.txt")
		fmt.Println("  mytool.exe q    # 快速连接，使用 qServer 中的第一行服务器")
		os.Exit(1)
	}

	mode := os.Args[1]
	switch mode {
	case "c":
		runCompare()
	case "q":
		runQuick()
	default:
		fmt.Println("未知模式:", mode)
	}
}

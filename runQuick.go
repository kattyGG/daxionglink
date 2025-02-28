package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"daxionglink/protocol" // 根据你的 module 路径进行调整
)

func runQuick() {
	config, _ := LoadConfig()
	servers, err := readLines(config.QServerFile)
	if err != nil || len(servers) == 0 {
		fmt.Println("无法读取", config.QServerFile, "或文件为空")
		return
	}
	server := servers[0] // 使用最快的服务器（排序后第一行）
	err = protocol.GenerateXrayConfig(server, config.Port, "config.json")
	if err != nil {
		fmt.Println("生成 config.json 失败:", err)
		return
	}
	fmt.Println("✅ 已生成 config.json, 正在启动", config.Core, "...")


	// 明确指定当前目录路径
	coreExe := "./" + config.Core
	if runtime.GOOS == "windows" {
		coreExe = ".\\" + config.Core + ".exe"
	}


	cmd := exec.Command(coreExe, "run", "-c", "config.json")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println(coreExe," 运行出错:", err)
	}
}

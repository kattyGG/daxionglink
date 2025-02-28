package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
)

type AppConfig struct {
	Port        int
	QServerFile string
	TestURL     string
	Core        string // 允许用户选择核心 (xray 或 sing-box)
	MaxCon		int
}

func LoadConfig() (*AppConfig, error) {
	// 默认值
	config := &AppConfig{
		Port:        10501,
		QServerFile: "servertested.txt",
		TestURL:     "https://fast.com",
		Core:        "xray", // 默认使用 xray
		MaxCon:      5,
	}

	file, err := os.Open("config.txt")
	if err != nil {
		// 若不存在则采用默认值
		return config, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "port":
			var p int
			fmt.Sscanf(value, "%d", &p)
			if p > 0 {
				config.Port = p
			}
		case "qserver":
			config.QServerFile = value
		case "testurl":
			config.TestURL = value
		case "core":
			if value == "xray" || value == "sing-box" {
				config.Core = value
			} else {
				fmt.Println("⚠️ 无效的核心选择，默认为 xray")
			}
		case "connection":
			// 将字符串转换为整数
			if maxCon, err := strconv.Atoi(value); err == nil {
				config.MaxCon = maxCon
			} else {
				fmt.Println("⚠️ 无效的连接数，使用默认值")
			}
		}
	}
	return config, scanner.Err()
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"path/filepath"
	"os/exec"
	"runtime"
	"sort"
	"sync"
	"time"
	"daxionglink/protocol" // 根据你的 module 路径进行调整
    "net"
	"net/url"
	"github.com/oschwald/geoip2-golang"
	
)

// ServerResult 保存单个服务器测速结果
type ServerResult struct {
	Link  string
	Speed int64 // 毫秒，0 表示测速失败
	Msg   string
}
func runCompare() {
	servers, err := readLines("server.txt")
	if err != nil || len(servers) == 0 {
		fmt.Println("无法读取 server.txt 或文件为空:", err)
		return
	}
	// 测试网址，默认使用 google.com
	//testURL := "https://www.google.com"

	config, _ := LoadConfig()
	basePort := config.Port
	testURL := config.TestURL
	maxConcurrency := config.MaxCon


	//--------------------------------
	// 解析 testURL 得到 testedSite
	u, err := url.Parse(testURL)
	testedSite := "unknown"
	if err == nil {
		testedSite = u.Host
	}

	// 使用正则表达式清理 testedSite，使其仅包含字母、数字、下划线和破折号
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	safeTestedSite := re.ReplaceAllString(testedSite, "_")

	dateStr := time.Now().Format("20060102")
	testedFileName := fmt.Sprintf("servertested_%s_%s.txt", safeTestedSite, dateStr)
	//--------------------------------

	
	//maxConcurrency := 5
	ch := make(chan string, len(servers))
	for _, s := range servers {
		ch <- s
	}
	close(ch)

	var wg sync.WaitGroup
	resultsMutex := &sync.Mutex{}
	var results []ServerResult

	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			port := basePort + workerID
			for link := range ch {
				configFile := fmt.Sprintf("config_%d.json", workerID)
				err := protocol.GenerateXrayConfig(link, port, configFile)
				if err != nil {
					addResult(&results, resultsMutex, ServerResult{Link: link, Speed: 0, Msg: fmt.Sprintf("生成配置失败: %v", err)})
					continue
				}
				xrayExe := "xray"
				if runtime.GOOS == "windows" {
					xrayExe = "./xray.exe"
				}
				cmd := exec.Command(xrayExe, "run", "-c", configFile)
				stdout, _ := cmd.StdoutPipe()
				stderr, _ := cmd.StderrPipe()
				if err := cmd.Start(); err != nil {
					addResult(&results, resultsMutex, ServerResult{Link: link, Speed: 0, Msg: fmt.Sprintf("Xray 启动失败: %v", err)})
					continue
				}
				go io.Copy(os.Stdout, stdout)
				go io.Copy(os.Stderr, stderr)
				time.Sleep(800 * time.Millisecond)
				duration, speedErr := protocol.MeasureSpeed(port, testURL)
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
				if speedErr != nil {
					addResult(&results, resultsMutex, ServerResult{Link: link, Speed: 0, Msg: fmt.Sprintf("测速失败: %v", speedErr)})
				} else {
					addResult(&results, resultsMutex, ServerResult{Link: link, Speed: duration, Msg: "OK"})
				}
				os.Remove(configFile)
			}
		}(i)
	}
	wg.Wait()

	// 对结果按测速耗时排序（非零速度越低越快）
	sort.Slice(results, func(i, j int) bool {
		if results[i].Speed == 0 {
			return false
		}
		if results[j].Speed == 0 {
			return true
		}
		return results[i].Speed < results[j].Speed
	})

	// 生成 servertested.txt：只写链接，一行一个（覆盖旧文件）
	fTested, err := os.Create(testedFileName)
	if err != nil {
		fmt.Println("无法创建", testedFileName ,":", err)
		return
	}
	writerTested := bufio.NewWriter(fTested)
	for _, r := range results {
		if r.Speed == 0 { // 如果测速失败，不写入
			continue
		}
		writerTested.WriteString(fmt.Sprintf("%s\n", r.Link))
	}
	writerTested.Flush()
	//fTested.Close()


	//				 复制内容到第二个文件
	file2, err := os.Create("servertested.txt")
	if err != nil {
		fmt.Println("无法创建 servertested.txt:", err)
		fTested.Close()
		return
	}
	defer file2.Close()

	// 将 fTested 的指针重新定位到文件开头，然后复制内容到 file2
	if _, err = fTested.Seek(0, 0); err != nil {
		fmt.Println("文件定位错误：", err)
		fTested.Close()
		return
	}
	if _, err = io.Copy(file2, fTested); err != nil {
		fmt.Println("复制文件错误：", err)
		fTested.Close()
		return
	}
	fTested.Close()
	//				复制内容到第二个文件

	// 读取数据库
	// 获取可执行文件所在目录
	ex, err := os.Executable()
	if err != nil {
		fmt.Println("获取可执行文件路径失败:", err)
		return
	}
	exPath := filepath.Dir(ex)
	dbPath := filepath.Join(exPath, "GeoLite2-City.mmdb")

	db, err := geoip2.Open(dbPath)
	if err != nil {
		fmt.Println("无法打开 GeoIP 数据库:", err)
		// 继续执行，db 置为 nil
		db = nil
	} else {
		defer db.Close()
	}




	// 生成 result.txt：写入详细测速信息（覆盖旧文件）
	fResult, err := os.Create("result.txt")
	if err != nil {
		fmt.Println("无法创建 result.txt:", err)
		return
	}
	writerResult := bufio.NewWriter(fResult)
	now := time.Now().Format("2006-01-02 15:04:05")
	for _, r := range results {

		// 初始化国家和地区为未知
		country, region := "未知", "未知"

		// 使用 net/url 解析 r.Link
		u, err := url.Parse(r.Link)
		if err != nil {
			fmt.Printf("解析 URL %s 失败: %v\n", r.Link, err)
		} else {
			host := u.Host
			if host == "" {
				// 如果 Host 为空，可以尝试使用 Opaque 字段（某些非标准链接可能如此）
				host = u.Opaque
			}
			// 分离主机名和端口
			hostname, _, err := net.SplitHostPort(host)
			if err != nil {
				// 如果没有端口，直接使用 host
				hostname = host
			}
			// 尝试解析 IP
			ip := net.ParseIP(hostname)
			if ip == nil {
				ips, err := net.LookupIP(hostname)
				if err == nil && len(ips) > 0 {
					ip = ips[0]
				}
			}
			// 打印调试信息
			//fmt.Printf("处理 link=%s, 解析后 ip=%v\n", r.Link, ip)

			// 根据 IP 查询 GeoIP 信息
			if ip != nil && db != nil {
				record, err := db.City(ip)
				if err != nil {
					fmt.Printf("查询 IP %v 错误: %v\n", ip, err)
				} else {
					if name, ok := record.Country.Names["en"]; ok {
						country = name
					}
					if len(record.Subdivisions) > 0 {
						if name, ok := record.Subdivisions[0].Names["en"]; ok {
							region = name
						}
					}
				}
			}
			
		}
		// 初始化国家和地区为未知 END

		//fmt.Printf("处理 link=%s, 解析后 ip=%v\n", r.Link, ip)

		//line := fmt.Sprintf("[%s] link=%s speed=%dms msg=%s\n", now, r.Link, r.Speed, r.Msg)
		line := fmt.Sprintf("[%s] link=%s speed=%dms msg=%s country=%s region=%s\n",
        now, r.Link, r.Speed, r.Msg, country, region)
		writerResult.WriteString(line)
	}
	writerResult.Flush()
	fResult.Close()

	fmt.Println("测速完成，结果已写入 servertested.txt 和 result.txt")
}

func addResult(results *[]ServerResult, mutex *sync.Mutex, result ServerResult) {
	mutex.Lock()
	*results = append(*results, result)
	mutex.Unlock()
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"rtm-proxy/client"
	"rtm-proxy/collector"
	"rtm-proxy/config"
	"rtm-proxy/executor"
	"rtm-proxy/wsclient"
)

var (
	cfgPath    = flag.String("c", "", "配置文件路径")
	versionFlg = flag.Bool("v", false, "显示版本")
)

// 这些变量由编译时 -ldflags -X 注入
var (
	agentVersion = "1.0.0"
	buildTime    = "unknown"
	goVersion    = "unknown"
)

func init() {
	// 设置全局时区为东八区（不依赖系统 zoneinfo 文件）
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*60*60)
	}
	time.Local = loc
}

func main() {
	flag.Parse()

	if *versionFlg {
		fmt.Printf("rtm-proxy version %s (built: %s, go: %s)\n", agentVersion, buildTime, goVersion)
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 确定本机 IP
	ip := cfg.IP
	if ip == "" {
		ip = getLocalIP()
	}

	// 确定主机名
	hostName := cfg.HostName
	if hostName == "" {
		hostName, _ = os.Hostname()
	}

	log.Printf("[RTM Agent] 启动...")
	log.Printf("  版本: %s", agentVersion)
	log.Printf("  平台: %s", cfg.PlatformURL)
	log.Printf("  IP: %s", ip)
	log.Printf("  主机名: %s", hostName)

	// 创建 HTTP 客户端
	httpClient := client.New(cfg.PlatformURL, ip)

	// 首次注册（或更新硬件信息）
	if err := doRegister(httpClient, cfg, ip, hostName); err != nil {
		log.Printf("注册失败: %v，将继续运行", err)
	}

	// 启动WebSocket客户端
	wsClient := wsclient.New(cfg, ip)
	wsClient.Start()
	defer wsClient.Stop()

	// 创建进程执行器
	cmdMap := make(map[string]string)
	for _, p := range cfg.Processes {
		cmdMap[p.Name] = p.ExecCmd
	}
	exec := executor.New(cmdMap)

	// 提取需要监控的进程名列表
	procNames := make([]string, len(cfg.Processes))
	for i, p := range cfg.Processes {
		procNames[i] = p.Name
	}

	// 启动定时任务
	var wg sync.WaitGroup
	ctx := make(chan struct{})

	// 心跳+健康度上报任务（合并，3秒一次）
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(cfg.HeartbeatInterval) * time.Second)
		defer ticker.Stop()
		heartbeatAndMetrics(httpClient, ip, hostName, procNames, cfg.HeartbeatInterval, ticker, ctx)
	}()

	// 指令轮询任务
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
		defer ticker.Stop()
		pollCommands(httpClient, ip, exec, ticker, ctx)
	}()

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("[RTM Agent] 正在退出...")
	close(ctx)
	wg.Wait()
	log.Println("[RTM Agent] 已退出")
}

// doRegister 首次注册
func doRegister(c *client.Client, cfg *config.Config, ip, hostName string) error {
	sysInfo, err := collector.CollectSystemInfo()
	if err != nil {
		log.Printf("采集系统信息失败: %v", err)
	}

	req := &client.RegisterRequest{
		IP:           ip,
		HostName:     hostName,
		OsType:       sysInfo.OsType,
		OsVersion:    sysInfo.OsVersion,
		CpuArch:      sysInfo.CpuArch,
		CpuCores:     sysInfo.CpuCores,
		CpuModel:     sysInfo.CpuModel,
		MemTotalMb:   sysInfo.MemTotalMb,
		DiskTotalGb:  sysInfo.DiskTotalGb,
		GpuInfo:      sysInfo.GpuInfo,
		MacAddress:   sysInfo.MacAddress,
		ProxyVersion: agentVersion,
	}

	resp, err := c.Register(req)
	if err != nil {
		return err
	}
	log.Printf("注册结果: registered=%v, message=%s", resp.Registered, resp.Message)
	return nil
}

// heartbeatAndMetrics 合并心跳和健康度上报（3秒一次）
func heartbeatAndMetrics(c *client.Client, ip, hostName string, procNames []string, heartbeatInterval int, ticker *time.Ticker, stopCh chan struct{}) {
	for {
		select {
		case <-ticker.C:
			// 采集系统指标
			metrics, err := collector.CollectSystemMetrics()
			if err != nil {
				log.Printf("采集系统指标失败: %v", err)
				// 继续上报心跳（不含指标）
				metrics = &collector.SystemMetrics{Timestamp: time.Now().Unix()}
			}

			// 采集进程指标
			procMetrics := collector.CollectProcessMetrics(procNames)
			clientProcs := make([]client.ProcessMetric, len(procMetrics))
			for i, p := range procMetrics {
				clientProcs[i] = client.ProcessMetric{
					Process:    p.Process,
					Status:     p.Status,
					CpuUsage:   p.CpuUsage,
					MemUsageMb: p.MemUsageMb,
				}
			}

			req := &client.HeartbeatRequest{
				IP:           ip,
				HostName:     hostName,
				ProxyVersion:      agentVersion,
				HeartbeatInterval: heartbeatInterval,
				CpuUsage:          metrics.CpuUsage,
				MemUsage:     metrics.MemUsage,
				MemUsedMb:    metrics.MemUsageMb,
				DiskUsage:    metrics.DiskUsage,
				NetInKbps:    metrics.NetIn,
				NetOutKbps:   metrics.NetOut,
				LoadAvg:      metrics.LoadAvg,
				Timestamp:    metrics.Timestamp,
				Processes:    clientProcs,
			}
			if err := c.Heartbeat(req); err != nil {
				log.Printf("心跳上报失败: %v", err)
			}
		case <-stopCh:
			return
		}
	}
}

// pollCommands 定时轮询指令
func pollCommands(c *client.Client, ip string, exec *executor.Executor, ticker *time.Ticker, stopCh chan struct{}) {
	for {
		select {
		case <-ticker.C:
			commands, err := c.Poll()
			if err != nil {
				log.Printf("轮询指令失败: %v", err)
				continue
			}
			if len(commands) == 0 {
				continue
			}

			for _, cmd := range commands {
				log.Printf("收到指令: type=%s, process=%s, cmdId=%s", cmd.Type, cmd.Process, cmd.CommandID)
				success, msg := exec.Execute(cmd.Type, cmd.Process, cmd.Params)
				log.Printf("指令执行结果: success=%v, message=%s", success, msg)

				ack := &client.AckRequest{
					IP:        ip,
					CommandID: cmd.CommandID,
					Success:   success,
					Message:   msg,
				}
				if err := c.Ack(ack); err != nil {
					log.Printf("上报指令结果失败: %v", err)
				}
			}
		case <-stopCh:
			return
		}
	}
}

// getLocalIP 获取本机非回环 IP
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Response 平台通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

// RegisterRequest Proxy 注册请求
type RegisterRequest struct {
	IP           string `json:"ip"`
	HostName     string `json:"hostName"`
	OsType       string `json:"osType"`
	OsVersion    string `json:"osVersion"`
	CpuArch      string `json:"cpuArch"`
	CpuCores     int    `json:"cpuCores"`
	CpuModel     string `json:"cpuModel"`
	MemTotalMb   uint64 `json:"memTotalMb"`
	DiskTotalGb  uint64 `json:"diskTotalGb"`
	GpuInfo      string `json:"gpuInfo"`
	MacAddress   string `json:"macAddress"`
	ProxyVersion string `json:"proxyVersion"`
}

// RegisterResponse Proxy 注册响应
type RegisterResponse struct {
	Registered bool   `json:"registered"`
	Message    string `json:"message"`
}

// HeartbeatRequest 心跳请求（合并健康度数据）
type HeartbeatRequest struct {
	IP                string            `json:"ip"`
	HostName          string            `json:"hostName"`
	ProxyVersion      string            `json:"proxyVersion"`
	HeartbeatInterval int               `json:"heartbeatInterval"`
	// 健康度数据
	CpuUsage   float64           `json:"cpuUsage"`
	MemUsage   float64           `json:"memUsage"`
	MemUsedMb  float64           `json:"memUsedMb"`
	DiskUsage  float64           `json:"diskUsage"`
	NetInKbps  float64           `json:"netInKbps"`
	NetOutKbps float64           `json:"netOutKbps"`
	LoadAvg    string            `json:"loadAvg"`
	Timestamp  int64             `json:"timestamp"`
	// 进程指标（可选）
	Processes []ProcessMetric   `json:"processes"`
}

// ProcessMetric 进程指标
type ProcessMetric struct {
	Process    string  `json:"process"`
	Status     string  `json:"status"`
	CpuUsage   float64 `json:"cpuUsage"`
	MemUsageMb float64 `json:"memUsageMb"`
}

// PollRequest 轮询请求
type PollRequest struct {
	IP string `json:"ip"`
}

// PollResponse 轮询响应
type PollResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    PollData `json:"data"`
}

// PollData 轮询数据
type PollData struct {
	Commands []ProxyCommand `json:"commands"`
}

// ProxyCommand Proxy 指令
type ProxyCommand struct {
	CommandID  string `json:"commandId"`
	Type       string `json:"type"`
	Process    string `json:"process"`
	Params     string `json:"params"`
	CreateTime string `json:"createTime"`
}

// AckRequest 指令执行结果上报
type AckRequest struct {
	IP        string `json:"ip"`
	CommandID string `json:"commandId"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
}

// Client HTTP 客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
	ip         string
}

// New 创建客户端
func New(baseURL, ip string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		ip: ip,
	}
}

// Register 向平台注册
func (c *Client) Register(req *RegisterRequest) (*RegisterResponse, error) {
	resp, err := c.post("/api/v1/proxy/register", req)
	if err != nil {
		return nil, err
	}
	var result RegisterResponse
	if data, ok := resp.Data.(map[string]interface{}); ok {
		if v, ok := data["registered"].(bool); ok {
			result.Registered = v
		}
		if v, ok := data["message"].(string); ok {
			result.Message = v
		}
	}
	return &result, nil
}

// Heartbeat 发送心跳（合并健康度数据）
func (c *Client) Heartbeat(req *HeartbeatRequest) error {
	_, err := c.post("/api/v1/proxy/heartbeat", req)
	return err
}

// Poll 轮询待执行指令
func (c *Client) Poll() ([]ProxyCommand, error) {
	resp, err := c.post("/api/v1/proxy/poll", &PollRequest{IP: c.ip})
	if err != nil {
		return nil, err
	}

	// 解析嵌套响应
	dataBytes, _ := json.Marshal(resp.Data)
	var pollData PollData
	if err := json.Unmarshal(dataBytes, &pollData); err != nil {
		// 尝试直接解析为 commands 数组
		var cmds []ProxyCommand
		if err2 := json.Unmarshal(dataBytes, &cmds); err2 == nil {
			return cmds, nil
		}
		return nil, fmt.Errorf("解析轮询响应失败: %w", err)
	}
	return pollData.Commands, nil
}

// Ack 上报指令执行结果
func (c *Client) Ack(req *AckRequest) error {
	_, err := c.post("/api/v1/proxy/ack", req)
	return err
}

func (c *Client) post(path string, body interface{}) (*Response, error) {
	url := c.baseURL + path
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 200 && result.Code != 0 {
		return nil, fmt.Errorf("平台返回错误: %s", result.Message)
	}

	return &result, nil
}

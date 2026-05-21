package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"rtm-server/internal/constant"
	"rtm-server/internal/model"
)

func (e *AlarmEngine) sendWechatNotification(ctx context.Context, record *model.AlarmRecord, ch model.AlarmChannel) (bool, string) {
	var cfg struct {
		Webhook string `json:"webhook"`
	}
	if err := json.Unmarshal([]byte(ch.Config), &cfg); err != nil {
		return e.writeNotifyLog(ctx, record, ch, false, "解析通道配置失败: "+err.Error())
	}

	if cfg.Webhook == "" {
		return e.writeNotifyLog(ctx, record, ch, false, "Webhook未配置")
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": record.Content,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return e.writeNotifyLog(ctx, record, ch, false, "序列化请求失败: "+err.Error())
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(cfg.Webhook, constant.ContentTypeJSON, bytes.NewReader(body))
	if err != nil {
		return e.writeNotifyLog(ctx, record, ch, false, "请求失败: "+err.Error())
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return e.writeNotifyLog(ctx, record, ch, false, "解析响应失败: "+err.Error())
	}

	if result.ErrCode != 0 {
		respStr := fmt.Sprintf("企业微信错误[%d]: %s", result.ErrCode, result.ErrMsg)
		if len(respStr) > 1024 {
			respStr = respStr[:1024]
		}
		return e.writeNotifyLog(ctx, record, ch, false, respStr)
	}

	return e.writeNotifyLog(ctx, record, ch, true, "发送成功")
}

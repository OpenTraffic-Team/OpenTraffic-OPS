package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"rtm-server/internal/constant"
	"rtm-server/internal/model"
)

func (e *AlarmEngine) sendDingTalkNotification(ctx context.Context, record *model.AlarmRecord, ch model.AlarmChannel) (bool, string) {
	var cfg struct {
		Webhook string `json:"webhook"`
		Secret  string `json:"secret"`
	}
	if err := json.Unmarshal([]byte(ch.Config), &cfg); err != nil {
		return e.writeNotifyLog(ctx, record, ch, false, "解析通道配置失败: "+err.Error())
	}

	if cfg.Webhook == "" {
		return e.writeNotifyLog(ctx, record, ch, false, "Webhook未配置")
	}

	webhookURL := cfg.Webhook
	if cfg.Secret != "" {
		timestamp, sign := dingTalkSign(cfg.Secret)
		sep := "?"
		if strings.Contains(cfg.Webhook, "?") {
			sep = "&"
		}
		webhookURL = fmt.Sprintf("%s%stimestamp=%s&sign=%s", cfg.Webhook, sep, timestamp, url.QueryEscape(sign))
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
	resp, err := client.Post(webhookURL, constant.ContentTypeJSON, bytes.NewReader(body))
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
		respStr := fmt.Sprintf("钉钉错误[%d]: %s", result.ErrCode, result.ErrMsg)
		if len(respStr) > 1024 {
			respStr = respStr[:1024]
		}
		return e.writeNotifyLog(ctx, record, ch, false, respStr)
	}

	respStr := fmt.Sprintf("发送成功: %s", result.ErrMsg)
	if len(respStr) > 1024 {
		respStr = respStr[:1024]
	}
	return e.writeNotifyLog(ctx, record, ch, true, respStr)
}

func dingTalkSign(secret string) (timestamp, sign string) {
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(ts + "\n" + secret))
	return ts, base64.StdEncoding.EncodeToString(h.Sum(nil))
}

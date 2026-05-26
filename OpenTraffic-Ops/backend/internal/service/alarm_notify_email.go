package service

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strings"

	"go.uber.org/zap"
	"opentraffic-ops-backend/internal/model"
)

func (e *AlarmEngine) sendEmailNotification(ctx context.Context, record *model.AlarmRecord, ch model.AlarmChannel) (bool, string) {
	var cfg struct {
		SmtpHost  string `json:"smtpHost"`
		SmtpPort  string `json:"smtpPort"`
		FromEmail string `json:"fromEmail"`
		Password  string `json:"password"`
		ToEmails  string `json:"toEmails"`
	}
	if err := json.Unmarshal([]byte(ch.Config), &cfg); err != nil {
		return e.writeNotifyLog(ctx, record, ch, false, "解析通道配置失败: "+err.Error())
	}

	if cfg.SmtpHost == "" || cfg.SmtpPort == "" || cfg.FromEmail == "" || cfg.Password == "" || cfg.ToEmails == "" {
		return e.writeNotifyLog(ctx, record, ch, false, "邮件配置不完整")
	}

	toList := strings.Split(cfg.ToEmails, ",")
	var recipients []string
	for _, t := range toList {
		t = strings.TrimSpace(t)
		if t != "" {
			recipients = append(recipients, t)
		}
	}
	if len(recipients) == 0 {
		return e.writeNotifyLog(ctx, record, ch, false, "收件人列表为空")
	}

	subject := "[告警通知] " + record.RuleName
	body := record.Content

	if err := sendMail(cfg.SmtpHost, cfg.SmtpPort, cfg.FromEmail, cfg.Password, recipients, subject, body); err != nil {
		zap.L().Error("发送邮件告警失败", zap.Error(err), zap.Int64("recordId", record.ID))
		return e.writeNotifyLog(ctx, record, ch, false, "发送失败: "+err.Error())
	}

	return e.writeNotifyLog(ctx, record, ch, true, "发送成功")
}

func sendMail(host, port, from, password string, to []string, subject, body string) error {
	addr := host + ":" + port

	var client *smtp.Client
	var err error

	if port == "465" {
		tlsConfig := &tls.Config{
			ServerName: host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("tls dial failed: %w", err)
		}

		client, err = smtp.NewClient(conn, host)
		if err != nil {
			conn.Close()
			return fmt.Errorf("smtp client failed: %w", err)
		}
		defer client.Close()
	} else {
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("smtp dial failed: %w", err)
		}
		defer client.Close()

		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName: host,
			}
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("starttls failed: %w", err)
			}
		}
	}

	auth := smtp.PlainAuth("", from, password, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail from failed: %w", err)
	}
	for _, rcpt := range to {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("rcpt failed: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data failed: %w", err)
	}

	msg := buildRFC822(from, to, subject, body)
	_, err = w.Write(msg)
	if err != nil {
		w.Close()
		return fmt.Errorf("write failed: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close data failed: %w", err)
	}

	return client.Quit()
}

func buildRFC822(from string, to []string, subject, body string) []byte {
	b64Subject := base64.StdEncoding.EncodeToString([]byte(subject))
	encodedSubject := "=?UTF-8?B?" + b64Subject + "?="

	var sb strings.Builder
	sb.WriteString("From: " + from + "\r\n")
	sb.WriteString("To: " + strings.Join(to, ", ") + "\r\n")
	sb.WriteString("Subject: " + encodedSubject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	return []byte(sb.String())
}

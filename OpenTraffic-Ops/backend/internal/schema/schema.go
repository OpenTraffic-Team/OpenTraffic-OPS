// Package schema 提供数据库表结构自动初始化与默认数据填充。
// 启动时检测期望表是否齐全：全新库自动执行内嵌 DDL 建表并创建默认
// admin 账号；表部分缺失时仅告警不改动数据，避免 DROP 误删。
package schema

import (
	"embed"
	"fmt"
	"sort"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"opentraffic-ops-backend/internal/model"
	"opentraffic-ops-backend/pkg/crypto"
)

//go:embed all:sql
var ddlFS embed.FS

// ddlFiles 建表脚本，按依赖顺序执行
var ddlFiles = []string{
	"sql/01_sys_tables.sql",
	"sql/02_bu_tables.sql",
	"sql/03_chat_tables.sql",
	"sql/04_alarm_tables.sql",
}

// expectedTables 平台全部期望表（与内嵌 DDL 一致）
var expectedTables = []string{
	"sys_user",
	"sys_oper_log",
	"sys_login_log",
	"bu_host_info",
	"bu_host_health",
	"bu_chat_session",
	"bu_chat_message",
	"bu_alarm_channel",
	"bu_alarm_rule",
	"bu_alarm_record",
	"bu_alarm_notify_log",
}

const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "admin123"
)

// Init 检查并初始化数据库表结构，随后填充默认数据
func Init(db *gorm.DB) error {
	missing, err := missingTables(db)
	if err != nil {
		return fmt.Errorf("failed to check existing tables: %w", err)
	}

	switch {
	case len(missing) == 0:
		zap.L().Info("Database schema is complete, skipping DDL initialization")
	case len(missing) == len(expectedTables):
		zap.L().Info("Empty database detected, initializing schema from embedded DDL")
		if err := executeDDL(db); err != nil {
			return err
		}
		zap.L().Info("Database schema initialized", zap.Int("tables", len(expectedTables)))
	default:
		sort.Strings(missing)
		zap.L().Warn("Database schema is incomplete; skipping DDL initialization to avoid data loss, please create missing tables manually",
			zap.Strings("missing_tables", missing))
		return nil
	}

	return seedAdmin(db)
}

// missingTables 返回当前库的默认 schema 中不存在的期望表
func missingTables(db *gorm.DB) ([]string, error) {
	var existing []string
	if err := db.Raw(
		"SELECT table_name FROM information_schema.tables WHERE table_schema = current_schema() AND table_name IN ?",
		expectedTables,
	).Scan(&existing).Error; err != nil {
		return nil, err
	}

	set := make(map[string]bool, len(existing))
	for _, t := range existing {
		set[t] = true
	}
	var missing []string
	for _, t := range expectedTables {
		if !set[t] {
			missing = append(missing, t)
		}
	}
	return missing, nil
}

// executeDDL 按顺序执行全部内嵌建表脚本
func executeDDL(db *gorm.DB) error {
	for _, name := range ddlFiles {
		data, err := ddlFS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("failed to read embedded DDL %s: %w", name, err)
		}
		if err := db.Exec(string(data)).Error; err != nil {
			return fmt.Errorf("failed to execute DDL %s: %w", name, err)
		}
	}
	return nil
}

// seedAdmin 在用户表为空时创建默认 admin 账号（admin/admin123）
func seedAdmin(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.SysUser{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count sys_user: %w", err)
	}
	if count > 0 {
		return nil
	}

	hash, err := crypto.HashPassword(defaultAdminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash default admin password: %w", err)
	}

	admin := &model.SysUser{
		UserName: defaultAdminUsername,
		NickName: "系统管理员",
		Password: hash,
	}
	admin.CreateBy = "system"
	admin.Remark = "系统初始化默认账号"

	if err := db.Create(admin).Error; err != nil {
		return fmt.Errorf("failed to create default admin user: %w", err)
	}

	zap.L().Warn("Default admin account created, please change the password after first login",
		zap.String("username", defaultAdminUsername))
	return nil
}

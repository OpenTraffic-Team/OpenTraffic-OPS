package repository

import (
	"fmt"
	"opentraffic-ops-init-backend/internal/model"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库实例
var DB *gorm.DB

// InitDatabase 初始化数据库
func InitDatabase(dataDir string) error {
	var err error
	dbPath := fmt.Sprintf("%s/rtm_init.db", dataDir)

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移
	err = DB.AutoMigrate(
		&model.Component{},
		&model.User{},
		&model.AuditLog{},
		&model.Server{},
		&model.DeployRecord{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 创建默认管理员用户
	createDefaultAdmin()

	return nil
}

// createDefaultAdmin 创建默认管理员用户
func createDefaultAdmin() {
	var count int64
	DB.Model(&model.User{}).Count(&count)
	if count == 0 {
		admin := &model.User{
			ID:       generateUUID(),
			Username: "admin",
			Role:     model.UserRoleAdmin,
		}
		// 默认密码: admin123 (需要在启动时由服务层设置hash)
		DB.Create(admin)
	}
}

// generateUUID 生成UUID
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

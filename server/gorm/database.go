package gorm

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建数据库目录失败：%w", err)
	}

	// 打开数据库连接
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("打开数据库失败：%w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败：%w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败：%w", err)
	}

	DB = db // 赋值全局数据库实例
	// 数据库迁移
	if err := runMigrations(); err != nil {
		return fmt.Errorf("数据库迁移失败：%w", err)
	}
	return nil
}

func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		sqlDB.Close()
	}
}

func (ChannelTag) TableName() string {
	return "channel_tag"
}

func runMigrations() error {
	if err := DB.AutoMigrate(&Tag{}, &Channel{}, &ChannelTag{}); err != nil {
		return fmt.Errorf("数据库迁移失败：%w", err)
	}
	return nil
}

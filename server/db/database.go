// 数据初始化文件：包含数据库连接、关闭、迁移、辅助事务，返回数据库实例
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB // 全局数据库实例

func InitDatabase(dbPath string) error {
	// 创建数据库目录（如果不存在）
	dir := filepath.Dir(dbPath)                     // 获取数据库文件所在目录，Dir方法删除文件路径的最后一个元素，返回目录路径
	if err := os.MkdirAll(dir, 0o755); err != nil { // MkdirAll方法会创建指定路径的目录，包括所有必要的父目录，然后返回nil。如果路径已经存在，MkdirAll会返回nil。
		return fmt.Errorf("创建数据库目录失败：%w", err)
	}
	log.Printf("数据库路径：%s\n", dbPath)
	// 打开数据库连接
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败：%w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Printf("数据库连接成功\n")
	// 测试数据库连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败：%w", err)
	}
	log.Printf("数据库连接测试成功\n")
	DB = db // 赋值全局数据库实例
	// 数据库迁移
	if err := runMigrations(); err != nil {
		return fmt.Errorf("数据库迁移失败：%w", err)
	}
	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// 数据库迁移函数，执行数据库初始化操作，创建必要的表
func runMigrations() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL
		);
		CREATE TABLE IF NOT EXISTS channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL
		);
		CREATE TABLE IF NOT EXISTS channel_tag (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			channel_id INTEGER,
			tag_id INTEGER,
			FOREIGN KEY (channel_id) REFERENCES channels(id),
			FOREIGN KEY (tag_id) REFERENCES tags(id)
		);
	`)
	if err != nil {
		return fmt.Errorf("数据库迁移失败：%w", err)
	}
	log.Printf("库表创建成功\n")
	return nil
}

// 事务处理辅助函数
func WithTransaction(fn func(tx *sql.Tx) error) error {
	// 开始事务
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("创建事务失败：%w", err)
	}
	//
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // 重新触发panic
		}
	}()
	//
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("事务执行失败：%w", err)
	}
	// 提交事务
	return tx.Commit()
}

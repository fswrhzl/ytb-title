/*
Slog 日志中间件：文件日志
定期清理30天前的文件
文件大小超过10M进行切割
*/
package middleware

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// 日志管理器结构体
type LogManager struct {
	logger  *slog.Logger
	logFile *os.File
	logDir  string
}

// 初始化日志管理器
func (lm *LogManager) Init() error {
	lm.logDir = "./logs"
	// 创建日志目录
	if err := os.MkdirAll(lm.logDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 生成日志文件路径：使用日期作为文件名
	logFilePath := filepath.Join(lm.logDir, time.Now().Format("2006-01-02")+".log")

	// 创建或打开日志文件
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	lm.logFile = file

	// 设置日志输出
	handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		AddSource:   true,           // 是否记录输出日志的位置
		Level:       slog.LevelInfo, // 日志级别
		ReplaceAttr: nil,
	})
	lm.logger = slog.New(handler)

	// 启动日志轮转
	go lm.logRotate(logFilePath)

	return nil
}

// 日志轮转：每天检查并轮换日志文件
// TODO：待优化：1. 如果每24小时检查一次，即使文件超过10M，但是已经进入新的一天，日志会写入新的文件，再去分割前一天的日志文件没有意义。检查循环的周期应该更短，但是更短的话，在并发情况下，是否会影响日志写入？
// TODO：2. 对旧日志文件进行压缩
func (lm *LogManager) logRotate(logFilePath string) {
	for {
		time.Sleep(24 * time.Hour) // 每24小时检查一次

		// 获取日志文件信息
		info, err := os.Stat(logFilePath)
		if err != nil {
			continue
		}

		// 判断文件大小是否超过10MB
		if info.Size() > 10*1024*1024 { // 10MB
			// 重命名当前日志文件
			newLogFilePath := fmt.Sprintf("%s_%s.log", logFilePath, time.Now().Format("2006-01-02"))
			err = os.Rename(logFilePath, newLogFilePath)
			if err != nil {
				fmt.Println("Error renaming log file:", err)
				continue
			}

			// 创建新的日志文件
			file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				fmt.Println("Error creating new log file:", err)
				continue
			}

			lm.logFile.Close()
			lm.logFile = file

			// 更新日志输出
			handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
				AddSource:   true,           // 是否记录输出日志的位置
				Level:       slog.LevelInfo, // 日志级别
				ReplaceAttr: nil,
			})
			lm.logger = slog.New(handler)
		}

		// 清理超过30天的日志文件
		lm.cleanOldLogs()
	}
}

// 清理超过30天的日志文件
func (lm *LogManager) cleanOldLogs() {
	files, err := os.ReadDir(lm.logDir)
	if err != nil {
		fmt.Println("Error reading log directory:", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 检查文件是否超过30天
		filePath := filepath.Join(lm.logDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		if time.Since(info.ModTime()).Hours() > 24*30 {
			err := os.Remove(filePath)
			if err != nil {
				fmt.Println("Error deleting old log file:", err)
			}
		}
	}
}

// 创建 Gin 日志中间件
func (lm *LogManager) LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 执行请求
		c.Next()

		// 获取日志信息
		statusCode := c.Writer.Status()
		method := c.Request.Method
		url := c.Request.URL.Path
		latency := time.Since(start)

		// 记录日志
		slogAttrs := []slog.Attr{
			slog.String("method", method),
			slog.String("url", url),
			slog.Int("status_code", statusCode),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		}
		if len(c.Errors) > 0 {
			slogAttrs = append(slogAttrs, slog.String("gin error", c.Errors.String()))
			lm.logger.LogAttrs(c.Request.Context(), slog.LevelError, "http request error", slogAttrs...)
			return
		}
		lm.logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "http request", slogAttrs...)
	}
}

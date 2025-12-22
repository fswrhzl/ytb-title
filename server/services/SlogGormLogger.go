package services

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm/logger"
)

type Logger struct {
	logger        *slog.Logger
	level         logger.LogLevel
	slowThreshold time.Duration
}

// New 创建一个 slog-based GORM Logger
func New(
	log *slog.Logger, // slog日志记录器
	level logger.LogLevel, // gorm日志级别，默认Info
	slowThreshold time.Duration, // 慢查询阈值，默认200ms
) *Logger {
	return &Logger{
		logger:        log,
		level:         level,
		slowThreshold: slowThreshold,
	}
}

/*************** gorm logger.Interface 实现 ***************/

func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	// GORM 期望返回一个新的 logger（或自身的拷贝）
	return &Logger{
		logger:        l.logger,
		level:         level,
		slowThreshold: l.slowThreshold,
	}
}

func (l *Logger) Info(ctx context.Context, msg string, data ...any) {
	if l.level >= logger.Info {
		l.logger.InfoContext(ctx, msg, data...)
	}
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...any) {
	if l.level >= logger.Warn {
		l.logger.WarnContext(ctx, msg, data...)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, data ...any) {
	if l.level >= logger.Error {
		l.logger.ErrorContext(ctx, msg, data...)
	}
}

// Trace 是 SQL 日志的核心
func (l *Logger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	// Silent 级别：直接返回
	if l.level == logger.Silent {
		return
	}

	elapsed := time.Since(begin)

	// 1. SQL 执行错误
	if err != nil && l.level >= logger.Error {
		sql, rows := fc()
		l.logger.ErrorContext(
			ctx,
			"gorm sql error",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("cost", elapsed),
			slog.String("error", err.Error()),
		)
		return
	}

	// 2. 慢 SQL
	if l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level >= logger.Warn {
		sql, rows := fc()
		l.logger.WarnContext(
			ctx,
			"gorm slow sql",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("cost", elapsed),
		)
		return
	}

	// 3. 普通 SQL（Info 级别）
	if l.level >= logger.Info {
		sql, rows := fc()
		l.logger.DebugContext(
			ctx,
			"gorm sql",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("cost", elapsed),
		)
	}
}

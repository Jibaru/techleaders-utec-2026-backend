package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// gormSlogLogger adapts GORM's logger.Interface to a *slog.Logger so all
// database logs flow through the same handler as the rest of the app.
type gormSlogLogger struct {
	logger        *slog.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func newGormSlogLogger(logger *slog.Logger) *gormSlogLogger {
	return &gormSlogLogger{
		logger:        logger.With("component", "gorm"),
		level:         gormlogger.Warn,
		slowThreshold: 200 * time.Millisecond,
	}
}

func (l *gormSlogLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

func (l *gormSlogLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		l.logger.InfoContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *gormSlogLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		l.logger.WarnContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *gormSlogLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		l.logger.ErrorContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *gormSlogLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()

	attrs := []any{
		"sql", sql,
		"rows", rows,
		"duration_ms", elapsed.Milliseconds(),
	}

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && l.level >= gormlogger.Error:
		l.logger.ErrorContext(ctx, "query failed", append(attrs, "err", err)...)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		l.logger.WarnContext(ctx, "slow query", attrs...)
	case l.level >= gormlogger.Info:
		l.logger.DebugContext(ctx, "query", attrs...)
	}
}

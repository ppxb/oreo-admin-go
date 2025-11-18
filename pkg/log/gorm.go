package log

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ppxb/oreo-admin-go/pkg/constant"
)

type gormLogger struct {
	Config
	normalStr, normalErrStr, slowStr, slowErrStr string
}

func NewDefaultGormLogger() logger.Interface {
	return NewGormLogger(Config{
		ops: DefaultWrapper.log.Options(),
		gorm: logger.Config{
			SlowThreshold: 200 * time.Millisecond,
		},
	})
}

func NewGormLogger(config Config) logger.Interface {
	var (
		normalStr    = "[%.3fms] [rows:%v] %s"
		slowStr      = "[%.3fms(slow)] [rows:%v] %s"
		normalErrStr = "%s\n[%.3fms] [rows:%v] %s"
		slowErrStr   = "%s\n[%.3fms(slow)] [rows:%v] %s"
	)

	if config.gorm.Colorful {
		normalStr = logger.Green + "[%.3fms] " + logger.Reset + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
		slowStr = logger.Yellow + "[%.3fms(slow)] " + logger.Reset + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
		normalErrStr = logger.RedBold + "%s\n" + logger.Reset + logger.Green + "[%.3fms] " + logger.Reset + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
		slowErrStr = logger.RedBold + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms(slow)] " + logger.Reset + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
	}

	return &gormLogger{
		Config:       config,
		normalStr:    normalStr,
		slowStr:      slowStr,
		normalErrStr: normalErrStr,
		slowErrStr:   slowErrStr,
	}
}

func (l *gormLogger) getLogger(ctx context.Context) Interface {
	return DefaultWrapper.WithContext(ctx).log.WithFields(DefaultWrapper.WithContext(ctx).fields)
}

func (l *gormLogger) getLoggerWithLineNum(ctx context.Context) Interface {
	skipHelper := true
	if v, ok := ctx.Value(constant.LogSkipHelperCtxKey).(bool); ok {
		skipHelper = v
	}
	lineNum := fileWithLineNum(
		l.ops,
		WithSkipGorm(true),
		WithSkipHelper(skipHelper),
	)
	return l.getLogger(ctx).WithFields(map[string]interface{}{
		constant.LogLineNumKey: lineNum,
	})
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.gorm.LogLevel = level
	return &newLogger
}

func (l *gormLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.gorm.LogLevel >= logger.Info {
		l.getLoggerWithLineNum(ctx).Logf(InfoLevel, format, args...)
	}
}

func (l *gormLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	if l.gorm.LogLevel >= logger.Warn {
		l.getLoggerWithLineNum(ctx).Logf(WarnLevel, format, args...)
	}
}

func (l *gormLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.gorm.LogLevel >= logger.Error {
		l.getLoggerWithLineNum(ctx).Logf(ErrorLevel, format, args...)
	}
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.gorm.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	elapsedF := float64(elapsed.Nanoseconds()) / 1e6
	sql, rows := fc()
	row := "-"
	if rows > -1 {
		row = fmt.Sprintf("%d", rows)
	}

	if hiddenSql, ok := ctx.Value(constant.LogHiddenSqlCtxKey).(bool); ok && hiddenSql {
		sql = "(sql is hidden)"
	}

	log := l.getLoggerWithLineNum(ctx)

	switch {
	case l.gorm.LogLevel >= logger.Error && err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		if l.gorm.SlowThreshold > 0 && elapsed > l.gorm.SlowThreshold {
			log.Logf(ErrorLevel, l.slowErrStr, err, elapsedF, row, sql)
		} else {
			log.Logf(ErrorLevel, l.normalErrStr, err, elapsedF, row, sql)
		}
	case l.gorm.LogLevel >= logger.Warn && l.gorm.SlowThreshold > 0 && elapsed > l.gorm.SlowThreshold:
		log.Logf(WarnLevel, l.slowStr, elapsedF, row, sql)
	case l.gorm.LogLevel >= logger.Info:
		log.Logf(InfoLevel, l.normalStr, elapsedF, row, sql)
	}
}

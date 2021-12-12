package el_mysql

import (
	"context"
	"errors"
	"fmt"
	"github.com/drip-in/eden_lib/logs"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type Logger struct {
	logger.LogLevel
	IgnoreRecordNotFoundError bool // ignore gorm.ErrRecordNotFound as error
	*logs.Logger
}

func (l Logger) Apply(config *gorm.Config) error {
	if l.LogLevel == 0 {
		l.LogLevel = logger.Error
	}

	if l.Logger == nil {
		l.Logger = logs.Default()
	}

	config.Logger = l
	return nil
}

func (l Logger) AfterInitialize(db *gorm.DB) error {
	return nil
}

func (l Logger) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = level
	return l
}

func (l Logger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Info {
		fields := l.Logger.ConvertToFields(args)
		l.Logger.Info("GORM LOG "+msg, fields...)
	}
}

func (l Logger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Warn {
		fields := l.Logger.ConvertToFields(args)
		l.Logger.Warn("GORM LOG "+msg, fields...)
	}
}

func (l Logger) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Error {
		fields := l.Logger.ConvertToFields(args)
		l.Logger.Error("GORM LOG "+msg, fields...)
	}
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel > logger.Silent {
		cost := float64(time.Since(begin).Nanoseconds()/1e4) / 100.0
		switch {
		case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
			sql, _ := fc()
			l.Logger.Error("GORM LOG", logs.String("sql", sql), logs.String("err", err.Error()))
		case l.LogLevel >= logger.Info:
			sql, rows /* affected rows */ := fc()
			l.Logger.Info("GORM LOG", logs.String("sql", sql), logs.Int64("rows", rows), logs.String("cost", fmt.Sprintf("%.2fms", cost)))
		}
	}
}


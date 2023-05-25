package el_mysql

import (
	"context"
	"github.com/drip-in/eden_lib/logs"
	"gorm.io/gorm"
	"time"
)

type ConnPool struct {
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
}

func (connPool ConnPool) Apply(*gorm.Config) error {
	return nil
}

func (connPool ConnPool) AfterInitialize(db *gorm.DB) (err error) {
	dbConnPool := db.ConnPool
	if connector, ok := db.ConnPool.(gorm.GetDBConnector); ok {
		if sqlDB, err := connector.GetDBConn(); err == nil {
			dbConnPool = sqlDB
		}
	}

	ctx := context.Background()
	if connPool.ConnMaxIdleTime > 0 {
		if conn, ok := dbConnPool.(interface{ SetConnMaxIdleTime(time.Duration) }); ok && conn != nil {
			conn.SetConnMaxIdleTime(connPool.ConnMaxIdleTime)
		} else {
			logs.CtxWarn(ctx, "GORM Connection Pool: failed to set max idle time, go 1.15+ support this option")
		}
	}

	if connPool.ConnMaxLifetime > 0 {
		if conn, ok := dbConnPool.(interface{ SetConnMaxLifetime(time.Duration) }); ok && conn != nil {
			conn.SetConnMaxLifetime(connPool.ConnMaxLifetime)
		} else {
			logs.CtxWarn(ctx, "GORM Connection Pool: failed to set max lifetime")
		}
	}

	if connPool.MaxIdleConns > 0 {
		if conn, ok := dbConnPool.(interface{ SetMaxIdleConns(int) }); ok && conn != nil {
			conn.SetMaxIdleConns(connPool.MaxIdleConns)
		} else {
			logs.CtxWarn(ctx, "GORM Connection Pool: failed to set max idle conns")
		}
	}

	if connPool.MaxOpenConns > 0 {
		if conn, ok := dbConnPool.(interface{ SetMaxOpenConns(int) }); ok && conn != nil {
			conn.SetMaxOpenConns(connPool.MaxOpenConns)
		} else {
			logs.CtxWarn(ctx, "GORM Connection Pool: failed to set max open conns")
		}
	}
	// ignore conn pool error
	return nil
}

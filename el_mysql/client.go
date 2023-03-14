package el_mysql

import (
	"context"
	"fmt"
	"github.com/drip-in/eden_lib/conf"
	"github.com/drip-in/eden_lib/logs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"sync"
	"time"
)

type DbType string

const (
	Read  = DbType("read")
	Write = DbType("write")

	TRANSACTION_KEY               = "db_transaction_key"
	TRANSACTION_SIGNAL            = "db_transaction_signal" // 事务完成时close掉，当广播用
	TRANSACTION_COMMITTED_MAP     = "db_transaction_committed"
	TRANSACTION_COMMITTED_MAP_KEY = "committed"
)

type DBClient struct {
	db *gorm.DB
}

func NewDBClient(dbConf *conf.Mysql, logLevel logger.LogLevel, customLogger *logs.Logger) (*DBClient, error) {
	//dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=True&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci", dbConf.User, dbConf.Password, dbConf.Host, dbConf.Port, dbConf.Name)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dbConf.Dsn(), // DSN data source name
		DefaultStringSize:         256,          // string 类型字段的默认长度
		DisableDatetimePrecision:  true,         // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,         // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,         // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,        // 根据版本自动配置
	}),
		&gorm.Config{QueryFields: true},
		Logger{
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Logger:                    customLogger,
		},
		ConnPool{
			ConnMaxIdleTime: 300 * time.Second,
			ConnMaxLifetime: 300 * time.Second,
			MaxIdleConns:    dbConf.MaxIdleConns,
			MaxOpenConns:    dbConf.MaxOpenConns,
		})
	if err != nil {
		return nil, err
	}
	return &DBClient{db: db}, nil
}

func (w *DBClient) GetDB(ctx context.Context, opType DbType) *gorm.DB {
	tx := w.GetTransaction(ctx)
	if tx != nil {
		return tx
	}

	db := w.db.WithContext(ctx)
	if opType == Write {
		db = db.Clauses(dbresolver.Write)
	}
	return db
}

func (w *DBClient) Begin(ctx context.Context, opType DbType) (context.Context, *gorm.DB) {
	tx := w.GetTransaction(ctx)
	if tx != nil {
		return ctx, tx
	}
	tx = w.GetDB(ctx, opType)
	newTx := tx.Begin()
	ctx = context.WithValue(ctx, TRANSACTION_KEY, newTx)
	ctx = context.WithValue(ctx, TRANSACTION_SIGNAL, make(chan struct{}, 1))
	ctx = context.WithValue(ctx, TRANSACTION_COMMITTED_MAP, &sync.Map{})
	return ctx, newTx
}

func (w *DBClient) Commit(ctx context.Context) error {
	tx := w.GetTransaction(ctx)
	if tx == nil {
		return fmt.Errorf("no transaction")
	}
	err := tx.Commit().Error
	if err == nil {
		m := ctx.Value(TRANSACTION_COMMITTED_MAP)
		if m != nil {
			m.(*sync.Map).Store(TRANSACTION_COMMITTED_MAP_KEY, true)
		}
		close(ctx.Value(TRANSACTION_SIGNAL).(chan struct{}))
	}
	return err
}

func (w *DBClient) Rollback(ctx context.Context) error {
	tx := w.GetTransaction(ctx)
	if tx == nil {
		return fmt.Errorf("no transaction")
	}
	// 不管Rollback是否有错，都返回false
	m := ctx.Value(TRANSACTION_COMMITTED_MAP)
	if m != nil {
		m.(*sync.Map).Store(TRANSACTION_COMMITTED_MAP_KEY, false)
	}
	close(ctx.Value(TRANSACTION_SIGNAL).(chan struct{}))
	return tx.Rollback().Error
}

func (w *DBClient) GetTransaction(ctx context.Context) *gorm.DB {
	value := ctx.Value(TRANSACTION_KEY)
	if value == nil {
		return nil
	}
	tx, ok := value.(*gorm.DB)
	if !ok || tx == nil {
		return nil
	}
	return tx
}

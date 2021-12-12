package el_mysql

import "context"

func Tx(ctx context.Context, dbClient *DBClient, f func(ctx context.Context) error) error {
	txCtx, db := dbClient.Begin(ctx, Write)
	if db.Error != nil {
		//log.V1.CtxError(ctx, "[nl_mysql] Tx, transaction begin error, err=%v", db.Error)
		return db.Error
	}
	var err error

	// 回滚事务
	defer func() {
		if err != nil {
			rbErr := dbClient.Rollback(txCtx)
			if rbErr != nil {
				//log.V1.CtxError(ctx, "[nl_mysql] Tx, rollback transaction error, err=%v", rbErr)
			}
		}
	}()

	// 执行事务
	err = f(txCtx)
	if err != nil {
		return err
	}

	// 提交事务
	cErr := dbClient.Commit(txCtx)
	if cErr != nil {
		//log.V1.CtxError(ctx, "[nl_mysql] Tx, commit transaction error, err=%v", cErr)
	}
	return cErr
}


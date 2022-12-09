package gopool

import (
	"context"
)

// defaultPool 不应该被修改或者 Closed，所以保护起来
var defaultPool Pool

func init() {
	defaultPool = NewPool(10000)
}

func Go(f func()) {
	CtxGo(context.Background(), f)
}

func CtxGo(ctx context.Context, f func()) {
	defaultPool.CtxGo(ctx, f)
}

// 不建议更改大小，容易造成全局其它调用者的问题
func SetCap(cap int32) {
	defaultPool.SetCap(cap)
}

// 设置默认 pool panic 情况下的 handler
func SetPanicHandler(f func(context.Context, interface{})) {
	defaultPool.SetPanicHandler(f)
}

// 获取默认 pool 中的 goroutine 数量
func WorkerCount() int32 {
	return defaultPool.WorkerCount()
}

// Logger is used for logging formatted messages.
type Logger interface {
	// Printf must have the same semantics as log.Printf.
	Printf(format string, args ...interface{})
}

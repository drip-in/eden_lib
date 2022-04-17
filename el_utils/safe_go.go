package el_utils

import (
	"context"
	"github.com/drip-in/eden_lib/logs"
	"runtime"
)

func GoSafeWithCtx(ctx context.Context, fn func(), cleanups ...func()) {
	go RunSafeFn(ctx, fn, cleanups...)
}

func GoSafe(fn func(), cleanups ...func()) {
	ctx := context.Background()
	go RunSafeFn(ctx, fn, cleanups...)
}

func RunSafeFn(ctx context.Context, fn func(), cleanups ...func()) {
	defer RecoverAndCleanup(ctx, cleanups...)
	fn()
}

func RecoverAndCleanup(ctx context.Context, cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		PrintErrStack(ctx, p)
	}
}

func PrintErrStack(ctx context.Context, err interface{}) {
	const size = 64 << 10
	buff := make([]byte, size)
	buff = buff[:runtime.Stack(buff, false)]
	logs.Error("panic info", logs.String("err", err.(error).Error()), logs.String("stack", string(buff)))
}

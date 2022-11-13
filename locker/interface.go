package locker

import (
	"context"
	"time"
)

var (
	LockerImpl ILocker
)

func InitLockerImpl(lock ILocker) {
	LockerImpl = lock
}

type ILocker interface {
	TryLock(ctx context.Context, key string) (unLockFunc func())
	TryLockWithDuration(ctx context.Context, key string, duration time.Duration) (unLockFunc func())
}

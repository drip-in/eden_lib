package el_tool

import (
	"context"
	"time"
)

var (
	LockerImpl  ILocker
	StorageImpl IStorage
)

func InitLockerImpl(lock ILocker) {
	LockerImpl = lock
}

func InitStorageImpl(s IStorage) {
	StorageImpl = s
}

type ILocker interface {
	TryLock(ctx context.Context, key string) (unLockFunc func())
	TryLockWithDuration(ctx context.Context, key string, duration time.Duration) (unLockFunc func())
	TryLockWithValAndDuration(ctx context.Context, key string, value string, duration time.Duration) bool
	UnLock(ctx context.Context, key string, value string) error
}

type IStorage interface {
	Set(ctx context.Context, key string, val interface{}, expiration time.Duration) error
	SetNX(ctx context.Context, key string, val interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Del(ctx context.Context, keys ...string) error
}

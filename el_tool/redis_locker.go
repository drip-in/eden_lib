package el_tool

import (
	"context"
	"fmt"
	"github.com/drip-in/eden_lib/logs"
	"github.com/go-redis/redis"
	"gorm.io/gorm/utils"
	"time"
)

const (
	EXPIRED_TIME = time.Second * 11
)

type Locker struct {
	namespace   string
	redisClient *redis.Client
}

func NewLocker(namespace string, redisClient *redis.Client) ILocker {
	if redisClient == nil {
		panic("invalid cache config for poi settle cache")
	}

	return &Locker{namespace, redisClient}
}

func (p *Locker) genCacheKey(key string) string {
	return fmt.Sprintf("%v_%v", p.namespace, key)
}

func (p *Locker) TryLock(ctx context.Context, key string) (unLockFunc func()) {
	return p.tryLock(ctx, key, EXPIRED_TIME)
}

func (p *Locker) TryLockWithDuration(ctx context.Context, key string, duration time.Duration) (unLockFunc func()) {
	return p.tryLock(ctx, key, duration)
}

func (p *Locker) TryLockWithValAndDuration(ctx context.Context, key string, value string, duration time.Duration) bool {
	if key == "" {
		panic("empty key")
	}

	cacheKey := p.genCacheKey(key)
	success, err := p.redisClient.WithContext(ctx).SetNX(cacheKey, value, duration).Result()
	if err != nil {
		logs.Error("redis client setnx", logs.String("err", err.Error()), logs.String("cacheKey", cacheKey))
		return false
	}
	return success
}

func (p *Locker) tryLock(ctx context.Context, key string, duration time.Duration) (unLockFunc func()) {
	if key == "" {
		panic("empty key")
	}

	cacheKey := p.genCacheKey(key)
	cacheValue := time.Now().UnixNano()

	success, err := p.redisClient.WithContext(ctx).SetNX(cacheKey, cacheValue, duration).Result()
	if err != nil {
		logs.Error("redis client setnx", logs.String("err", err.Error()), logs.String("cacheKey", cacheKey))
		return nil
	}
	if success {
		return func() {
			p.UnLock(ctx, key, utils.ToString(cacheValue))
		}
	} else {
		return nil
	}
}

func (p *Locker) UnLock(ctx context.Context, key string, value string) error {
	cacheKey := p.genCacheKey(key)
	currentValue, err := p.redisClient.WithContext(ctx).Get(cacheKey).Result()
	if err != nil {
		logs.Warn("redis client get", logs.String("err", err.Error()), logs.String("cacheKey", cacheKey))
		return err
	}
	if currentValue == value { // 防止本锁超时后, 解锁别人的锁
		_, err = p.redisClient.WithContext(ctx).Del(cacheKey).Result()
		if err != nil {
			logs.Error("redis client del", logs.String("err", err.Error()), logs.String("cacheKey", cacheKey))
			return err
		}
	} else {
		logs.CtxError(ctx, "redis unlock fail, value not equal")
	}
	return nil
}

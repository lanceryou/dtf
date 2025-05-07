package redis

import (
	"context"
	"github.com/lanceryou/dtf/utils/dlock"
	"time"

	"github.com/go-redis/redis"
	"log/slog"
)

const (
	lockPrefix = "/lock/"
)

// RedisLocker use redis achieve dist lock
type RedisLocker struct {
	*redis.Client
	Unique string
	Lease  *dlock.LockLeaser
}

// Lock dist lock
func (r *RedisLocker) Lock(ctx context.Context, key string, options ...dlock.LockOption) (err error) {
	var op dlock.Options
	for _, o := range options {
		o(&op)
	}

	lw := dlock.NewLockWatcher(op.Duration, op.Block)
	set, err := lw.Lock(ctx, func() (b bool, e error) {
		return r.SetNX(lockKey(key), r.Unique, op.Duration).Result()
	})
	if err != nil || !set {
		slog.Error("redis lock error", key, err)
		return err
	}

	slog.Info("redis lock success", key)
	if op.AutoLease {
		r.Lease.Lease(ctx, key, op.Duration)
	}
	return nil
}

func (r *RedisLocker) Extend(ctx context.Context, key string, duration time.Duration) error {
	return r.Expire(lockKey(key), duration).Err()
}

// UnLock unlock dist lock
func (r *RedisLocker) UnLock(ctx context.Context, key string) error {
	script := `
	if redis.call("get",KEYS[1]) == ARGV[1] then
		return redis.call("del",KEYS[1])
	else
		return 0
	end`
	err := r.Do("eval", script, 1, []string{lockKey(key)}, r.Unique).Err()
	if err == nil {
		r.Lease.Cancel(key)
	}
	return err
}

// String use redis
func (r *RedisLocker) String() string {
	return "redis"
}

// NewLocker new a dist locker use redis
func NewLocker(client *redis.Client, uniq string) dlock.DistLock {
	locker := &RedisLocker{
		Client: client,
		Unique: uniq,
	}
	locker.Lease = dlock.NewLockLeaser(locker.Extend)
	return locker
}

func lockKey(key string) string {
	return lockPrefix + key
}

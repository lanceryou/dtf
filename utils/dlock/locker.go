package dlock

import (
	"context"
	"sync"
	"time"

	"log/slog"
)

func NewLockWatcher(lockTime time.Duration, keepAlive bool) *LockWatcher {
	return &LockWatcher{
		lockTime:  lockTime,
		keepAlive: keepAlive,
	}
}

// LockWatcher 负责自动锁
type LockWatcher struct {
	lockTime  time.Duration
	keepAlive bool
	lease     func(ctx context.Context)
}

func (d *LockWatcher) Lock(ctx context.Context, fn func() (bool, error)) (bool, error) {
	lock, err := fn()
	if err != nil {
		return false, err
	}
	if lock {
		return lock, nil
	}

	// 没有锁住
	if !d.keepAlive {
		return false, ExistLockKeyErr
	}

	// 自动watch续期
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			time.Sleep(d.lockTime / 4)
			lock, err = fn()
			if !lock {
				continue
			}

			return lock, err
		}
	}
}

type LockLeaser struct {
	fn  func(ctx context.Context, key string, expire time.Duration) error
	mtx sync.Mutex
	lm  map[string]chan bool
}

func NewLockLeaser(fn func(ctx context.Context, key string, expire time.Duration) error) *LockLeaser {
	return &LockLeaser{
		fn: fn,
		lm: make(map[string]chan bool),
	}
}

// Lease 注意实现的续期是通过go，不建议在并发高的场景使用
func (l *LockLeaser) Lease(ctx context.Context, key string, expire time.Duration) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	l.lm[key] = make(chan bool, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				l.Cancel(key)
				return
			case <-l.lm[key]:
				return
			default:
				time.Sleep(expire / 5)
				err := l.fn(ctx, key, expire)
				if err != nil {
					slog.Error("lease lock error", err)
				}
			}
		}
	}()
}

func (l *LockLeaser) Cancel(key string) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	if ch, ok := l.lm[key]; ok {
		close(ch)
		delete(l.lm, key)
	}
}

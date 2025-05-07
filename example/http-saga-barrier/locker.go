package main

import (
	"context"
	"github.com/lanceryou/dtf/utils/dlock"
	"sync"
	"time"
)

var _ dlock.DistLock = &Locker{}

type Locker struct {
	mtx sync.Mutex
}

func (l *Locker) Lock(ctx context.Context, key string, option ...dlock.LockOption) error {
	l.mtx.Lock()
	return nil
}

func (l *Locker) Extend(ctx context.Context, key string, duration time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (l *Locker) UnLock(ctx context.Context, key string) error {
	l.mtx.Unlock()
	return nil
}

func (l *Locker) String() string {
	//TODO implement me
	panic("implement me")
}

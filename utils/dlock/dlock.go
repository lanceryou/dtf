package dlock

import (
	"context"
	"errors"
	"time"
)

var (
	// ExistLockKeyErr exist lock key err
	ExistLockKeyErr = errors.New("lock key exist")
)

// DistLock distributed lock abstract
type DistLock interface {
	Lock(ctx context.Context, key string, option ...LockOption) error
	Extend(ctx context.Context, key string, duration time.Duration) error
	UnLock(ctx context.Context, key string) error
	String() string
}

// Op DistLock option
type Options struct {
	Duration  time.Duration
	Block     bool
	AutoLease bool
}

func (o *Options) Apply() {
	if o.Duration == 0 {
		o.Duration = time.Second * 30
	}
}

type LockOption func(*Options)

// WithExpire set the expire duration of the lock
func WithExpire(duration time.Duration) LockOption {
	return func(opt *Options) {
		opt.Duration = duration
	}
}

// WithBlock will try to get the lock until succ or fatal error. NOT suggested!
func WithBlock() LockOption {
	return func(opt *Options) {
		opt.Block = true
	}
}

func WithLease() LockOption {
	return func(opt *Options) {
		opt.AutoLease = true
	}
}

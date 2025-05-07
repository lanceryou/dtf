package tc

import "time"

type RunTimer interface {
	NextRunTime(t time.Time) int64
}

type TimerFunc func(t time.Time) int64

func (f TimerFunc) NextRunTime(t time.Time) int64 {
	return f(t)
}

type DurationTimer struct {
	d time.Duration
}

func NewDurationTimer(d time.Duration) *DurationTimer {
	return &DurationTimer{d: d}
}

func (d *DurationTimer) NextRunTime(t time.Time) int64 {
	return t.Add(d.d).Unix()
}

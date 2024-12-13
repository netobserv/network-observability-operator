package operational

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Timer struct {
	startTime *time.Time
	observer  prometheus.Observer
}

func NewTimer(o prometheus.Observer) *Timer {
	return &Timer{
		observer: o,
	}
}

// Start starts or restarts the timer, regardless if a previous call occurred without being observed first
func (t *Timer) Start() time.Time {
	now := time.Now()
	t.startTime = &now
	return now
}

// StartOnce starts the timer just the first time. Subsequent calls will be ignored,
// until the timer is observed
func (t *Timer) StartOnce() time.Time {
	if t.startTime == nil {
		now := time.Now()
		t.startTime = &now
	}
	return *t.startTime
}

func (t *Timer) ObserveMilliseconds() {
	t.observe(func(d time.Duration) float64 { return float64(d.Milliseconds()) })
}

func (t *Timer) ObserveSeconds() {
	t.observe(time.Duration.Seconds)
}

func (t *Timer) observe(f func(d time.Duration) float64) {
	if t.startTime != nil {
		duration := time.Since(*t.startTime)
		t.observer.Observe(f(duration))
		t.startTime = nil
	}
}

package clock

import (
	"sync"
	"time"
)

type Fixed struct {
	mu  sync.RWMutex
	now time.Time
}

func NewFixed(now time.Time) *Fixed { return &Fixed{now: now.UTC()} }

func (f *Fixed) Now() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.now
}

func (f *Fixed) Set(now time.Time) {
	f.mu.Lock()
	f.now = now.UTC()
	f.mu.Unlock()
}

func (f *Fixed) Advance(d time.Duration) {
	f.mu.Lock()
	f.now = f.now.Add(d)
	f.mu.Unlock()
}

package watcher

import (
	"sync"
	"time"
)

// Debouncer delays function execution until after a period of inactivity
type Debouncer struct {
	duration time.Duration
	timer    *time.Timer
	mu       sync.Mutex
	onFire   func()
}

// NewDebouncer creates a new debouncer with the given duration and callback
func NewDebouncer(duration time.Duration, onFire func()) *Debouncer {
	return &Debouncer{
		duration: duration,
		onFire:   onFire,
	}
}

// Trigger resets the debounce timer
// If called multiple times within the duration, only the last call triggers the callback
func (d *Debouncer) Trigger() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.duration, d.onFire)
}

// Cancel stops any pending callback
func (d *Debouncer) Cancel() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}

// IsPending returns true if there's a pending callback
func (d *Debouncer) IsPending() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.timer != nil
}

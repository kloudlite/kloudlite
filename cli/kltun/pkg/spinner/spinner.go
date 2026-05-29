package spinner

import (
	"fmt"
	"sync"
	"time"
)

// Spinner displays an animated progress indicator
type Spinner struct {
	message  string
	frames   []string
	interval time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}
	mu       sync.Mutex
	running  bool
}

// New creates a new spinner with the given message
func New(message string) *Spinner {
	return &Spinner{
		message:  message,
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		interval: 80 * time.Millisecond,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.doneCh)
		i := 0
		for {
			select {
			case <-s.stopCh:
				return
			default:
				fmt.Printf("\r%s %s", s.frames[i%len(s.frames)], s.message)
				i++
				time.Sleep(s.interval)
			}
		}
	}()
}

// Stop stops the spinner and shows success
func (s *Spinner) Stop(success bool) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()

	<-s.doneCh

	if success {
		fmt.Printf("\r✓ %s\n", s.message)
	} else {
		fmt.Printf("\r✗ %s\n", s.message)
	}
}

// StopWithMessage stops the spinner with a custom message
func (s *Spinner) StopWithMessage(success bool, msg string) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()

	<-s.doneCh

	if success {
		fmt.Printf("\r✓ %s\n", msg)
	} else {
		fmt.Printf("\r✗ %s\n", msg)
	}
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

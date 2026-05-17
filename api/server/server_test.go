package server

import (
	"errors"
	"testing"
	"time"
)

func TestWaitForHTTPSStartupReturnsImmediateError(t *testing.T) {
	want := errors.New("bind failed")
	errCh := make(chan error, 1)
	errCh <- want

	if err := waitForHTTPSStartup(errCh, time.Second); !errors.Is(err, want) {
		t.Fatalf("expected immediate startup error %v, got %v", want, err)
	}
}

func TestWaitForHTTPSStartupContinuesAfterDelay(t *testing.T) {
	errCh := make(chan error, 1)

	if err := waitForHTTPSStartup(errCh, time.Nanosecond); err != nil {
		t.Fatalf("expected nil after startup delay, got %v", err)
	}
}

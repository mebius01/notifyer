package async_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"notifyer/internal/platform/async"
)

func TestGroupWaitsForAll(t *testing.T) {
	var g async.Group
	var done atomic.Int32
	for range 5 {
		g.Go(func() {
			time.Sleep(20 * time.Millisecond)
			done.Add(1)
		})
	}

	if err := g.Wait(context.Background()); err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if got := done.Load(); got != 5 {
		t.Errorf("finished = %d, want 5", got)
	}
}

func TestGroupWaitTimeout(t *testing.T) {
	var g async.Group
	release := make(chan struct{})
	g.Go(func() { <-release })

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := g.Wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want DeadlineExceeded", err)
	}
	close(release) // let the goroutine finish
}

func TestGroupWaitEmpty(t *testing.T) {
	var g async.Group
	if err := g.Wait(context.Background()); err != nil {
		t.Errorf("Wait on empty group: %v", err)
	}
}

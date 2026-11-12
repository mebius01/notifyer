package async

import (
	"context"
	"sync"
)

// Group runs goroutines and lets the owner wait for them.
type Group struct{ wg sync.WaitGroup }

func (g *Group) Go(fn func()) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		fn()
	}()
}

// Wait blocks until all goroutines finish or ctx expires.
func (g *Group) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() { g.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

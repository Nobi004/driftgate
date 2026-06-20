package runner

import (
	"context"
	"fmt"
	// "sync"

	"golang.org/x/sync/errgroup"
)

// WorkerPool runs test cases in parallel with controlled concurrency
type WorkerPool struct {
	concurrency int
}

func NewWorkerPool(concurrency int) *WorkerPool {
	return &WorkerPool{concurrency: concurrency}
}

// Run executes all test cases in parallel, stopping on first fatal error
func (wp *WorkerPool) Run(ctx context.Context, tasks []func() error) error {
	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, wp.concurrency) // semaphore

	for i, task := range tasks {
		i, task := i, task // CAPTURE LOOP VARIABLES — THIS IS CRITICAL
		g.Go(func() error {
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			select {
			case <-ctx.Done():
				return ctx.Err() // cancelled by parent
			default:
				if err := task(); err != nil {
					return fmt.Errorf("task %d: %w", i, err)
				}
				return nil
			}
		})
	}

	return g.Wait()
}

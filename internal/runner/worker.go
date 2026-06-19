package runner

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Worker Pool runs test cases  in parallel with controlled concurrency
type WorkerPool struct {
	concurrency int
}

func newWorkerPool(concurrency int) *WorkerPool {
	return &WorkerPool{concurrency: concurrency}

}

// Run executes all test cases in parallel, stopping on first fatal error
func (wp *WorkerPool) Run(ctx context.Context, tasks []func() error) error {
	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, wp.concurrency)

	for i, task := range tasks {
		i, task := i, task // capture loop variables this is critical
		g.Go(func() error {
			sem <- struct{}{}        //acquire
			defer func() { <-sem }() //release

			select {
			case <-ctx.Done():
				return ctx.Err() // cancel by parent
			default:
				if err := task(); err != nil {
					return fmt.Errorf("task %d : %w", i, err) // wrap error with task index

				}
				return nil

			}
		})
	}
	return g.Wait()
}

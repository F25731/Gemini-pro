package app

import (
	"context"
	"errors"
	"sync/atomic"
)

var ErrQueueFull = errors.New("server queue is full")

type WorkerPool struct {
	jobs      chan func()
	workers   int
	active    atomic.Int64
	completed atomic.Int64
	failed    atomic.Int64
}

type PoolStats struct {
	Workers   int   `json:"workers"`
	Queued    int   `json:"queued"`
	Active    int64 `json:"active"`
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
}

func NewWorkerPool(workers int, queue int) *WorkerPool {
	pool := &WorkerPool{jobs: make(chan func(), queue), workers: workers}
	for i := 0; i < workers; i++ {
		go func() {
			for job := range pool.jobs {
				pool.active.Add(1)
				job()
				pool.active.Add(-1)
			}
		}()
	}
	return pool
}

func (p *WorkerPool) Run(ctx context.Context, fn func(context.Context) error) error {
	result := make(chan error, 1)
	job := func() {
		err := fn(ctx)
		if err != nil {
			p.failed.Add(1)
		} else {
			p.completed.Add(1)
		}
		result <- err
	}
	select {
	case p.jobs <- job:
	default:
		return ErrQueueFull
	}
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *WorkerPool) Stats() PoolStats {
	return PoolStats{
		Workers:   p.workers,
		Queued:    len(p.jobs),
		Active:    p.active.Load(),
		Completed: p.completed.Load(),
		Failed:    p.failed.Load(),
	}
}

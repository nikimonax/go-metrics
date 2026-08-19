package work

import (
	"context"
	"sync"
)

type WorkerPool struct {
	lifecycle

	config PoolConfig
	queue  chan Task
	wg     sync.WaitGroup
}

func (pool *WorkerPool) onStart(ctx context.Context) error {
	pool.queue = make(chan Task, pool.config.QueueSize)
	worker := NewWorker(pool.config.WorkerConfig)

	for range pool.config.WorkerCount {
		pool.wg.Add(1)
		go func() {
			defer pool.wg.Done()
			worker.Run(ctx, pool.queue)
		}()
	}

	return nil
}

func (pool *WorkerPool) onStop() error {
	pool.wg.Wait()
	return nil
}

func (pool *WorkerPool) Submit(ctx context.Context, task Task) error {
	pool.mu.Lock()

	if pool.state != lifecycleRunning {
		pool.mu.Unlock()
		return newErrInvalidState("not running")
	}

	queue := pool.queue
	runCtx := pool.runCtx
	pool.mu.Unlock()

	// NOTE: здесь не гарантируется, что при конкурентном
	// 		 Stop и Submit задача будет выполнена
	select {
	case queue <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-runCtx.Done():
		return newErrInvalidState("not running")
	}
}

func NewPool(config PoolConfig) (*WorkerPool, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	pool := WorkerPool{config: config}
	pool.lifecycle.onStart = pool.onStart
	pool.lifecycle.onStop = pool.onStop
	pool.lifecycle.stopTimeout = config.StopTimeout //nolint:staticcheck

	return &pool, nil
}

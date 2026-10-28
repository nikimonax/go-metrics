package work

import (
	"context"
	"sync"
)

// Пул воркеров с очередью задач фиксированной ёмкости.
// Управление запуском и остановкой осуществляется
// через встроенный жизненный цикл (Start/Stop/Run).
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

// Submit implements [Submitter].
//
// Помещает задачу в очередь пула. Если задача помечена
// SkipIfFullChan, при заполненной очереди она молча пропускается.
// Иначе блокируется до записи в очередь, отмены ctx или остановки пула.
func (pool *WorkerPool) Submit(ctx context.Context, task Task) error {
	pool.mu.Lock()

	if pool.state != lifecycleRunning {
		pool.mu.Unlock()
		return newErrInvalidState("not running")
	}

	queue := pool.queue
	runCtx := pool.runCtx
	pool.mu.Unlock()

	if task.SkipIfFullChan {
		select {
		case queue <- task:
		default:
		}

		return nil
	}

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

// Создаёт пул воркеров с указанной конфигурацией.
// Возвращает ошибку, если конфигурация некорректна.
func NewPool(config PoolConfig) (*WorkerPool, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	pool := WorkerPool{config: config}
	pool.lifecycle.onStart = pool.onStart
	pool.lifecycle.onStop = pool.onStop
	pool.stopTimeout = config.StopTimeout

	return &pool, nil
}

package work

import "context"

// Определяет интерфейс постановки задач на выполнение.
type Submitter interface {
	// Помещает задачу в очередь выполнения. Может заблокироваться
	// до освобождения места в очереди, отмены ctx или остановки пула.
	Submit(context.Context, Task) error
}

// Адаптирует *WorkerPool к интерфейсу Submitter.
func PoolAsSubmitter(pool *WorkerPool) Submitter {
	return pool
}

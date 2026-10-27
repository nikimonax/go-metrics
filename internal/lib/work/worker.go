package work

import "context"

// Выполняет задачи из очереди, обрабатывая ошибки и паники
// через колбэки конфигурации.
type Worker struct {
	config WorkerConfig
}

// Запускает цикл обработки задач из queue до отмены ctx или
// закрытия канала очереди.
func (w *Worker) Run(ctx context.Context, queue <-chan Task) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-queue:
			if !ok {
				return
			}

			w.Handle(ctx, task)
		}
	}
}

// Выполняет одну задачу.
// При ошибке вызывает OnError, при панике - OnPanic (если они заданы).
func (w *Worker) Handle(ctx context.Context, task Task) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if w.config.OnPanic != nil {
				w.config.OnPanic(task.Name, recovered)
			}
		}
	}()

	err := task.Callback(ctx)

	if err != nil && w.config.OnError != nil {
		w.config.OnError(task.Name, err)
	}
}

// Создаёт воркера с указанной конфигурацией.
func NewWorker(config WorkerConfig) *Worker {
	return &Worker{config: config}
}

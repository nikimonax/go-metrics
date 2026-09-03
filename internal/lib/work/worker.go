package work

import "context"

type Worker struct {
	config WorkerConfig
}

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

func NewWorker(config WorkerConfig) *Worker {
	return &Worker{config: config}
}

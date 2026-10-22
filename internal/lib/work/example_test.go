package work_test

import (
	"context"
	"fmt"
	"time"

	"github.com/nikimonax/go-metrics/internal/lib/work"
)

func ExampleNewPool() {
	pool, err := work.NewPool(work.PoolConfig{
		WorkerCount: 4,
		QueueSize:   16,
	})

	if err != nil {
		fmt.Println("create pool:", err)
		return
	}

	task, err := work.NewTask("print", func(ctx context.Context) error {
		fmt.Println("task done")
		return nil
	})

	if err != nil {
		fmt.Println("create task:", err)
		return
	}

	ctx := context.Background()
	if err := pool.Start(ctx); err != nil {
		fmt.Println("start pool:", err)
		return
	}

	defer func() {
		if err := pool.Stop(context.Background()); err != nil {
			fmt.Println("stop pool:", err)
		}
	}()

	if err := pool.Submit(ctx, task); err != nil {
		fmt.Println("submit:", err)
		return
	}

	// Output:
	// task done
}

func ExampleNewScheduler() {
	pool, err := work.NewPool(work.PoolConfig{
		WorkerCount: 2,
		QueueSize:   8,
	})

	if err != nil {
		fmt.Println("create pool:", err)
		return
	}

	scheduler, err := work.NewScheduler(pool, work.SchedulerConfig{})

	if err != nil {
		fmt.Println("create scheduler:", err)
		return
	}

	task, err := work.NewTask("tick", func(ctx context.Context) error {
		fmt.Println("tick done")
		return nil
	})

	if err != nil {
		fmt.Println("create task:", err)
		return
	}

	if err := scheduler.Schedule(task, 10*time.Millisecond); err != nil {
		fmt.Println("schedule:", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := pool.Start(ctx); err != nil {
		fmt.Println("start pool:", err)
		return
	}
	defer func() {
		if err := pool.Stop(context.Background()); err != nil {
			fmt.Println("stop pool:", err)
		}
	}()

	if err := scheduler.Start(ctx); err != nil {
		fmt.Println("start scheduler:", err)
		return
	}

	defer func() {
		if err := scheduler.Stop(context.Background()); err != nil {
			fmt.Println("stop scheduler:", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)
}

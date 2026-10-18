package work

import "context"

type Submitter interface {
	Submit(context.Context, Task) error
}

func PoolAsSubmitter(pool *WorkerPool) Submitter {
	return pool
}

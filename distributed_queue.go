package gocq

type DistributedQueue[T, R any] interface {
	IBaseQueue
	// Time complexity: O(1)
	Add(data T, configs ...JobConfigFunc) bool
}

type distributedQueue[T, R any] struct {
	queue IDistributedQueue
}

func NewDistributedQueue[T, R any](queue IDistributedQueue) DistributedQueue[T, R] {
	return &distributedQueue[T, R]{
		queue: queue,
	}
}

func (q *distributedQueue[T, R]) PendingCount() int {
	return q.queue.Len()
}

func (q *distributedQueue[T, R]) Add(data T, c ...JobConfigFunc) bool {
	j := newVoidJob[T, R](data, withRequiredJobId(loadJobConfigs(newConfig(), c...)))

	jBytes, err := j.Json()

	if err != nil {
		return false
	}

	return q.queue.Enqueue(jBytes)
}

func (q *distributedQueue[T, R]) Purge() {
	q.queue.Purge()
}

func (q *distributedQueue[T, R]) Close() error {
	return q.queue.Close()
}

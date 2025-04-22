package gocq

// queue is the base implementation of the Queue interface
// It contains an externalQueue for worker management and an internalQueue for job storage
type queue[T, R any] struct {
	*externalQueue[T, R]
	internalQueue IQueue
}

type Queue[T, R any] interface {
	IExternalQueue[T, R]
	// Add adds a new Job to the queue and returns a EnqueuedJob to handle the job.
	// Time complexity: O(1)
	Add(data T, configs ...JobConfigFunc) EnqueuedJob[R]
	// AddAll adds multiple Jobs to the queue and returns a EnqueuedGroupJob to handle the job.
	// Time complexity: O(n) where n is the number of Jobs added
	AddAll(data []Item[T]) EnqueuedGroupJob[R]
}

// Item represents a data item to be processed by a worker
// It combines a unique identifier with a value of any generic type
type Item[T any] struct {
	ID    string
	Value T
}

// newQueue creates a new queue with the given worker and internal queue implementation
// It sets the worker's queue to the provided queue and creates a new external queue for job management
func newQueue[T, R any](worker *worker[T, R], q IQueue) *queue[T, R] {
	worker.setQueue(q)

	return &queue[T, R]{
		externalQueue: newExternalQueue(worker),
		internalQueue: q,
	}
}

func (q *queue[T, R]) Add(data T, configs ...JobConfigFunc) EnqueuedJob[R] {
	j := newJob[T, R](data, loadJobConfigs(q.configs, configs...))

	q.internalQueue.Enqueue(j)
	q.postEnqueue(j)

	return j
}

func (q *queue[T, R]) AddAll(items []Item[T]) EnqueuedGroupJob[R] {
	l := len(items)
	groupJob := newGroupJob[T, R](uint32(l))

	for _, item := range items {
		j := groupJob.NewJob(item.Value, loadJobConfigs(q.configs, WithJobId(item.ID)))
		q.internalQueue.Enqueue(j)
		q.postEnqueue(j)
	}

	return groupJob
}

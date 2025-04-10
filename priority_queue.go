package gocq

type priorityQueue[T, R any] struct {
	*externalQueue[T, R]
	internalQueue IPriorityQueue
}

type PQItem[T any] struct {
	// ID is an optional identifier for the item
	ID string
	// Value contains the actual data stored in the queue
	Value T
	// Priority determines the item's position in the queue
	Priority int
}

type PriorityQueue[T, R any] interface {
	IExternalQueue[T, R]
	// Add adds a new Job with the given priority to the queue and returns a channel to receive the result.
	// Time complexity: O(log n)
	Add(data T, priority int, configs ...JobConfigFunc) EnqueuedJob[R]
	// AddAll adds multiple Jobs with the given priority to the queue and returns a channel to receive all responses.
	// Time complexity: O(n log n) where n is the number of Jobs added
	AddAll(items []PQItem[T]) EnqueuedGroupJob[R]
}

// NewPriorityQueue creates a new priorityQueue with the specified concurrency and worker function.
func newPriorityQueue[T, R any](worker *worker[T, R], pq IPriorityQueue) *priorityQueue[T, R] {
	worker.setQueue(pq)

	return &priorityQueue[T, R]{
		externalQueue: newExternalQueue(worker),
		internalQueue: pq,
	}
}

func (q *priorityQueue[T, R]) Add(data T, priority int, configs ...JobConfigFunc) EnqueuedJob[R] {
	j := newJob[T, R](data, loadJobConfigs(q.configs, configs...))

	q.internalQueue.Enqueue(j, priority)
	q.postEnqueue(j)

	return j
}

func (q *priorityQueue[T, R]) AddAll(items []PQItem[T]) EnqueuedGroupJob[R] {
	l := len(items)
	groupJob := newGroupJob[T, R](uint32(l))

	for _, item := range items {
		j := groupJob.NewJob(item.Value, loadJobConfigs(q.configs, WithJobId(item.ID)))

		q.internalQueue.Enqueue(j, item.Priority)
		q.postEnqueue(j)
	}

	return groupJob
}

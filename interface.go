package varmq

// IBaseQueue is the root interface of queue operations. workers queue needs to implement this interface.
type IBaseQueue interface {
	Len() int
	Dequeue() (any, bool)
	Values() []any
	Purge()
	Close() error
}

// IQueue is the root interface of queue operations.
type IQueue interface {
	IBaseQueue
	Enqueue(item any) bool
}

// IPriorityQueue is the root interface of priority queue operations.
type IPriorityQueue interface {
	IBaseQueue
	Enqueue(item any, priority int) bool
}

// IAcknowledgeable is the root interface of acknowledgeable operations.
type IAcknowledgeable interface {
	// Returns true if the item was successfully acknowledged, false otherwise.
	Acknowledge(ackID string) bool
	// DequeueWithAckId dequeues an item from the queue
	// Returns the item, a boolean indicating if the operation was successful, and the acknowledgment ID.
	DequeueWithAckId() (any, bool, string)
}

type IPersistentQueue interface {
	IQueue
	IAcknowledgeable
}

// IPersistentPriorityQueue is the root interface of persistent priority queue operations.
type IPersistentPriorityQueue interface {
	IPriorityQueue
	IAcknowledgeable
}

// ISubscribable is the root interface of subscribable operations.
type ISubscribable interface {
	Subscribe(func(action string))
}

// IDistributedQueue is the root interface of distributed queue operations.
type IDistributedQueue interface {
	IPersistentQueue
	ISubscribable
}

// IDistributedPriorityQueue is the root interface of distributed priority queue operations.
type IDistributedPriorityQueue interface {
	IPersistentPriorityQueue
	ISubscribable
}

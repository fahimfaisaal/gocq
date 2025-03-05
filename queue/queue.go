package queue

import (
	"container/list"
)

// Queue holds a linked list-based queue of generic items.
type Queue[T any] struct {
	internal *list.List
}

// newQueue creates an empty Queue using container/list.
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{internal: new(list.List)}
}

func (q *Queue[T]) Init() {
	q.internal.Init()
}

func (q *Queue[T]) Values() []T {
	values := make([]T, 0)

	for e := q.internal.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value.(T))
	}

	return values
}

func (q *Queue[T]) Len() int {
	return q.internal.Len()
}

// Enqueue adds an item at the back of the list in O(1).
func (q *Queue[T]) Enqueue(item Item[T]) {
	q.internal.PushBack(item.Value)
}

// Dequeue removes and returns the front item in O(1).
func (q *Queue[T]) Dequeue() (T, bool) {
	front := q.internal.Front()
	if front == nil {
		// Return zero value + false if empty
		var zeroValue T
		return zeroValue, false
	}

	// Retrieve the item’s value and remove the element in O(1).
	val := front.Value.(T)
	q.internal.Remove(front)
	return val, true
}

package varmq

import (
	"errors"
	"sync/atomic"
)

// ResultController manages result channels and provides safe operations for receiving
// both successful results and errors from asynchronous operations.
type ResultController[R any] struct {
	ch       chan Result[R] // Channel for sending/receiving results
	consumed atomic.Value   // Tracks if the channel has been consumed
	Output   Result[R]      // Stores the last result/error
}

// newResultController creates a new ResultController with a channel of the specified buffer size.
func newResultController[R any](bufferSize int) *ResultController[R] {
	return &ResultController[R]{
		ch: make(chan Result[R], bufferSize),
	}
}

// Read returns the underlying channel for reading.
// The channel can only be consumed once.
func (rc *ResultController[R]) Read() (<-chan Result[R], error) {
	if rc.consumed.Load() != nil {
		return nil, errors.New("result channel has already been consumed")
	}

	rc.consumed.Store(true)
	return rc.ch, nil
}

// Send sends a result to the channel.
func (rc *ResultController[R]) Send(result Result[R]) {
	rc.ch <- result
}

// Close closes the ResultController's channel.
func (rc *ResultController[R]) Close() error {
	close(rc.ch)

	return nil
}

package gocq

import (
	"fmt"
	"strings"
	"time"
)

type externalQueue[T, R any] struct {
	*worker[T, R]
}

type IExternalBaseQueue interface {
	// PendingCount returns the number of Jobs pending in the queue.
	PendingCount() int
	// Purge removes all pending Jobs from the queue.
	// Time complexity: O(n) where n is the number of pending Jobs
	Purge()
	// Close closes the queue and resets all internal states.
	// Time complexity: O(n) where n is the number of channels
	Close() error
}

// IExternalQueue is the root interface of concurrent queue operations.
type IExternalQueue[T, R any] interface {
	IExternalBaseQueue

	// Worker returns the worker.
	Worker() Worker[T, R]
	// JobById returns the job with the given id.
	JobById(id string) (EnqueuedJob[R], error)
	// GroupsJobById returns the groups job with the given id.
	GroupsJobById(id string) (EnqueuedSingleGroupJob[R], error)
	// WaitUntilFinished waits until all pending Jobs in the queue are processed.
	// Time complexity: O(n) where n is the number of pending Jobs
	WaitUntilFinished()
	// WaitAndClose waits until all pending Jobs in the queue are processed and then closes the queue.
	// Time complexity: O(n) where n is the number of pending Jobs
	WaitAndClose() error
}

func newExternalQueue[T, R any](worker *worker[T, R]) *externalQueue[T, R] {
	return &externalQueue[T, R]{
		worker: worker,
	}
}

func (wbq *externalQueue[T, R]) postEnqueue(j iJob[T, R]) {
	defer wbq.notifyToPullNextJobs()
	j.ChangeStatus(queued)

	if id := j.ID(); id != "" {
		wbq.Cache.Store(id, j)
	}
}

func (wbq *externalQueue[T, R]) PendingCount() int {
	return wbq.Queue.Len()
}

func (wbq *externalQueue[T, R]) Worker() Worker[T, R] {
	return wbq.worker
}

func (wbq *externalQueue[T, R]) JobById(id string) (EnqueuedJob[R], error) {
	val, ok := wbq.Cache.Load(id)
	if !ok {
		return nil, fmt.Errorf("job not found for id: %s", id)
	}

	return val.(EnqueuedJob[R]), nil
}

func (wbq *externalQueue[T, R]) GroupsJobById(id string) (EnqueuedSingleGroupJob[R], error) {
	if !strings.HasPrefix(id, groupIdPrefixed) {
		id = generateGroupId(id)
	}

	val, ok := wbq.Cache.Load(id)

	if !ok {
		return nil, fmt.Errorf("groups job not found for id: %s", id)
	}

	return val.(EnqueuedSingleGroupJob[R]), nil
}

func (wbq *externalQueue[T, R]) WaitUntilFinished() {
	// to ignore deadlock error if the queue is paused
	if wbq.IsPaused() {
		wbq.Resume()
	}

	wbq.sync.wg.Wait()

	// wait until all ongoing processes are done if still pending
	for wbq.PendingCount() > 0 || wbq.CurrentProcessingCount() > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}

func (wbq *externalQueue[T, R]) Purge() {
	prevValues := wbq.Queue.Values()
	wbq.Queue.Purge()

	// close all pending channels to avoid routine leaks
	for _, val := range prevValues {
		if j, ok := val.(iJob[T, R]); ok {
			j.CloseResultChannel()
		}
	}
}

func (q *externalQueue[T, R]) Close() error {
	q.Purge()
	q.Stop()
	q.WaitUntilFinished()

	return nil
}

func (q *externalQueue[T, R]) WaitAndClose() error {
	q.WaitUntilFinished()
	return q.Close()
}

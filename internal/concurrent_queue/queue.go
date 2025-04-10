package concurrent_queue

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/fahimfaisaal/gocq/v2/internal/job"
	"github.com/fahimfaisaal/gocq/v2/internal/queue"
	"github.com/fahimfaisaal/gocq/v2/types"
)

type ConcurrentQueue[T, R any] struct {
	Concurrency uint32
	Worker      any
	// channels for each concurrency level and store them in a stack.
	ChannelsStack []chan *job.Job[T, R]
	curProcessing uint32
	JobQueue      queue.IQueue[*job.Job[T, R]]
	wg            sync.WaitGroup
	mx            sync.Mutex
	isPaused      atomic.Bool
}

// Creates a new ConcurrentQueue with the specified concurrency and worker function.
// Internally it calls Init() to start the worker goroutines based on the concurrency.
func NewQueue[T, R any](concurrency uint32, worker types.Worker[T, R]) *ConcurrentQueue[T, R] {
	concurrentQueue := &ConcurrentQueue[T, R]{
		Concurrency:   concurrency,
		Worker:        worker,
		ChannelsStack: make([]chan *job.Job[T, R], concurrency),
		JobQueue:      queue.NewQueue[*job.Job[T, R]](),
	}

	concurrentQueue.Restart()
	return concurrentQueue
}

func (q *ConcurrentQueue[T, R]) Restart() {
	// first pause the queue to avoid routine leaks or deadlocks
	q.Pause()
	// wait until all ongoing processes are done to gracefully close the channels if any.
	q.WaitUntilFinished()

	// restart the queue with new channels and start the worker goroutines
	for i := range q.ChannelsStack {
		// close old channels to avoid routine leaks
		if q.ChannelsStack[i] != nil {
			close(q.ChannelsStack[i])
		}

		// This channel stack is used to pick the next available channel for processing a Job inside a worker goroutine.
		q.ChannelsStack[i] = make(chan *job.Job[T, R])
		go q.spawnWorker(q.ChannelsStack[i])
	}

	// resume the queue to process pending Jobs
	q.Resume()
}

// spawnWorker starts a worker goroutine to process jobs from the channel.
func (q *ConcurrentQueue[T, R]) spawnWorker(channel chan *job.Job[T, R]) {
	for j := range channel {
		switch worker := q.Worker.(type) {
		case types.VoidWorker[T]:
			err := worker(j.Data)
			j.ResultChannel.Err <- err
		case types.Worker[T, R]:
			result, err := worker(j.Data)
			if err != nil {
				j.ResultChannel.Err <- err
			} else {
				j.ResultChannel.Data <- result
			}
		default:
			// Log or handle the invalid type to avoid silent failures
			j.ResultChannel.Err <- errors.New("unsupported worker type passed to queue")
		}

		j.ChangeStatus(job.Finished)
		j.Close()

		q.mx.Lock()
		// push the channel back to the stack, so it can be used for the next Job
		q.ChannelsStack = append(q.ChannelsStack, channel)
		q.curProcessing--
		q.wg.Done()

		if q.shouldProcessNextJob("worker") {
			q.processNextJob()
		}
		q.mx.Unlock()
	}
}

// pickNextChannel picks the next available channel for processing a Job.
// Time complexity: O(1)
func (q *ConcurrentQueue[T, R]) pickNextChannel() chan<- *job.Job[T, R] {
	q.mx.Lock()
	defer q.mx.Unlock()
	l := len(q.ChannelsStack)

	// pop the last free channel
	channel := q.ChannelsStack[l-1]
	q.ChannelsStack = q.ChannelsStack[:l-1]
	return channel
}

// shouldProcessNextJob determines if the next job should be processed based on the current state.
func (q *ConcurrentQueue[T, R]) shouldProcessNextJob(action string) bool {
	switch action {
	case "add":
		return !q.isPaused.Load() && q.curProcessing < q.Concurrency
	case "resume":
		return q.curProcessing < q.Concurrency && q.JobQueue.Len() > 0
	case "worker":
		return !q.isPaused.Load() && q.JobQueue.Len() != 0
	default:
		return false
	}
}

// processNextJob processes the next Job in the queue.
func (q *ConcurrentQueue[T, R]) processNextJob() {
	j, has := q.JobQueue.Dequeue()

	if !has {
		return
	}

	if j.IsClosed() {
		q.wg.Done()
		// process next Job recursively if the current one is closed
		q.processNextJob()
		return
	}

	q.curProcessing++
	j.ChangeStatus(job.Processing)

	go func(job *job.Job[T, R]) {
		q.pickNextChannel() <- job
	}(j)
}

// PauseQueue pauses the processing of jobs.
func (q *ConcurrentQueue[T, R]) PauseQueue() {
	q.isPaused.Store(true)
}

func (q *ConcurrentQueue[T, R]) PendingCount() int {
	return q.JobQueue.Len()
}

func (q *ConcurrentQueue[T, R]) IsPaused() bool {
	return q.isPaused.Load()
}

func (q *ConcurrentQueue[T, R]) CurrentProcessingCount() uint32 {
	return q.curProcessing
}

func (q *ConcurrentQueue[T, R]) Pause() types.IConcurrentQueue[T, R] {
	q.PauseQueue()
	return q
}

func (q *ConcurrentQueue[T, R]) AddJob(enqItem queue.EnqItem[*job.Job[T, R]]) {
	q.wg.Add(1)
	q.mx.Lock()
	defer q.mx.Unlock()
	q.JobQueue.Enqueue(enqItem)
	enqItem.Value.ChangeStatus(job.Queued)

	// process next Job only when the current processing Job count is less than the concurrency
	if q.shouldProcessNextJob("add") {
		q.processNextJob()
	}
}

func (q *ConcurrentQueue[T, R]) Resume() {
	q.isPaused.Store(false)

	// Process pending jobs if any
	q.mx.Lock()
	defer q.mx.Unlock()

	// Process jobs up to concurrency limit
	for q.shouldProcessNextJob("resume") {
		q.processNextJob()
	}
}

func (q *ConcurrentQueue[T, R]) Add(data T) types.EnqueuedJob[R] {
	j := job.New[T, R](data)

	q.AddJob(queue.EnqItem[*job.Job[T, R]]{Value: j})
	return j
}

func (q *ConcurrentQueue[T, R]) AddAll(data []T) types.EnqueuedGroupJob[R] {
	groupJob := job.NewGroupJob[T, R](q.Concurrency).FanInResult(len(data))

	for _, item := range data {
		q.AddJob(queue.EnqItem[*job.Job[T, R]]{Value: groupJob.NewJob(item).Lock()})
	}

	return groupJob
}

func (q *ConcurrentQueue[T, R]) WaitUntilFinished() {
	q.wg.Wait()
}

func (q *ConcurrentQueue[T, R]) Purge() {
	q.mx.Lock()
	defer q.mx.Unlock()

	prevValues := q.JobQueue.Values()
	q.JobQueue.Init()
	q.wg.Add(-len(prevValues))

	// close all pending channels to avoid routine leaks
	for _, job := range prevValues {
		if job.ResultChannel.Data == nil {
			continue
		}

		close(job.ResultChannel.Data)
	}
}

func (q *ConcurrentQueue[T, R]) Close() error {
	q.Purge()

	// wait until all ongoing processes are done to gracefully close the channels
	q.wg.Wait()

	for _, channel := range q.ChannelsStack {
		if channel == nil {
			continue
		}

		close(channel)
	}

	q.ChannelsStack = make([]chan *job.Job[T, R], q.Concurrency)
	return nil
}

func (q *ConcurrentQueue[T, R]) WaitAndClose() error {
	q.wg.Wait()
	return q.Close()
}

# 🚀 GoCQ: High-Performance Concurrent Queue for Gophers

Package gocq offers a concurrent queue system using channels and goroutines, supporting both FIFO and priority operations.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)

GoCQ is a high-performance concurrent queue for Go, optimized for efficient task processing. It supports both FIFO and priority queues, featuring non-blocking job submission, dedicated worker channels, and a pre-allocated worker pool to ensure smooth and controlled concurrency. With optimized memory management, GoCQ minimizes allocations and prevents goroutine leaks, making it a reliable choice for high-throughput applications

## 🌟 Features

- Generic type support for both data and results
- Configurable concurrency limits
- FIFO queue with O(1) operations
- Priority queue support with O(log n) operations
- Pause/Resume functionality
- Clean and graceful shutdown mechanisms
- Thread-safe operations
- Non-blocking job submission

## 📋 Table of Contents

- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [API Reference](#-api-reference)
  - [Standard Queue](#standard-queue-fifo)
  - [Priority Queue](#priority-queue)
- [Examples](#-examples)
- [Performance](#-performance)

## 🔧 Installation

```bash
go get github.com/fahimfaisaal/gocq
```

## 🚀 Quick Start

```go
package main

import (
  "fmt"
  "time"

  "github.com/fahimfaisaal/gocq"
)

func main() {
  // Create a queue with 2 concurrent workers
  queue := gocq.NewQueue(2, func(data int) int {
    time.Sleep(500 * time.Millisecond)
    return data * 2
  })
  defer queue.Close()

  // Add a single job
  result := <-queue.Add(5)
  fmt.Println(result) // Output: 10

  // Add multiple jobs
  results := queue.AddAll(1, 2, 3, 4, 5)
  for result := range results {
    fmt.Println(result) // Output: 2, 4, 6, 8, 10 (unordered)
  }
}
```

## 📚 API Reference

### Standard Queue (FIFO)

#### State Management Methods

#### `PendingCount() int`

Returns the number of jobs waiting to be processed.

- Time Complexity: O(1)
- Returns: Number of pending jobs

#### `CurrentProcessingCount() uint`

Returns the number of jobs currently being processed.

- Time Complexity: O(1)
- Returns: Number of active jobs

#### `IsPaused() bool`

Checks if the queue is currently paused.

- Time Complexity: O(1)
- Returns: true if paused, false otherwise

#### Queue Operation Methods

#### `NewQueue[T, R any](concurrency uint, worker func(T) R) *ConcurrentQueue[T, R]`

Creates a new concurrent FIFO queue.

- Time Complexity: O(c) where c is concurrency and spawns c goroutines
- Parameters:
  - `concurrency`: Maximum number of concurrent workers
  - `worker`: Function to process each job
- Returns: A new concurrent queue instance

#### `Add(data T) <-chan R`

Adds a single job to the queue.

- Time Complexity: O(1)
- Returns: Merged channel to receive all results

#### `AddAll(data ...T) <-chan R`

Adds multiple jobs to the queue.

- Time Complexity: O(n) where n is number of jobs
- Returns: Channel to receive all results in order

#### `Pause() *ConcurrentQueue[T, R]`

Pauses job processing.

- Time Complexity: O(1)
- Returns: Queue instance for chaining

#### `Resume()`

Resumes job processing.

- Time Complexity: O(c) where c is the concurrency

#### Cleanup and Wait Methods

#### `Purge()`

Removes all pending jobs from the queue.

- Time Complexity: O(n) where n is number of pending jobs
- Note: Closes response channels for all purged jobs
- Effect: All pending jobs are removed, but currently processing jobs continue

#### `WaitUntilFinished()`

Blocks until all pending jobs complete.

#### `Close()`

Closes the queue and cleans up resources.

> Note: Waits for current processing jobs to finish

#### `WaitAndClose()`

Waits for completion of each pending job and closes the queue. combines `WaitUntilFinished()` and `Close()`

### Priority Queue

**The priority queue extends the standard queue with priority support.**

#### `NewPriorityQueue[T, R any](concurrency uint, worker func(T) R) *ConcurrentPriorityQueue[T, R]`

Creates a new concurrent priority queue.

- Time Complexity: O(1)
- Parameters:
  - `concurrency`: Maximum number of concurrent workers
  - `worker`: Function to process each job
- Returns: A new priority queue instance

#### `Add(data T, priority int) <-chan R`

Adds a job with priority (lower number = higher priority).

- Time Complexity: O(log n) where n is queue size
- Parameters:
  - `priority`: Lower value means higher priority
- Returns: Channel to receive the result

#### `AddAll(items []PQItem[T]) <-chan R`

Adds multiple prioritized jobs.

- Time Complexity: O(n log n) where n is number of items
- Returns: Merged channel to receive all results in priority order

## 💡 Examples

### Priority Queue Example

```go
queue := gocq.NewPriorityQueue(1, func(data int) int {
    return data * 2
})
defer queue.WaitAndClose()

// Add jobs with different priorities
items := []gocq.PQItem[int]{
    {Value: 1, Priority: 2}, // Lowest priority
    {Value: 2, Priority: 1}, // Medium priority
    {Value: 3, Priority: 0}, // Highest priority
}

results := queue.AddAll(items)
for result := range results {
    fmt.Println(result) // Output: 6, 4, 2 (processed by priority)
}
```

### Pause/Resume Example

```go
queue := gocq.NewQueue(2, func(data int) int {
    return data * 2
}).Pause() // paused

// Add jobs while paused (non-blocking)
resp1 := queue.Add(1)
resp2 := queue.Add(2)

// Resume processing
queue.Resume()

fmt.Println(<-resp1, <-resp2) // Output: 2 4
```

## 🚀 Performance

The implementation uses efficient data structures:

- Standard Queue: Based on `container/list` with O(1) operations
- Priority Queue: Based on `container/heap` implementation with O(log n) operations
- Non-blocking job submission
- Efficient worker pool management using channels and goroutines

## 👤 Author (Fahim Faisaal)

- GitHub: [@fahimfaisaal](https://github.com/fahimfaisaal)
- LinkedIn: [in/fahimfaisaal](https://www.linkedin.com/in/fahimfaisaal/)
- Twitter: [@FahimFaisaal](https://twitter.com/FahimFaisaal)

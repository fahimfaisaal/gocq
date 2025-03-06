package gocq

import (
	"testing"
	"time"
)

func TestConcurrentPriorityQueue(t *testing.T) {
	t.Run("Add with Priority", func(t *testing.T) {
		worker := func(data int) int {
			time.Sleep(100 * time.Millisecond)
			return data * 2
		}

		q := NewPriorityQueue(1, worker).Pause()

		resp1 := q.Add(1, 2)
		resp2 := q.Add(2, 1)
		resp3 := q.Add(3, 0)

		if count := q.PendingCount(); count != 3 {
			t.Errorf("Expected pending count to be 3, got %d", count)
		}

		q.Resume()
		if result := <-resp3; result != 6 {
			t.Errorf("Expected result to be 6, got %d", result)
		}
		if result := <-resp2; result != 4 {
			t.Errorf("Expected result to be 4, got %d", result)
		}
		if result := <-resp1; result != 2 {
			t.Errorf("Expected result to be 2, got %d", result)
		}

		q.WaitUntilFinished()
		if count := q.PendingCount(); count != 0 {
			t.Errorf("Expected pending count to be 0, got %d", count)
		}
		if count := q.CurrentProcessingCount(); count != 0 {
			t.Errorf("Expected current processing count to be 0, got %d", count)
		}
	})

	t.Run("AddAll with Priority", func(t *testing.T) {
		worker := func(data int) int {
			time.Sleep(100 * time.Millisecond)
			return data * 2
		}

		q := NewPriorityQueue(2, worker)

		data := []int{1, 2, 3, 4, 5}
		resultChan := q.AddAll(1, data...)

		results := make([]int, 0)
		for result := range resultChan {
			results = append(results, result)
		}

		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}

		expectedSum := 30 // sum of [2,4,6,8,10]
		actualSum := 0
		for _, r := range results {
			actualSum += r
		}

		if actualSum != expectedSum {
			t.Errorf("Expected sum of results to be %d, got %d", expectedSum, actualSum)
		}
	})
}

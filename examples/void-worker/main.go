package main

import (
	"fmt"
	"time"

	"github.com/goptics/varmq"
)

func main() {
	w := varmq.NewVoidWorker(func(data int) {
		// fmt.Printf("Processing: %d\n", data)
		time.Sleep(1 * time.Second)
	}, 100)

	q := w.BindQueue()
	start := time.Now()
	defer func() {
		fmt.Println("Time taken:", time.Since(start))
	}()
	defer q.WaitUntilFinished()
	defer fmt.Println("Added jobs")

	go func() {
		q.WaitUntilFinished()
		fmt.Println("Finished")
	}()

	for i := range 1000 {
		q.Add(i)
	}
}

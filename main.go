package main

import (
	"fmt"
	"time"
	"workerPool/worker"
)

func main() {

	pool := worker.NewWorkerPool(5)

	go pool.ProcessErrors()

	fmt.Println("Adding workers...")
	for i := 0; i < 3; i++ {
		pool.AddWorker()
	}

	fmt.Println()
	fmt.Println("Adding tasks...")
	for i := 1; i <= 10; i++ {
		task := fmt.Sprintf("Task-%d", i)
		pool.AddTask(task)
		fmt.Printf("Added %s\n", task)
		time.Sleep(200 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)

	fmt.Println()
	fmt.Println("Worker statuses:")
	pool.GetWorkerStatuses()

	fmt.Println()
	fmt.Println("Removing a worker...")
	if pool.WorkerCount() > 0 {
		for id := range pool.Workers {
			pool.RemoveWorkerByID(id)
			break
		}
	}

	fmt.Println()
	fmt.Println("Worker statuses after removal:")
	pool.GetWorkerStatuses()

	fmt.Println()
	fmt.Println("Removing all remaining workers...")
	pool.RemoveAllWorkers()

	fmt.Println()
	fmt.Println("Closing worker pool...")
	pool.Close()

	time.Sleep(2 * time.Second)

	fmt.Println()
	fmt.Println("Finished processing.")
}

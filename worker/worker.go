package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type WorkerStatus string

const (
	StatusIdle       WorkerStatus = "idle"
	StatusProcessing WorkerStatus = "processing"
	StatusError      WorkerStatus = "error"
	StatusExiting    WorkerStatus = "exiting"
)

type Worker struct {
	id     string
	status WorkerStatus
	cancel context.CancelFunc
}

type WorkerPool struct {
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	taskQueue chan string
	errorChan chan error
	Workers   map[string]*Worker
}

func NewWorkerPool(queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		ctx:       ctx,
		cancel:    cancel,
		taskQueue: make(chan string, queueSize),
		Workers:   make(map[string]*Worker),
		errorChan: make(chan error),
	}
}

func (wp *WorkerPool) AddWorker() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	id := uuid.New().String()
	workerCtx, cancel := context.WithCancel(wp.ctx)

	worker := &Worker{
		id:     id,
		status: StatusIdle,
		cancel: cancel,
	}
	wp.Workers[id] = worker

	go wp.Worker(workerCtx, worker)
	fmt.Printf("Worker %v added\n", id)
}

func (wp *WorkerPool) Worker(ctx context.Context, worker *Worker) {
	for {
		select {
		case <-ctx.Done():
			worker.status = StatusExiting
			fmt.Printf("Worker %v exiting\n", worker.id)
			return
		case task, ok := <-wp.taskQueue:
			if !ok {
				worker.status = StatusExiting
				fmt.Printf("Worker %v: task queue closed\n", worker.id)
				return
			}

			worker.status = StatusProcessing
			fmt.Printf("Worker %v processing task: %s\n", worker.id, task)

			// JUST FOR SIMULATION OF WORK
			time.Sleep(time.Second)

			worker.status = StatusIdle
			fmt.Printf("Worker %v completed task: %s\n", worker.id, task)
		}
	}
}

func (wp *WorkerPool) RemoveWorkerByID(id string) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if worker, exists := wp.Workers[id]; exists {
		worker.cancel()
		delete(wp.Workers, id)
		fmt.Printf("Worker %v removed\n", id)
	} else {
		fmt.Printf("Worker %v doesn't exist\n", id)
	}
}

func (wp *WorkerPool) RemoveAllWorkers() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	for id, worker := range wp.Workers {
		worker.cancel()
		delete(wp.Workers, id)
		fmt.Printf("Worker %v removed\n", id)
	}
}

func (wp *WorkerPool) AddTask(task string) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	select {
	case wp.taskQueue <- task:
	default:
		wp.errorChan <- fmt.Errorf("task queue is full, unable to add task: %s", task)
	}
}

func (wp *WorkerPool) Close() {
	wp.cancel()

	close(wp.taskQueue)
	close(wp.errorChan)
}

func (wp *WorkerPool) ProcessErrors() {
	for err := range wp.errorChan {
		fmt.Printf("Error occurred: %v\n", err)
	}
}

func (wp *WorkerPool) WorkerCount() int {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	return len(wp.Workers)
}

func (wp *WorkerPool) GetWorkerStatuses() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	for id, worker := range wp.Workers {
		fmt.Printf("Worker %v status: %s\n", id, worker.status)
	}
}

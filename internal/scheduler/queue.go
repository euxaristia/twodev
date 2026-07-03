package scheduler

import (
	"context"
	"sync"
)

// JobRequest describes work to schedule.
type JobRequest struct {
	ProjectPath string
	JobName     string
	BuildNumber int
}

// Queue is an in-memory job queue.
type Queue struct {
	mu    sync.Mutex
	items []JobRequest
	subs  []chan JobRequest
}

// NewQueue creates a job queue.
func NewQueue() *Queue {
	return &Queue{}
}

// Enqueue adds a job request.
func (q *Queue) Enqueue(req JobRequest) {
	q.mu.Lock()
	q.items = append(q.items, req)
	subs := append([]chan JobRequest(nil), q.subs...)
	q.mu.Unlock()
	for _, sub := range subs {
		select {
		case sub <- req:
		default:
		}
	}
}

// Subscribe returns a channel notified on enqueue.
func (q *Queue) Subscribe() <-chan JobRequest {
	ch := make(chan JobRequest, 32)
	q.mu.Lock()
	q.subs = append(q.subs, ch)
	q.mu.Unlock()
	return ch
}

// Drain returns queued items and clears the queue.
func (q *Queue) Drain() []JobRequest {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := q.items
	q.items = nil
	return out
}

// Worker processes jobs from the queue until ctx is canceled.
type Worker struct {
	queue   *Queue
	handler func(context.Context, JobRequest) error
}

// NewWorker creates a queue worker.
func NewWorker(queue *Queue, handler func(context.Context, JobRequest) error) *Worker {
	return &Worker{queue: queue, handler: handler}
}

// Run starts processing jobs.
func (w *Worker) Run(ctx context.Context) error {
	sub := w.queue.Subscribe()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case req := <-sub:
			if err := w.handler(ctx, req); err != nil {
				return err
			}
		}
	}
}
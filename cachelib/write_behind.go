package cachelib

import (
	"sync"
	"time"
)

type WriteTask struct {
	Key   string
	Value string
}

type WriteQueue struct {
	queue     []WriteTask
	mu        sync.Mutex
	processor func(task WriteTask)
}

func NewWriteQueue(processor func(task WriteTask)) *WriteQueue {
	wq := &WriteQueue{queue: make([]WriteTask, 0), processor: processor}
	go wq.startWorker()
	return wq
}

func (w *WriteQueue) Enqueue(task WriteTask) {
	w.mu.Lock()
	w.queue = append(w.queue, task)
	w.mu.Unlock()
}

func (w *WriteQueue) startWorker() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		w.mu.Lock()
		if len(w.queue) > 0 {
			task := w.queue[0]
			w.queue = w.queue[1:]
			w.mu.Unlock()
			w.processor(task)
		} else {
			w.mu.Unlock()
		}
	}
}

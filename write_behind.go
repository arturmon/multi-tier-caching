package multi_tier_caching

import (
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type WriteTask struct {
	Key   string
	Value string
	TTL   time.Duration
}

type WriteQueue struct {
	debug     bool
	queue     []WriteTask
	mu        sync.Mutex
	cond      *sync.Cond
	processor func(task WriteTask)
	stopChan  chan struct{}
	interval  time.Duration
}

func NewWriteQueue(processor func(task WriteTask), debug bool) *WriteQueue {
	wq := &WriteQueue{
		queue:     make([]WriteTask, 0),
		processor: processor,
		stopChan:  make(chan struct{}),
		debug:     debug,
		interval:  600 * time.Millisecond,
	}
	wq.cond = sync.NewCond(&wq.mu)
	go wq.startWorker()
	prometheus.MustRegister(queueLengthGauge, processedTasksCounter, taskProcessingHistogram)
	return wq
}

func (w *WriteQueue) Enqueue(task WriteTask) {
	if w.debug {
		log.Printf("[WRITE QUEUE] Enqueuing task for key=%s", task.Key)
	}
	w.mu.Lock()
	w.queue = append(w.queue, task)
	queueLengthGauge.Set(float64(len(w.queue)))
	w.updateInterval()
	w.mu.Unlock()
	w.cond.Signal()
}

func (w *WriteQueue) startWorker() {
	var timer *time.Timer
	for {
		w.mu.Lock()
		for len(w.queue) == 0 {
			w.cond.Wait()
		}
		w.updateInterval()
		timer = time.NewTimer(w.interval)
		w.mu.Unlock()

		select {
		case <-timer.C:
			w.mu.Lock()
			if len(w.queue) == 0 {
				w.mu.Unlock()
				continue
			}
			task := w.queue[0]
			w.queue = w.queue[1:]
			w.updateInterval()
			w.mu.Unlock()
			startTime := time.Now()
			w.processor(task)
			taskProcessingHistogram.Observe(time.Since(startTime).Seconds())
			processedTasksCounter.Inc()
			queueLengthGauge.Set(float64(len(w.queue)))
		case <-w.stopChan:
			timer.Stop()
			return
		}
	}
}

func (w *WriteQueue) updateInterval() {
	queueLen := len(w.queue)
	switch {
	case queueLen > 10:
		w.interval = 80 * time.Millisecond
	case queueLen > 7:
		w.interval = 200 * time.Millisecond
	case queueLen > 4:
		w.interval = 400 * time.Millisecond
	default:
		w.interval = 600 * time.Millisecond
	}
}

func (w *WriteQueue) Stop() {
	close(w.stopChan)
}

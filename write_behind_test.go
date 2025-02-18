package multi_tier_caching

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriteQueue(t *testing.T) {
	var processedTasks []WriteTask
	var mu sync.Mutex

	processor := func(task WriteTask) {
		mu.Lock()
		processedTasks = append(processedTasks, task)
		mu.Unlock()
	}

	wq := NewWriteQueue(processor)

	task1 := WriteTask{Key: "key1", Value: "value1"}
	task2 := WriteTask{Key: "key2", Value: "value2"}

	wq.Enqueue(task1)
	wq.Enqueue(task2)

	// Wait for the queue to process tasks
	time.Sleep(2 * time.Second)

	mu.Lock()
	defer mu.Unlock()

	assert.Len(t, processedTasks, 2, "There are 2 tasks to be processed")
	assert.Equal(t, "key1", processedTasks[0].Key, "The first task must be key1")
	assert.Equal(t, "value1", processedTasks[0].Value, "The first task must contain value1")
	assert.Equal(t, "key2", processedTasks[1].Key, "The second task must be key2")
	assert.Equal(t, "value2", processedTasks[1].Value, "The second task must contain value2")
}

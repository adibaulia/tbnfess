package queue

import (
	"sync"
)

type Item []byte

type Queue struct {
	items []Item
	mutex sync.Mutex
}

func (queue *Queue) Enqueue(item Item) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	queue.items = append(queue.items, item)
}

func (queue *Queue) Len() int {
	return len(queue.items)
}

func (queue *Queue) Dequeue() []byte {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if len(queue.items) == 0 {
		return nil
	}

	lastItem := queue.items[0]
	queue.items = queue.items[1:]

	return lastItem
}

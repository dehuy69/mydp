package service

import (
	"fmt"

	"github.com/gammazero/deque"
)

// QueueManager to manage multiple queues
type QueueManager struct {
	queues map[string]*deque.Deque[map[string]interface{}]
}

// NewQueueManager creates a new QueueManager
func NewQueueManager() *QueueManager {
	return &QueueManager{
		queues: make(map[string]*deque.Deque[map[string]interface{}]),
	}
}

// GetOrCreateQueue retrieves an existing queue or creates a new one if not exists
func (qm *QueueManager) GetOrCreateQueue(name string) *deque.Deque[map[string]interface{}] {
	if q, ok := qm.queues[name]; ok {
		return q
	}

	q := deque.New[map[string]interface{}]()
	qm.queues[name] = q
	return q
}

// AddToQueue adds an item to a specific queue by name
func (qm *QueueManager) AddToQueue(name string, item map[string]interface{}) {
	queue := qm.GetOrCreateQueue(name)
	queue.PushBack(item)
}

// GetFromQueue retrieves and removes an item from a specific queue by name
func (qm *QueueManager) GetFromQueue(name string) map[string]interface{} {
	queue, ok := qm.queues[name]
	if !ok {
		return nil
	}

	item := queue.PopFront()

	return item
}

func (qm *QueueManager) GetAllCurrentQueueAndTheirFirstData() map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	fmt.Println("DEBUG: GetAllCurrentQueueAndTheirFirstData queues", qm.queues)
	for name, queue := range qm.queues {
		result[name] = queue.Front()
	}
	return result
}

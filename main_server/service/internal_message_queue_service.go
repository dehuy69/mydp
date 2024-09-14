package service

// Queue struct
type Queue struct {
	items []interface{}
}

// NewQueue creates a new Queue
func NewQueue() *Queue {
	return &Queue{
		items: []interface{}{},
	}
}

// Enqueue adds an item to the queue
func (q *Queue) Enqueue(item interface{}) {
	q.items = append(q.items, item)
}

// Dequeue removes and returns the first item from the queue
func (q *Queue) Dequeue() interface{} {
	if len(q.items) == 0 {
		return nil
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

// QueueManager to manage multiple queues
type QueueManager struct {
	queues map[string]*Queue
}

// NewQueueManager creates a new QueueManager
func NewQueueManager() *QueueManager {
	return &QueueManager{
		queues: make(map[string]*Queue),
	}
}

// GetOrCreateQueue retrieves an existing queue or creates a new one if not exists
func (qm *QueueManager) GetOrCreateQueue(name string) *Queue {
	if queue, exists := qm.queues[name]; exists {
		return queue
	}
	newQueue := NewQueue()
	qm.queues[name] = newQueue
	return newQueue
}

// AddToQueue adds an item to a specific queue by name
func (qm *QueueManager) AddToQueue(name string, item interface{}) {
	queue := qm.GetOrCreateQueue(name)
	queue.Enqueue(item)
}

// GetFromQueue retrieves and removes an item from a specific queue by name
func (qm *QueueManager) GetFromQueue(name string) interface{} {
	queue := qm.GetOrCreateQueue(name)
	return queue.Dequeue()
}

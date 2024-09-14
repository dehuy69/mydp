package consumer

import (
	"log"
	"time"

	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/main_server/service"
)

type WriteCollectionConsumer struct {
	queueManager *service.QueueManager
	queueName    string
	stopChan     chan struct{}
}

func NewWriteCollectionConsumer(cfg *config.Config) (*WriteCollectionConsumer, error) {
	qm := service.NewQueueManager()
	return &WriteCollectionConsumer{
		queueManager: qm,
		queueName:    "write-collection",
		stopChan:     make(chan struct{}),
	}, nil
}

func (cs *WriteCollectionConsumer) Start() error {
	log.Println("Consumer service started")
	for {
		select {
		case <-cs.stopChan:
			return nil
		default:
			item := cs.queueManager.GetFromQueue(cs.queueName)
			if item != nil {
				log.Printf("Message received: %v", item)
				// Xử lý message write data vào collection ở đây
				// Create collection wrapper
				// Write data to collection

			} else {
				time.Sleep(1 * time.Second) // Sleep một chút để tránh vòng lặp busy-wait
			}
		}
	}
}

func (cs *WriteCollectionConsumer) Shutdown() error {
	log.Println("Consumer service shutting down")
	close(cs.stopChan)
	return nil
}

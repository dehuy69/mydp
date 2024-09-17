package consumer

import (
	"log"
	"time"

	"github.com/dehuy69/mydp/main_server/service"
)

type WriteCollectionConsumer struct {
	SQLiteCatalogService *service.SQLiteCatalogService
	BadgerService        *service.BadgerService
	QueueManager         *service.QueueManager
	queueName            string
	stopChan             chan struct{}
}

func NewWriteCollectionConsumer(SQLiteCatalogService *service.SQLiteCatalogService,
	BadgerService *service.BadgerService,
	QueueManager *service.QueueManager,
) (*WriteCollectionConsumer, error) {

	queueManager := service.NewQueueManager()
	return &WriteCollectionConsumer{
		queueName:            "write-collection",
		stopChan:             make(chan struct{}),
		SQLiteCatalogService: SQLiteCatalogService,
		BadgerService:        BadgerService,
		QueueManager:         queueManager,
	}, nil
}

func (cs *WriteCollectionConsumer) Start() error {
	log.Println("Consumer service started")
	for {
		select {
		case <-cs.stopChan:
			return nil
		default:
			item := cs.QueueManager.GetFromQueue(cs.queueName)
			if item != nil {
				log.Printf("Message received: %v", item)
				// Xử lý message write data vào collection ở đây
				// Get collection from catalog
				// collection := cs.SQLiteCatalogService.GetCollection(item["_collection_id"].(string))
				// Create collection wrapper
				// Write data to collection
				// wrapper := domain.NewCollectionWrapper(cs.SQLiteCatalogService, cs.SQLiteIndexService, item.Collection, cs.BadgerService)

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

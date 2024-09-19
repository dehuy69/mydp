package consumer

import (
	"fmt"
	"log"
	"time"

	"github.com/dehuy69/mydp/main_server/domain"
	"github.com/dehuy69/mydp/main_server/service"
)

type WriteCollectionConsumer struct {
	SQLiteCatalogService *service.SQLiteCatalogService
	BadgerService        *service.BadgerService
	QueueManager         *service.QueueManager
	BboltService         *service.BboltService
	queueName            string
	stopChan             chan struct{}
}

func NewWriteCollectionConsumer(SQLiteCatalogService *service.SQLiteCatalogService, BadgerService *service.BadgerService, QueueManager *service.QueueManager, BboltService *service.BboltService) (*WriteCollectionConsumer, error) {

	return &WriteCollectionConsumer{
		queueName:            "write-collection",
		stopChan:             make(chan struct{}),
		SQLiteCatalogService: SQLiteCatalogService,
		BadgerService:        BadgerService,
		QueueManager:         QueueManager,
		BboltService:         BboltService,
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
				collectionID := item["_collection_id"].(int)

				// Xử lý message write data vào collection ở đây
				// Retrieve collection
				collection, err := cs.SQLiteCatalogService.GetCollectionByID(collectionID)
				if err != nil {
					return fmt.Errorf("failed to get collection: %v", err)
				}

				// Remove _connection_id from itemMap
				delete(item, "_collection_id")

				// Create collection wrapper
				// Write data to collection
				wrapper := domain.NewCollectionWrapper(collection, cs.SQLiteCatalogService, cs.BadgerService, cs.BboltService)
				if err := wrapper.Write(item); err != nil {
					return fmt.Errorf("failed to write collection: %v", err)
				}

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

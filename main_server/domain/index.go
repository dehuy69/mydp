package domain

import (
	"github.com/dehuy69/mydp/main_server/models"
	service "github.com/dehuy69/mydp/main_server/service"
)

type IndexWrapper struct {
	SQLiteCatalogService *service.SQLiteCatalogService // Kết nối cơ sở dữ liệu
	Index                *models.Index                 // Chứa đối tượng Collection từ models
	SQLiteIndexService   *service.SQLiteIndexService   // Kết nối cơ sở dữ liệu
}

// NewCollectionWrapper khởi tạo một instance mới của CollectionWrapper
func NewIndexWrapper(index *models.Index, SQLiteCatalogService *service.SQLiteCatalogService, SQLiteIndexService *service.SQLiteIndexService) *IndexWrapper {
	return &IndexWrapper{
		SQLiteCatalogService: SQLiteCatalogService,
		Index:                index,
		SQLiteIndexService:   SQLiteIndexService,
	}
}

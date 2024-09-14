package domain

import (
	"log"

	"github.com/dehuy69/mydp/main_server/models"
	service "github.com/dehuy69/mydp/main_server/service"
)

type IndexWrapper struct {
	SQLiteCatalogService *service.SQLiteCatalogService // Kết nối cơ sở dữ liệu
	Index                *models.Index                 // Chứa đối tượng Collection từ models
	SQLiteIndexService   *service.SQLiteIndexService   // Kết nối cơ sở dữ liệu
}

// NewIndexWrapper khởi tạo một instance mới của IndexWrapper
// Một index sẽ có tên bảng là động nhưng cấu trúc của bảng sẽ giống nhau
// value: int/string, keys: string
func NewIndexWrapper(index *models.Index, SQLiteCatalogService *service.SQLiteCatalogService, SQLiteIndexService *service.SQLiteIndexService) *IndexWrapper {
	return &IndexWrapper{
		SQLiteCatalogService: SQLiteCatalogService,
		Index:                index,
		SQLiteIndexService:   SQLiteIndexService,
	}
}

// Query 1 record với giá trị value
func (iw *IndexWrapper) Query(value interface{}) ([]models.IndexTableStruct, error) {
	dbclient, err := iw.SQLiteIndexService.GetConnection(iw.Index.Collection.ID)
	if err != nil {
		return nil, err
	}
	// Truy vấn tất cả các bản ghi từ bảng có tên động
	var indexTableRecords []models.IndexTableStruct
	if err := dbclient.Table(iw.Index.Name).Where("value = ?", value).Find(&indexTableRecords).Error; err != nil {
		log.Fatalf("failed to query records: %v", err)
	}

	return indexTableRecords, nil

}

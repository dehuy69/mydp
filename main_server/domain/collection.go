package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dehuy69/mydp/main_server/models"
	service "github.com/dehuy69/mydp/main_server/service"
)

// CollectionWrapper là struct bọc để thêm các phương thức vào Collection
type CollectionWrapper struct {
	SQLiteCatalogService *service.SQLiteCatalogService // Kết nối cơ sở dữ liệu
	SQLiteIndexService   *service.SQLiteIndexService   // Kết nối đến SQLite index service
	Collection           *models.Collection            // Chứa đối tượng Collection từ models
	BadgerService        *service.BadgerService        // Kết nối cơ sở dữ liệu
}

// NewCollectionWrapper khởi tạo một instance mới của CollectionWrapper
func NewCollectionWrapper(SQLiteCatalogService *service.SQLiteCatalogService, SQLiteIndexService *service.SQLiteIndexService, collection *models.Collection, BadgerService *service.BadgerService) *CollectionWrapper {
	return &CollectionWrapper{
		SQLiteCatalogService: SQLiteCatalogService,
		SQLiteIndexService:   SQLiteIndexService,
		Collection:           collection,
		BadgerService:        BadgerService,
	}
}

// Create Collection in catalog
func (cw *CollectionWrapper) CreateCollectionInCatalog(collectionName string) error {
	// Tạo một collection mới
	collection := models.Collection{
		Name: collectionName,
	}

	// Tạo collection trong cơ sở dữ liệu
	if err := cw.SQLiteCatalogService.CreateCollection(&collection); err != nil {
		return fmt.Errorf("failed to create collection in catalog: %v", err)
	}

	return nil

}

// Write dữ liệu vào collection với input là một map bất kỳ
func (cw *CollectionWrapper) Write(input map[string]interface{}) error {
	error := cw.writeBadger(input)
	if error != nil {
		return error
	}

	// write index
	// Tìm	tất cả các index của collection
	for _, index := range cw.Collection.Indexes {
		// Tạo một instance của IndexWrapper
		indexWrapper := NewIndexWrapper(&index, cw.SQLiteCatalogService, cw.SQLiteIndexService)
		err := indexWrapper.Insert(input)
		if err != nil {
			return fmt.Errorf("failed to insert record into index: %v", err)
		}
	}

	return nil
}

func (cw *CollectionWrapper) writeBadger(input map[string]interface{}) error {
	// Lấy giá trị của trường `_key` từ input map
	keyField, ok := input["_key"]
	if !ok {
		return fmt.Errorf("input map must contain a '_key' field")
	}

	// Tạo key bằng cách kết hợp ID collection và giá trị của trường `_key`
	combinedKey := fmt.Sprintf("%d_%v", cw.Collection.ID, keyField)

	// Chuyển đổi input map thành chuỗi JSON
	valueBytes, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input map to JSON: %v", err)
	}

	// Ghi dữ liệu vào Badger với key và value
	err = cw.BadgerService.Set([]byte(combinedKey), valueBytes)
	if err != nil {
		return fmt.Errorf("failed to write data to Badger: %v", err)
	}

	return nil
}

// Read đọc dữ liệu từ collection với key
func (cw *CollectionWrapper) Read(key string) (map[string]interface{}, error) {
	// Tạo key bằng cách kết hợp ID collection và key
	combinedKey := fmt.Sprintf("%d_%v", cw.Collection.ID, key)

	// Đọc dữ liệu từ Badger với key
	valueBytes, err := cw.BadgerService.Get([]byte(combinedKey))
	if err != nil {
		return nil, fmt.Errorf("failed to read data from Badger: %v", err)
	}

	// Chuyển đổi chuỗi JSON thành map
	var valueMap map[string]interface{}
	if err := json.Unmarshal(valueBytes, &valueMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %v", err)
	}

	return valueMap, nil
}

// Function kiểm tra dữ liệu ghi vào collection có thỏa các ràng buộc của index
func (cw *CollectionWrapper) CheckIndexConstraints(input map[string]interface{}) error {
	// Tìm tất cả các index của collection có is_unique = true
	uniqueIndexes := make([]models.Index, 0)
	for _, index := range cw.Collection.Indexes {
		if index.IsUnique {
			uniqueIndexes = append(uniqueIndexes, index)
		}
	}

	// Nếu không có index nào có is_unique = true, không cần kiểm tra ràng buộc
	if len(uniqueIndexes) == 0 {
		return nil
	}

	// Lấy kết nối đến SQLite index service của collection
	dbclient_collection, err := cw.SQLiteIndexService.GetConnection(cw.Collection.ID)
	if err != nil {
		return fmt.Errorf("failed to get connection to SQLite index service: %v", err)
	}

	// Kiểm tra ràng buộc của từng index
	for _, index := range uniqueIndexes {
		// Loại index có nhiều hơn 2 field, contruct value từ các field
		indexFields := strings.Split(index.Fields, ",")
		value := ""
		if len(indexFields) > 1 {
			for _, field := range indexFields {
				value += fmt.Sprintf("%v", input[string(field)])
			}
		} else {
			value = fmt.Sprintf("%v", input[index.Fields])
		}

		// Truy vấn tất cả các bản ghi từ bảng index có tên động
		var indexTableRecords []models.IndexTableStruct
		if err := dbclient_collection.Table(index.Name).Where("value = ?", value).Find(&indexTableRecords).Error; err != nil {
			return fmt.Errorf("failed to query records: %v", err)
		}

		// Nếu có bản ghi, và giá trị keys giống với key trong input, trả về lỗi
		if len(indexTableRecords) > 0 {
			indexTableRecord := indexTableRecords[0]
			if indexTableRecord.Keys == input["_key"] {
				return fmt.Errorf("unique constraint violation on index %s", index.Name)
			}
		}
	}
	return nil
}

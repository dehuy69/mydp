package domain

import (
	"encoding/json"
	"fmt"

	"github.com/dehuy69/mydp/main_server/models"
	"github.com/dehuy69/mydp/main_server/service"
)

// CollectionWrapper là struct bọc để thêm các phương thức vào Collection
type CollectionWrapper struct {
	Collection    *models.Collection     // Chứa đối tượng Collection từ models
	BadgerService *service.BadgerService // Kết nối cơ sở dữ liệu
}

// NewCollectionWrapper khởi tạo một instance mới của CollectionWrapper
func NewCollectionWrapper(collection *models.Collection, BadgerService *service.BadgerService) *CollectionWrapper {
	return &CollectionWrapper{
		Collection:    collection,
		BadgerService: BadgerService,
	}
}

// Write dữ liệu vào collection với input là một map bất kỳ
func (cw *CollectionWrapper) Write(input map[string]interface{}) error {
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

	// Nếu có index, thêm dữ liệu vào index
	// if len(cw.Collection.Indexes) > 0 {
	// 	// Tạo key cho index bằng cách kết hợp ID index và giá trị của trường `_key`
	// 	combinedIndexKey := fmt.Sprintf("%d_%v", cw.Collection.Index.ID, keyField)

	// 	// Ghi dữ liệu vào Badger với key và value
	// 	err = cw.BadgerService.Set([]byte(combinedIndexKey), valueBytes)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to write data to Badger: %v", err)
	// 	}
	// }

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

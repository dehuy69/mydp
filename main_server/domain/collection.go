package domain

import (
	"encoding/json"
	"fmt"

	"github.com/dehuy69/mydp/main_server/models"
	service "github.com/dehuy69/mydp/main_server/service"
)

// CollectionWrapper là struct bọc để thêm các phương thức vào Collection
type CollectionWrapper struct {
	SQLiteCatalogService *service.SQLiteCatalogService // Kết nối cơ sở dữ liệu
	Collection           *models.Collection            // Chứa đối tượng Collection từ models
	BadgerService        *service.BadgerService        // Kết nối cơ sở dữ liệu
	BboltService         *service.BboltService         // Kết nối cơ sở dữ liệu
}

// NewCollectionWrapper khởi tạo một instance mới của CollectionWrapper
func NewCollectionWrapper(collection *models.Collection, SQLiteCatalogService *service.SQLiteCatalogService,
	BadgerService *service.BadgerService, BboltService *service.BboltService) *CollectionWrapper {
	// Kiểm tra tồn tại trước khi preload
	fmt.Println("Collection ID:", collection.ID)
	if collection.ID != 0 {
		// Preload các liên kết thủ công
		preload := SQLiteCatalogService.Db.Preload("Workspace").Preload("Indexes").Preload("Shards").First(collection)
		err := preload.Error
		if err != nil {
			fmt.Println("Error preloading collection:", err)
			return nil
		}
	} else {
		fmt.Println("Collection ID is 0, không cần preload")
	}

	wrapper := CollectionWrapper{
		SQLiteCatalogService: SQLiteCatalogService,
		Collection:           collection,
		BadgerService:        BadgerService,
		BboltService:         BboltService,
	}
	return &wrapper
}

// Create Collection in catalog
func (cw *CollectionWrapper) CreateCollection() error {
	// Lấy server localhost
	server, err := cw.SQLiteCatalogService.GetServerByHost("localhost")
	if err != nil {
		return fmt.Errorf("failed to get server: %v", err)
	}

	cw.Collection.ShardKey = "_key"

	// Tạo collection trong catalog
	if err := cw.SQLiteCatalogService.CreateCollection(cw.Collection); err != nil {
		return fmt.Errorf("failed to create collection in catalog: %v", err)
	}

	// Tạo Shard default cho collection
	shard := models.Shard{
		CollectionID: cw.Collection.ID,
		ShardNumber:  0,
		ShardKey:     "_key",
		Status:       "active",
		ServerID:     server.ID,
	}
	err = cw.SQLiteCatalogService.CreateShard(&shard)
	if err != nil {
		return fmt.Errorf("func %s, failed to create shard in catalog: %v", "CreateCollection", err)
	}

	return nil

}

// Write dữ liệu vào collection với input là một map bất kỳ
func (cw *CollectionWrapper) Write(input map[string]interface{}) error {
	// Gọi GetByKey Kiểm tra _key trong input có tồn tại chưa, nếu có rồi thì gọi qua update để cập nhật dữ liệu
	_, err := cw.BadgerService.Get([]byte(cw.CreateBadgerKey(input["_key"].(string))))
	if err == nil {
		// Nếu không có lỗi, tức là đã tồn tại dữ liệu, gọi qua update
		return fmt.Errorf("record already exists")
	}

	// write index
	// Tìm	tất cả các index của collection
	for _, index := range cw.Collection.Indexes {
		// Tạo một instance của IndexWrapper
		indexWrapper := NewIndexWrapper(&index, cw.SQLiteCatalogService, cw.BadgerService, cw.BboltService)
		err := indexWrapper.InsertWithCheckingStatus(input)
		if err != nil {
			fmt.Println("DEBUG (cw *CollectionWrapper) Write ", err)
			return fmt.Errorf("failed to insert record into index: %v", err)
		}

	}

	// Ghi dữ liệu vào badger
	error := cw.writeData(input)
	if error != nil {
		return error
	}

	return nil
}

func (cw *CollectionWrapper) writeData(input map[string]interface{}) error {
	// Lấy giá trị của trường `_key` từ input map
	keyField, ok := input["_key"]
	if !ok {
		return fmt.Errorf("input map must contain a '_key' field")
	}

	// Chuyển đổi input map thành chuỗi JSON
	valueBytes, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input map to JSON: %v", err)
	}

	// Ghi dữ liệu vào Badger với key và value
	keyFieldStr, ok := keyField.(string)
	if !ok {
		return fmt.Errorf("keyField must be a string")
	}
	err = cw.BadgerService.Set([]byte(cw.CreateBadgerKey(keyFieldStr)), valueBytes)
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

	// Kiểm tra ràng buộc của từng index
	for _, index := range uniqueIndexes {
		// Tạo wrapper cho index
		indexWrapper := NewIndexWrapper(&index, cw.SQLiteCatalogService, cw.BadgerService, cw.BboltService)
		err := indexWrapper.CheckIndexConstraints(input)
		if err != nil {
			return err
		}

	}
	return nil
}

// Tạo key badger
func (cw *CollectionWrapper) CreateBadgerKey(inputKey string) string {
	return fmt.Sprintf("%d||%v", cw.Collection.ID, inputKey)
}

// Kiểm tra exist key
func (cw *CollectionWrapper) ExistKey(inputKey string) bool {
	_, err := cw.BadgerService.Get([]byte(cw.CreateBadgerKey(inputKey)))
	return err == nil
}

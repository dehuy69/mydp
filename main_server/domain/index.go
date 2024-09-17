package domain

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dehuy69/mydp/main_server/models"
	service "github.com/dehuy69/mydp/main_server/service"
)

type IndexWrapper struct {
	SQLiteCatalogService *service.SQLiteCatalogService // Kết nối cơ sở dữ liệu
	Index                *models.Index                 // Chứa đối tượng Collection từ models
	SQLiteIndexService   *service.SQLiteIndexService   // Kết nối cơ sở dữ liệu
	BadgerService        *service.BadgerService
}

// NewIndexWrapper khởi tạo một instance mới của IndexWrapper
// Một index sẽ có tên bảng là động nhưng cấu trúc của bảng sẽ giống nhau
// value: int/string, keys: string
func NewIndexWrapper(index *models.Index, SQLiteCatalogService *service.SQLiteCatalogService, SQLiteIndexService *service.SQLiteIndexService, BadgerService *service.BadgerService) *IndexWrapper {
	if index.ID != 0 {
		// Preload các liên kết thủ công
		preload := SQLiteCatalogService.Db.Preload("Collection").First(index)
		err := preload.Error
		if err != nil {
			fmt.Println("Error preloading collection:", err)
			return nil
		}
	} else {
		fmt.Println("Index ID is 0, không cần preload")
	}

	return &IndexWrapper{
		SQLiteCatalogService: SQLiteCatalogService,
		Index:                index,
		SQLiteIndexService:   SQLiteIndexService,
		BadgerService:        BadgerService,
	}
}

// Taọ index, tạo index trong catalog -> tạo index trong indexService
func (iw *IndexWrapper) CreateIndex() error {
	// Set status của index là building
	iw.Index.Status = models.IndexStatusBuilding

	// Tạo index trong catalog
	err := iw.SQLiteCatalogService.CreateIndex(iw.Index)
	if err != nil {
		return fmt.Errorf("failed to create index: %v", err)
	}

	// Retrive index
	// Truy vấn với Preload để tải dữ liệu của Collection
	err = iw.SQLiteCatalogService.GetModelWithAllAssociations(iw.Index, iw.Index.ID)
	if err != nil {
		return fmt.Errorf("failed to get index: %v", err)
	}

	// Gọi index service, tạo index
	err = iw.SQLiteIndexService.CreateIndex(iw.Index)
	if err != nil {
		return fmt.Errorf("failed to create index: %v", err)
	}

	// Scan dữ liệu hiện có trong collection, insert vào index
	err = iw.ScanDataAndInsert()
	if err != nil {
		return fmt.Errorf("failed to scan data and insert: %v", err)
	}

	// Set status của index là ready
	iw.Index.Status = models.IndexStatusActive
	err = iw.SQLiteCatalogService.UpdateIndex(iw.Index)
	if err != nil {
		return fmt.Errorf("failed to update index: %v", err)
	}

	// Thêm data cache vào index
	err = iw.addCacheToIndex()
	if err != nil {
		// Nếu là lỗi không có file cache, không cần xử lý
		if strings.Contains(err.Error(), "no such file or directory") {
			return nil
		}
		return fmt.Errorf("failed to add cache to index: %v", err)
	}

	return nil
}

func (iw *IndexWrapper) ScanDataAndInsert() error {
	// query bảng <collection_name>
	keys, err := iw.SQLiteIndexService.FindKeys(iw.Index.Collection.Name)
	if err != nil {
		return fmt.Errorf("failed to find keys: %v", err)
	}

	// lặp qua các key và chèn data vào index
	for _, key := range keys {
		// Lấy data từ badger
		input, err := iw.BadgerService.Get([]byte(key))
		if err != nil {
			return fmt.Errorf("failed to get data from badger: %v", err)
		}

		// Unmarshal input từ []byte sang map[string]interface{}
		var inputData map[string]interface{}
		if err := json.Unmarshal(input, &inputData); err != nil {
			return fmt.Errorf("failed to unmarshal input: %v", err)
		}
		// Insert data vào index
		err = iw.Insert(inputData)
		if err != nil {
			return fmt.Errorf("failed to insert data: %v", err)
		}

	}
	return nil
}

// Đọc dữ liệu từ cache của index
func (iw *IndexWrapper) readCache() ([]map[string]interface{}, error) {
	// Đường dẫn tới file cache
	cachePath := fmt.Sprintf("data/cache/collection_%s/index_%s/data.json", iw.Index.Collection.Name, iw.Index.Name)

	// Mở file cache
	f, err := os.Open(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache file: %v", err)
	}
	defer f.Close()

	// Đọc dữ liệu từ file cache
	var data []map[string]interface{}
	dec := json.NewDecoder(f)
	for {
		var record map[string]interface{}
		if err := dec.Decode(&record); err != nil {
			break
		}
		data = append(data, record)
	}
	return data, nil

}

// Thêm data cache vào index
func (iw *IndexWrapper) addCacheToIndex() error {
	// Đọc dữ liệu từ cache
	data, err := iw.readCache()
	if err != nil {
		return fmt.Errorf("failed to read cache: %v", err)
	}
	for _, record := range data {
		err := iw.Insert(record)
		if err != nil {
			return fmt.Errorf("failed to insert record: %v", err)
		}
	}
	// Xóa file cache
	// cachePath := fmt.Sprintf("data/cache/collection_%s/index_%s/data.json", iw.Index.Collection.Name, iw.Index.Name)
	// if err := os.Remove(cachePath); err != nil {
	// 	return fmt.Errorf("failed to remove cache file: %v", err)
	// }
	return nil
}

// Query 1 record với giá trị value
func (iw *IndexWrapper) Query(value interface{}) ([]models.IndexTableStruct, error) {
	dbclient, err := iw.SQLiteIndexService.GetConnection(iw.Index.Collection.Name)
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

// Insert 1 input vào bảng index
func (iw *IndexWrapper) Insert(input map[string]interface{}) error {
	// input phải có trường _key
	if _, ok := input["_key"]; !ok {
		return fmt.Errorf("input must contain a '_key' field")
	}

	// Kiểm tra constraints của index
	if err := iw.CheckIndexConstraints(input); err != nil {
		return err
	}

	// Nếu trạng thái status của index là building, thì cache lại trong data/cache/collection_<collection_name>/index_<index_name>/data.json
	if iw.Index.Status == models.IndexStatusBuilding {
		return iw.insertWhenIndexIsBuilding(input)
	}

	// Lấy kết nối tới bảng index
	dbclient, err := iw.SQLiteIndexService.GetConnection(iw.Index.Collection.Name)
	if err != nil {
		return err
	}

	// Lấy giá trị của value
	// Nếu index là loại hỗn hợp (type là hash)), thì giá trị value sẽ là một tổ hợp md5 %s%s của các trường khác nhau
	// Nếu index là loại đơn, thì giá trị value sẽ là giá trị của trường đó
	value := iw.getValueFromInput(input)

	// Lấy giá trị của key. Là giá trị của trường _key trong input
	key := input["_key"].(string)

	// Thêm key vào danh sách keys
	indexOfValue := models.IndexTableStruct{}
	dbclient.Table(iw.Index.Name).Where("value = ?", value).First(indexOfValue)
	// Thêm key vào keys
	keys := iw.appendKey(indexOfValue.Keys, key)

	// Update record
	if err := dbclient.Table(iw.Index.Name).Where("value = ?", value).Update("keys", keys).Error; err != nil {
		log.Fatalf("failed to update record: %v", err)
	}

	return nil
}

func (iw *IndexWrapper) insertWhenIndexIsBuilding(input map[string]interface{}) error {
	// Lấy đường dẫn tới file cache, tạo thư mục nếu chưa tồn tại
	cachePath := fmt.Sprintf("data/cache/collection_%s/index_%s/data.json", iw.Index.Collection.Name, iw.Index.Name)
	if err := os.MkdirAll(filepath.Dir(cachePath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create cache folder: %v", err)
	}

	// Ghi input vào file cache, ghi vào dòng cuối cùng
	f, err := os.Create(cachePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(input); err != nil {
		return fmt.Errorf("failed to write to cache file: %v", err)
	}
	return nil
}

// Hàm chèn thêm một key vào Keys đã có
func (iw *IndexWrapper) appendKey(keys string, key string) string {
	if keys == "" {
		return key
	}
	// Kiểm tra xem key đã tồn tại trong keys chưa
	if strings.Contains(keys, key) {
		return keys
	}
	return fmt.Sprintf("%s,%s", keys, key)
}

// Hàm lấy value từ input
func (iw *IndexWrapper) getValueFromInput(input map[string]interface{}) interface{} {
	if iw.Index.IndexType == models.IndexTypeHash {
		fieldsList := strings.Split(iw.Index.Fields, ",")
		preHashedValue := ""
		for _, field := range fieldsList {
			preHashedValue += input[field].(string)
		}
		md5Value := md5.Sum([]byte(preHashedValue))
		return fmt.Sprintf("%x", md5Value)
	}
	return input[iw.Index.Fields]
}

// Hàm kiểm tra input có thỏa mãn ràng buộc của index không
func (iw *IndexWrapper) CheckIndexConstraints(input map[string]interface{}) error {
	// Loại index có nhiều hơn 2 field, contruct value từ các field
	value := iw.getValueFromInput(input)

	dbclient_collection, err := iw.SQLiteIndexService.GetConnection(iw.Index.Collection.Name)
	if err != nil {
		return fmt.Errorf("failed to get connection to SQLite index service: %v", err)
	}

	// Truy vấn tất cả các bản ghi từ bảng index có tên động
	var indexTableRecords interface{}
	switch iw.Index.DataType {
	case models.DataTypeInt:
		indexTableRecords = models.IndexTableStructInt{}
	case models.DataTypeFloat:
		indexTableRecords = models.IndexTableStructFloat{}
	default:
		indexTableRecords = models.IndexTableStructString{}
	}

	if err := dbclient_collection.Table(iw.Index.Name).Where("value = ?", value).First(&indexTableRecords).Error; err != nil {
		// Nếu không tìm thấy bản ghi, không cần kiểm tra ràng buộc
		if err.Error() == "record not found" {
			return nil
		}
		return fmt.Errorf("failed to query records: %v", err)
	}

	// Nếu có bản ghi, và giá trị keys khac với key trong input, trả về lỗi
	// value, keys
	//     1, 1
	if indexTableRecords != input["_key"] {
		return fmt.Errorf("unique constraint failed for index %s", iw.Index.Name)
	}

	return nil
}

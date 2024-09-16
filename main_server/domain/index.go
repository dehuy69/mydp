package domain

import (
	"crypto/md5"
	"fmt"
	"log"
	"strings"

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

// Insert 1 input vào bảng index
func (iw *IndexWrapper) Insert(input map[string]interface{}) error {
	dbclient, err := iw.SQLiteIndexService.GetConnection(iw.Index.Collection.ID)
	if err != nil {
		return err
	}

	// Lấy giá trị của value
	// Nếu index là loại hỗn hợp (type là hash)), thì giá trị value sẽ là một tổ hợp md5 %s%s của các trường khác nhau
	// Nếu index là loại đơn, thì giá trị value sẽ là giá trị của trường đó
	var value interface{}
	if iw.Index.IndexType == models.IndexTypeHash {
		fieldsList := strings.Split(iw.Index.Fields, ",")
		preHashedValue := ""
		for _, field := range fieldsList {
			preHashedValue += input[field].(string)
		}
		md5Value := md5.Sum([]byte(preHashedValue))
		value = fmt.Sprintf("%x", md5Value)
	} else {
		value = input[iw.Index.Fields]
		// Ép kiểu value về  datatype của index
		value = iw.assertValue(value)
	}

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

// Hàm ép kiểu value về datatype của index
func (iw *IndexWrapper) assertValue(value interface{}) interface{} {
	switch iw.Index.DataType {
	case models.DataTypeInt:
		return value.(int)
	case models.DataTypeFloat:
		return value.(float64)
	default:
		return value.(string)
	}
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

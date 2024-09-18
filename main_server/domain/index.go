package domain

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dehuy69/mydp/main_server/models"
	service "github.com/dehuy69/mydp/main_server/service"
	"github.com/dgraph-io/badger/v4"
)

type IndexWrapper struct {
	SQLiteCatalogService *service.SQLiteCatalogService // Kết nối cơ sở dữ liệu
	Index                *models.Index                 // Chứa đối tượng Collection từ models
	BadgerService        *service.BadgerService
	BboltService         *service.BboltService
}

// type node struct {
// 	isUniqueType bool
// 	key          string
// 	values       []string
// 	uniqueValues mapset.Set[string]
// }

// NewIndexWrapper khởi tạo một instance mới của IndexWrapper
// Một index sẽ có tên bảng là động nhưng cấu trúc của bảng sẽ giống nhau
// value: int/string, keys: string
func NewIndexWrapper(index *models.Index, SQLiteCatalogService *service.SQLiteCatalogService, BadgerService *service.BadgerService, bboltService *service.BboltService) *IndexWrapper {
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
		BadgerService:        BadgerService,
		BboltService:         bboltService,
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

	// Tạo bucket trong bbolt
	err = iw.BboltService.CreateBucketIfNotExists([]byte(fmt.Sprintf("%d", iw.Index.ID)))
	if err != nil {
		return fmt.Errorf("failed to create bucket: %v", err)
	}

	// Scan dữ liệu hiện có trong collection, insert vào index
	err = iw.scanDataAndInsert()
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

func (iw *IndexWrapper) scanDataAndInsert() error {
	// Thực hiện một View transaction để đọc dữ liệu từ Badger
	err := iw.BadgerService.Db.View(func(txn *badger.Txn) error {
		// Tạo một iterator với các tùy chọn mặc định
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close() // Đảm bảo đóng iterator sau khi sử dụng xong

		// Xác định prefix để duyệt qua các key thuộc về một collection cụ thể
		prefix := []byte(fmt.Sprintf("%d||", iw.Index.Collection.ID))

		// Duyệt qua các key bắt đầu với prefix
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			// Lấy giá trị của item và chèn vào index
			err := item.Value(func(value []byte) error {
				// Gọi hàm Insert để chèn dữ liệu vào index
				var record map[string]interface{}
				err := json.Unmarshal(value, &record)
				if err != nil {
					return err
				}
				// Kiểm tra ràng buộc của index
				err = iw.CheckIndexConstraints(record)
				if err != nil {
					return fmt.Errorf("scanDataAndInsert() failed to check index constraints: %v", err)
				}
				// Insert record vào index
				err = iw.insertWithoutCheckingConstraint(record)
				if err != nil {
					return fmt.Errorf("failed to insert record: %v", err)
				}
				return nil
			})

			// Xử lý lỗi nếu có
			if err != nil {
				return err
			}
		}
		return nil
	})

	// Trả về lỗi nếu có trong quá trình xử lý
	if err != nil {
		return err
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
func (iw *IndexWrapper) Query(value interface{}) (map[string]interface{}, error) {
	return nil, nil

}

// Insert 1 input vào index
// Bucket    || Key
// <index-id>||<value> chứa các keys

func (iw *IndexWrapper) Insert(input map[string]interface{}) error {
	// input phải có trường _key
	if _, ok := input["_key"]; !ok {
		return fmt.Errorf("input must contain a '_key' field")
	}

	// Kiểm tra input có field nằm trong index không, nếu không có, thì không cần xử lý
	if _, ok := input[iw.Index.Fields]; !ok {
		return nil
	}

	// Kiểm tra constraints của index
	if err := iw.CheckIndexConstraints(input); err != nil {
		fmt.Println("DEBUG: CheckIndexConstraints failed")
		return err
	}

	// Nếu trạng thái status của index là building, thì cache lại trong data/cache/collection_<collection_name>/index_<index_name>/data.json
	if iw.Index.Status == models.IndexStatusBuilding {
		fmt.Println("DEBUG: Index is building")
		return iw.insertToCache(input)
	}

	return iw.insertWithoutCheckingConstraint(input)
}

func (iw *IndexWrapper) insertWithoutCheckingConstraint(input map[string]interface{}) error {
	// Lấy giá trị của value
	// Nếu index là loại hỗn hợp (type là hash)), thì giá trị value sẽ là một tổ hợp md5 %s%s của các trường khác nhau
	// Nếu index là loại đơn, thì giá trị value sẽ là giá trị của trường đó
	value := iw.getValueFromInput(input)

	// Lấy giá trị của key. Là giá trị của trường _key trong input
	key := input["_key"].(string)

	// Lấy index data từ bbolt
	indexData, err := iw.BboltService.GetAndParseAsArrayString([]byte(fmt.Sprintf("%d", iw.Index.ID)), []byte(fmt.Sprintf("%v", value)))
	if err != nil {
		// Nếu không tìm thấy index data, tạo mới
		if err.Error() == fmt.Sprintf("key %s not found in bucket %d", value, iw.Index.ID) {
			indexData = []string{}
		} else {
			return fmt.Errorf("failed to get index data: %v", err)
		}
	}

	// Ép thành mảng hoặc set, thêm key vào mảng/set, ghi lại vào indexData
	// indexData -> [key1, key2, key3]
	// Nếu index là unique, ép thành set
	if iw.Index.IsUnique {
		indexDataSet := mapset.NewSet[string](
			indexData...,
		)
		indexDataSet.Add(key)
		indexData = indexDataSet.ToSlice()
	} else {
		indexData = append(indexData, key)
	}

	idxbucket := []byte(fmt.Sprintf("%d", iw.Index.ID))
	idxkey := []byte(fmt.Sprintf("%v", value))

	err = iw.BboltService.SetArrayString(idxbucket, idxkey, indexData)
	if err != nil {
		return fmt.Errorf("failed to write to bbolt: %v", err)
	}

	return nil

}

func (iw *IndexWrapper) insertToCache(input map[string]interface{}) error {
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
// func (iw *IndexWrapper) appendKey(keys []byte, key string) ([]string, error) {
// 	// Parse indexData từ dạng JSON list
// 	var keysInIndex []string
// 	if err := json.Unmarshal(keys, &keysInIndex); err != nil {
// 		return nil, fmt.Errorf("failed to parse index data: %v", err)
// 	}

// 	return append(keysInIndex, key), nil
// }

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
	if iw.Index.IsUnique {
		return iw.checkUniqueConstraints(input)
	}
	return nil

}

// check contraints unique của index
func (iw *IndexWrapper) checkUniqueConstraints(input map[string]interface{}) error {
	// Loại index có nhiều hơn 2 field, contruct value từ các field
	value := iw.getValueFromInput(input)

	// Lấy index data trong bbolt theo bucket:key <index-id>||<value>
	indexData, err := iw.BboltService.GetAndParseAsArrayString([]byte(fmt.Sprintf("%d", iw.Index.ID)), []byte(value.(string)))
	if err != nil {
		// Nếu lỗi là không tìm thấy key, không cần kiểm tra ràng buộc
		//  key huy4 not found in bucket 10
		if err.Error() == fmt.Sprintf("key %s not found in bucket %d", value, iw.Index.ID) {
			return nil
		}
		// Hoặc lỗi không tìm thấy bucket, không cần kiểm tra ràng buộc
		// bucket 10 not found
		if err.Error() == fmt.Sprintf("bucket %d not found", iw.Index.ID) {
			return nil
		}
		return fmt.Errorf("failed to get index data: %v", err)
	}
	fmt.Println("DEBUG: checkUniqueConstraints: indexName", iw.Index.Name)
	fmt.Println("DEBUG: checkUniqueConstraints: input", input)
	fmt.Println("DEBUG: Index data: ", indexData)

	// indexData Có dạng [key1, key2, key3]
	// Parse indexData từ dạng JSON list
	// var keysInIndex []string
	// if err := json.Unmarshal(indexData, &keysInIndex); err != nil {
	// 	return fmt.Errorf("failed to parse index data: %v", err)
	// }

	// Kiểm tra xem trong indexData [key1, key2, key3]
	// có chứa key của input không
	inputKey := input["_key"].(string)
	if !contains(indexData, inputKey) {
		fmt.Println("DEBUG violates unique constraint")
		return fmt.Errorf("input violates unique constraint")
	}
	return nil
}

// Hàm phụ trợ để kiểm tra xem một slice có chứa một phần tử cụ thể không
func contains(slice []string, element string) bool {
	fmt.Println("DEBUG: Checking if slice contains element", element)
	for _, e := range slice {
		fmt.Println("DEBUG: Checking element", e)
		if e == element {
			return true
		}
	}
	return false
}

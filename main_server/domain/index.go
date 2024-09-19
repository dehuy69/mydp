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

	// Tạo index trong indexService
	err = iw.BboltService.CreateIndex(iw.Index)
	if err != nil {
		return fmt.Errorf("failed to create index: %v", err)
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
				// Parse giá trị JSON thành `map[string]interface{}`
				var record map[string]interface{}
				err := json.Unmarshal(value, &record)
				if err != nil {
					return err
				}

				// Insert record vào index
				err = iw.insertWithCheckingConstraint(record)
				if err != nil {
					return fmt.Errorf("failed to insert record: %v. Building index fail", err)
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
		err := iw.insertWithCheckingConstraint(record)
		if err != nil {
			return fmt.Errorf("failed to insert record: %v", err)
		}
	}
	return nil
}

// Query 1 record với giá trị value
func (iw *IndexWrapper) Query(value interface{}) (map[string]interface{}, error) {
	return nil, nil

}

// Insert 1 input vào index
// Bucket    || Key
// <index-id>||<value> chứa các keys

func (iw *IndexWrapper) InsertWithCheckingStatus(input map[string]interface{}) error {
	// input phải có trường _key
	if _, ok := input["_key"]; !ok {
		return fmt.Errorf("input must contain a '_key' field")
	}

	// Kiểm tra input có field nằm trong index không, nếu không có, thì không cần xử lý
	if _, ok := input[iw.Index.Fields]; !ok {
		return nil
	}

	// Nếu trạng thái status của index là building, thì cache lại trong data/cache/collection_<collection_name>/index_<index_name>/data.json
	if iw.Index.Status == models.IndexStatusBuilding {
		fmt.Println("DEBUG: Index is building")
		return iw.insertToCache(input)
	}

	return iw.insertWithCheckingConstraint(input)
}

func (iw *IndexWrapper) insertWithCheckingConstraint(input map[string]interface{}) error {
	// Lấy giá trị của value
	// Nếu index là loại hỗn hợp (type là hash)), thì giá trị value sẽ là một tổ hợp md5 %s%s của các trường khác nhau
	// Nếu index là loại đơn, thì giá trị value sẽ là giá trị của trường đó
	value := iw.getValueFromInput(input)

	// Lấy giá trị của key. Là giá trị của trường _key trong input
	key := input["_key"].(string)

	// check constraints
	err := iw.CheckIndexConstraints(input)
	if err != nil {
		return fmt.Errorf("failed to check index constraints: %v", err)
	}

	// Thêm key vào node
	err = iw.AddKeyToNode(value, key)
	if err != nil {
		return fmt.Errorf("failed to add key to node: %v", err)
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

	inputKey := input["_key"].(string)

	// Kiểm tra node, nếu node không tồn tại, không cần kiểm tra ràng buộc
	nodeExist, err := iw.NodeExist(value)
	if err != nil {
		return fmt.Errorf("failed to check node exist: %v", err)
	}
	if !nodeExist {
		return nil
	}

	// So sánh inputKey với các key trong node, nếu tìm thấy inputKey trong node.keys, thì thỏa ràng buộc
	keyExist, err := iw.KeyExist(value, inputKey)
	if err != nil {
		return fmt.Errorf("failed to check key exist: %v", err)
	}
	if keyExist {
		return nil
	} else {
		return fmt.Errorf("input violates unique constraint")
	}
}

// Thêm 1 key vào node, với value là giá trị của key
func (iw *IndexWrapper) AddKeyToNode(value interface{}, key string) error {
	// filename
	filename := iw.BboltService.GetFileNameFromIndex(iw.Index)

	// Ép kiểu value
	valueAsBytes, err := InterfaceToBytes(value)
	if err != nil {
		fmt.Println("Failed to convert value to bytes: ", err)
		return err
	}

	isNodeExist, err := iw.NodeExist(value)
	if err != nil {
		fmt.Println("Failed to check node exist: ", err)
		return err
	}

	if isNodeExist {
		// Lấy node
		node, err := iw.BboltService.GetAndParseAsNode(filename, []byte("default"), valueAsBytes)
		if err != nil {
			fmt.Println("Failed to get node: ", err)
			return err
		}

		// Thêm key vào node
		node.Keys.Add(key)

		// Ghi lại node
		err = iw.BboltService.SetNode(filename, []byte("default"), node)
		if err != nil {
			fmt.Println("Failed to set node: ", err)
			return err
		}
	} else {
		// Tạo node mới
		node := &models.Node{
			Value: valueAsBytes,
			Keys:  mapset.NewSet[string](),
		}
		node.Keys.Add(key)
		// Ghi node
		err = iw.BboltService.SetNode(filename, []byte("default"), node)
		if err != nil {
			fmt.Println("Failed to set node: ", err)
			return err
		}
	}

	return nil
}

// Kiểm tra node exist
func (iw *IndexWrapper) NodeExist(value interface{}) (bool, error) {
	// filename
	filename := iw.BboltService.GetFileNameFromIndex(iw.Index)

	// Ép kiểu value
	valueAsBytes, err := InterfaceToBytes(value)
	if err != nil {
		fmt.Println("Failed to convert value to bytes: ", err)
		return false, err
	}

	// Lấy node
	_, err = iw.BboltService.GetAndParseAsNode(filename, []byte("default"), valueAsBytes)
	// Nếu mã lỗi là 'key %v not found', thì node không tồn tại
	if err != nil && strings.Contains(err.Error(), fmt.Sprintf("key %v not found", value)) {
		return false, nil
	}
	if err != nil {
		fmt.Println("Failed to get node: ", err)
		return false, err
	}

	return true, nil
}

// Kiểm tra key có tồn tại trong node.keys không
func (iw *IndexWrapper) KeyExist(value interface{}, key string) (bool, error) {
	// filename
	filename := iw.BboltService.GetFileNameFromIndex(iw.Index)

	// Ép kiểu value
	valueAsBytes, err := InterfaceToBytes(value)
	if err != nil {
		fmt.Println("Failed to convert value to bytes: ", err)
		return false, err
	}

	// Lấy node
	node, err := iw.BboltService.GetAndParseAsNode(filename, []byte("default"), valueAsBytes)
	if err != nil {
		fmt.Println("Failed to get node: ", err)
		return false, err
	}

	// Kiểm tra nếu node không tồn tại
	if node == nil {
		return false, nil
	}

	// Kiểm tra nếu key tồn tại trong node.Keys
	if node.Keys.Contains(key) {
		return true, nil
	}

	return false, nil
}

// InterfaceToBytes chuyển đổi một interface{} thành []byte dựa trên kiểu của nó
func InterfaceToBytes(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case int:
		// Chuyển đổi int thành []byte
		return []byte(fmt.Sprintf("%d", v)), nil
	case float64:
		// Chuyển đổi float64 thành []byte
		return []byte(fmt.Sprintf("%f", v)), nil
	case string:
		// Chuyển đổi string trực tiếp thành []byte
		return []byte(v), nil
	default:
		// Trường hợp không hỗ trợ kiểu dữ liệu
		return nil, fmt.Errorf("unsupported type: %T", value)
	}
}

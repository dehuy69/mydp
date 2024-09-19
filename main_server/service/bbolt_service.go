package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"go.etcd.io/bbolt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/main_server/models"
)

// BboltService struct đại diện cho một dịch vụ lưu trữ dữ liệu sử dụng bbolt
// BboltService sẽ lưu trữ một map các kết nối đến các cơ sở dữ liệu bbolt
// Tên file có kiểu collection_id_<collection_id>_index_id_<index_id>.db
// Bucket mặc định là "default"

type BboltService struct {
	DbConnection map[string]*bbolt.DB
	cfg          *config.Config
}

func NewBboltService(cfg *config.Config) (*BboltService, error) {
	// Path to bbolt database folder
	pathToDB := path.Join(cfg.DataFolderDefault, "index")

	// Ensure the folder exists
	err := os.MkdirAll(pathToDB, os.ModePerm)
	if err != nil {
		return nil, err
	}

	DbConnection := make(map[string]*bbolt.DB)

	// Walk through the files in the folder
	err = filepath.Walk(pathToDB, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a file (not a directory)
		if !info.IsDir() {
			// Extract the file name from the file path
			fileName := filepath.Base(filePath)

			// Open bbolt database
			db, err := bbolt.Open(filePath, 0666, nil)
			if err != nil {
				return err
			}

			// Save the database connection using fileName as the key
			DbConnection[fileName] = db
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &BboltService{DbConnection: DbConnection, cfg: cfg}, nil
}

// CreateIndex tạo một cơ sở dữ liệu mới với tên file là collection_id_<collection_id>_table_id_<table_id>.db
func (bs *BboltService) CreateIndex(index *models.Index) error {
	// Lấy tên file từ models.Index
	fileName := bs.GetFileNameFromIndex(index)

	// Tạo một cơ sở dữ liệu mới
	pathToDB := path.Join(bs.cfg.DataFolderDefault, "index", fileName)
	db, err := bbolt.Open(pathToDB, 0666, nil)
	if err != nil {
		return err
	}
	// Tạo bucket mặc định
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("default"))
		return err
	})
	if err != nil {
		return err
	}
	// Lưu trữ kết nối đến cơ sở dữ liệu
	bs.DbConnection[fileName] = db
	return nil
}

// Hàm lấy filename từ models.Index
func (bs *BboltService) GetFileNameFromIndex(index *models.Index) string {
	return fmt.Sprintf("collection_id_%d_index_id_%d.db", index.CollectionID, index.ID)
}

// Set dữ liệu vào bbolt database
func (bs *BboltService) Set(filename string, bucket, key, value []byte) error {
	db, ok := bs.DbConnection[filename]
	if !ok {
		return fmt.Errorf("database %s not found", filename)
	}

	err := db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
	return err
}

// Get dữ liệu từ bbolt database
func (bs *BboltService) Get(filename string, bucket, key []byte) ([]byte, error) {
	db, ok := bs.DbConnection[filename]
	if !ok {
		return nil, fmt.Errorf("database %s not found", filename)
	}
	var value []byte
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucket)
		}
		val := b.Get(key)
		if val == nil {
			return fmt.Errorf("key %s not found in bucket %s", key, bucket)
		}
		value = append([]byte{}, val...) // Sao chép giá trị để tránh vấn đề về con trỏ
		return nil
	})
	return value, err
}

// Delete dữ liệu từ bbolt database
func (bs *BboltService) Delete(filename string, bucket, key []byte) error {
	db, ok := bs.DbConnection[filename]
	if !ok {
		return fmt.Errorf("database %s not found", filename)
	}
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		return b.Delete(key)
	})
	return err
}

// GetAndParseAsNode lấy dữ liệu từ bbolt database và parse thành models.Node
func (bs *BboltService) GetAndParseAsNode(filename string, bucket, key []byte) (*models.Node, error) {
	// Lấy dữ liệu từ database
	value, err := bs.Get(filename, bucket, key)
	if err != nil {
		return nil, err
	}

	// Parse giá trị từ database thành slice []string trước
	var keysAsSlice []string
	err = json.Unmarshal(value, &keysAsSlice)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal value as []string: %w", err)
	}

	// Chuyển slice []string thành mapset.Set[string]
	keysAsSet := mapset.NewSet[string]()
	for _, key := range keysAsSlice {
		keysAsSet.Add(key)
	}

	// Trả về Node với key và set đã được parse
	return &models.Node{
		Value: key,       // key là giá trị của node
		Keys:  keysAsSet, // Keys là set của các chuỗi đã được chuyển đổi
	}, nil
}

// SetNode lưu models.Node vào bbolt database
func (bs *BboltService) SetNode(filename string, bucket []byte, node *models.Node) error {
	// Parse models.Node thành JSON
	value, err := json.Marshal(node.Keys)
	if err != nil {
		return err
	}
	return bs.Set(filename, bucket, node.Value, value)
}

// GetAllBbolt lấy tất cả dữ liệu từ tất cả các bucket trong tất cả các cơ sở dữ liệu
func (bs *BboltService) GetAllBbolt() ([]map[string]interface{}, error) {
	// Lưu trữ tất cả dữ liệu từ các bucket
	var data []map[string]interface{}

	// Duyệt qua tất cả các cơ sở dữ liệu
	for _, db := range bs.DbConnection {
		// Duyệt qua tất cả các bucket trong cơ sở dữ liệu
		err := db.View(func(tx *bbolt.Tx) error {
			// Duyệt qua tất cả các bucket cấp cao nhất
			return tx.ForEach(func(bucketName []byte, b *bbolt.Bucket) error {
				fmt.Printf("Bucket: %s\n", bucketName)

				// Tạo một con trỏ để duyệt qua tất cả các cặp key-value trong bucket
				c := b.Cursor()

				// Duyệt qua tất cả các key-value trong bucket
				for k, v := c.First(); k != nil; k, v = c.Next() {
					var record map[string]interface{}
					// Parse giá trị JSON thành `map[string]interface{}`
					err := json.Unmarshal(v, &record)
					if err != nil {
						// Nếu không phải JSON hoặc gặp lỗi khi parse, lưu trữ giá trị dưới dạng chuỗi
						record = map[string]interface{}{
							"connection": db.Path(),
							"bucket":     string(bucketName),
							"key":        string(k),
							"value":      string(v),
						}
					} else {
						// Thêm key vào bản ghi JSON đã parse
						record["_key"] = string(k)
					}
					data = append(data, record)
				}

				return nil
			})
		})

		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// func (bs *BboltService) GetAllBbolt() ([]map[string]interface{}, error) {
// 	// Lưu trữ tất cả dữ liệu từ các bucket
// 	var data []map[string]interface{}

// 	// Duyệt qua tất cả các bucket trong cơ sở dữ liệu
// 	err := bs.Db.View(func(tx *bbolt.Tx) error {
// 		// Duyệt qua tất cả các bucket cấp cao nhất
// 		return tx.ForEach(func(bucketName []byte, b *bbolt.Bucket) error {
// 			fmt.Printf("Bucket: %s\n", bucketName)

// 			// Tạo một con trỏ để duyệt qua tất cả các cặp key-value trong bucket
// 			c := b.Cursor()

// 			// Duyệt qua tất cả các key-value trong bucket
// 			for k, v := c.First(); k != nil; k, v = c.Next() {
// 				var record map[string]interface{}
// 				// Parse giá trị JSON thành `map[string]interface{}`
// 				err := json.Unmarshal(v, &record)
// 				if err != nil {
// 					// Nếu không phải JSON hoặc gặp lỗi khi parse, lưu trữ giá trị dưới dạng chuỗi
// 					record = map[string]interface{}{
// 						"bucket": string(bucketName),
// 						"key":    string(k),
// 						"value":  string(v),
// 					}
// 				} else {
// 					// Thêm key vào bản ghi JSON đã parse
// 					record["_key"] = string(k)
// 				}
// 				data = append(data, record)
// 			}

// 			return nil
// 		})
// 	})

// 	return data, err
// }

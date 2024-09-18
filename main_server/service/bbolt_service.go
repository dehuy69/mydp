package service

import (
	"encoding/json"
	"fmt"
	"path"

	"go.etcd.io/bbolt"

	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/utils"
)

type BboltService struct {
	Db *bbolt.DB
}

func NewBboltService(cfg *config.Config) (*BboltService, error) {
	// Path to bbolt database folder
	path := path.Join(cfg.DataFolderDefault, "index")

	// Create the folders if not exist
	utils.EnsureFolder(path)

	filedb := path + "/indexdata.db"

	// Open bbolt database
	db, err := bbolt.Open(filedb, 0666, nil)
	if err != nil {
		return nil, err
	}

	return &BboltService{Db: db}, nil
}

func (bs *BboltService) Close() error {
	return bs.Db.Close()
}

func (bs *BboltService) Set(bucket, key, value []byte) error {
	err := bs.Db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
	return err
}

func (bs *BboltService) Get(bucket, key []byte) ([]byte, error) {
	var value []byte
	err := bs.Db.View(func(tx *bbolt.Tx) error {
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

func (bs *BboltService) Delete(bucket, key []byte) error {
	err := bs.Db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		return b.Delete(key)
	})
	return err
}

func (bs *BboltService) GetAndParseAsArrayString(bucket, key []byte) ([]string, error) {
	value, err := bs.Get(bucket, key)
	if err != nil {
		return nil, err
	}
	var result []string
	fmt.Println("DEBUG GetAndParseAsArrayString", value)
	err = json.Unmarshal(value, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (bs *BboltService) SetArrayString(bucket, key []byte, value []string) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return bs.Set(bucket, key, valueBytes)
}

func (bs *BboltService) GetAllBbolt() ([]map[string]interface{}, error) {
	// Lưu trữ tất cả dữ liệu từ các bucket
	var data []map[string]interface{}

	// Duyệt qua tất cả các bucket trong cơ sở dữ liệu
	err := bs.Db.View(func(tx *bbolt.Tx) error {
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
						"bucket": string(bucketName),
						"key":    string(k),
						"value":  string(v),
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

	return data, err
}

func (bs *BboltService) CreateBucketIfNotExists(bucketName []byte) error {
	return bs.Db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
}

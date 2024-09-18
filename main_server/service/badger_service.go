package service

import (
	"encoding/json"
	"path"

	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/utils"
	"github.com/dgraph-io/badger/v4"
)

type BadgerService struct {
	Db *badger.DB
}

// NewBadgerService tạo một instance mới của BadgerService và mở kết nối tới cơ sở dữ liệu Badger
func NewBadgerService(cfg *config.Config) (*BadgerService, error) {
	// Path to Badger database folder
	path := path.Join(cfg.DataFolderDefault, "collection", "badger")

	// Create the folders if not exist
	utils.EnsureFolder(path)

	// Open Badger database
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerService{Db: db}, nil
}

// Close đóng kết nối tới cơ sở dữ liệu Badger
func (bs *BadgerService) Close() error {
	return bs.Db.Close()
}

// Set ghi một cặp khóa-giá trị vào cơ sở dữ liệu Badger
func (bs *BadgerService) Set(key, value []byte) error {
	err := bs.Db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
	return err
}

// Get đọc giá trị từ cơ sở dữ liệu Badger dựa trên khóa
func (bs *BadgerService) Get(key []byte) ([]byte, error) {
	var value []byte
	err := bs.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	return value, err
}

// Delete xóa một cặp khóa-giá trị từ cơ sở dữ liệu Badger
func (bs *BadgerService) Delete(key []byte) error {
	err := bs.Db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
	return err
}

func (bs *BadgerService) GetAllBadger() ([]map[string]interface{}, error) {
	var data []map[string]interface{}

	err := bs.Db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var record map[string]interface{}
				err := json.Unmarshal(val, &record)
				if err != nil {
					return err
				}
				data = append(data, record)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return data, err
}

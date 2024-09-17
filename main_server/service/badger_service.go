package service

import (
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

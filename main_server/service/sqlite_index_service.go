package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dehuy69/mydp/main_server/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteIndexService struct {
	DB            *gorm.DB
	BadgerService *BadgerService
	Connections   map[string]*gorm.DB
}

func NewSQLiteIndexService(db *gorm.DB, badgerService *BadgerService) *SQLiteIndexService {
	return &SQLiteIndexService{
		DB:            db,
		BadgerService: badgerService,
		Connections:   make(map[string]*gorm.DB),
	}
}

func (s *SQLiteIndexService) EnsureIndexes() error {
	// Lấy danh sách các index từ catalog
	var indexes []models.Index
	if err := s.DB.Where("server_id = ?", "localhost").Find(&indexes).Error; err != nil {
		return fmt.Errorf("failed to get indexes from catalog: %v", err)
	}

	for _, index := range indexes {
		// Đảm bảo file SQLite cho collection đã được tạo hoặc tồn tại
		dbPath := filepath.Join("data", "index", fmt.Sprintf("%d.db", *index.CollectionID))
		if err := s.ensureDBFile(dbPath); err != nil {
			return fmt.Errorf("failed to ensure db file for collection %d: %v", *index.CollectionID, err)
		}

		// Khởi tạo kết nối cơ sở dữ liệu bằng GORM
		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}

		// Đảm bảo table trùng tên với index đã được tạo hoặc tồn tại
		if err := s.ensureTable(db, index); err != nil {
			return fmt.Errorf("failed to ensure table for index %s: %v", index.Name, err)
		}

		// Lưu kết nối vào mảng Connections
		s.Connections[dbPath] = db
	}

	return nil
}

func (s *SQLiteIndexService) ensureDBFile(dbPath string) error {
	// Kiểm tra nếu file đã tồn tại
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// Tạo thư mục nếu chưa tồn tại
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for db file: %v", err)
		}

		// Tạo file SQLite
		file, err := os.Create(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create db file: %v", err)
		}
		file.Close()
	}

	return nil
}

func (s *SQLiteIndexService) ensureTable(db *gorm.DB, index models.Index) error {
	// Tạo table sử dụng model.IndexTableStruct với tên table là index.Name
	if err := db.Table(index.Name).AutoMigrate(&models.IndexTableStruct{}); err != nil {
		return fmt.Errorf("failed to auto migrate table: %v", err)
	}
	return nil
}

// Thêm hàm tạo mới một bảng index trong một collection
func (s *SQLiteIndexService) CreateIndex(collectionID int, indexName string, fields string, isUnique bool) error {
	// Tạo một đối tượng Index mới
	index := models.Index{
		CollectionID: &collectionID,
		Name:         indexName,
		Fields:       fields,
		IsUnique:     isUnique,
	}

	// Lưu index vào catalog
	if err := s.DB.Create(&index).Error; err != nil {
		return fmt.Errorf("failed to create index in catalog: %v", err)
	}

	// Đảm bảo file SQLite cho collection đã được tạo hoặc tồn tại
	dbPath := filepath.Join("data", "index", fmt.Sprintf("%d.db", collectionID))
	if err := s.ensureDBFile(dbPath); err != nil {
		return fmt.Errorf("failed to ensure db file for collection %d: %v", collectionID, err)
	}

	// Khởi tạo kết nối cơ sở dữ liệu bằng GORM
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Tạo table mới, tên table là indexName
	if err := db.Table(index.Name).AutoMigrate(&models.IndexTableStruct{}); err != nil {
		return fmt.Errorf("failed to auto migrate table: %v", err)
	}

	// Lưu kết nối vào mảng Connections
	s.Connections[dbPath] = db

	return nil
}

// Trả về kết nối đến SQLite database của một collection
func (s *SQLiteIndexService) GetConnection(collectionID int) (*gorm.DB, error) {
	dbPath := filepath.Join("data", "index", fmt.Sprintf("%d.db", collectionID))
	db, ok := s.Connections[dbPath]
	if !ok {
		return nil, fmt.Errorf("connection for collection %d not found", collectionID)
	}
	return db, nil
}

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
	SqliteCatalogService *SQLiteCatalogService
	BadgerService        *BadgerService
	Connections          map[string]*gorm.DB
}

func NewSQLiteIndexService(sqliteCatalogService *SQLiteCatalogService, badgerService *BadgerService) *SQLiteIndexService {
	return &SQLiteIndexService{
		SqliteCatalogService: sqliteCatalogService,
		BadgerService:        badgerService,
		Connections:          make(map[string]*gorm.DB),
	}
}

func (s *SQLiteIndexService) EnsureIndexes() error {
	// Lấy danh sách các index từ catalog
	var indexes []models.Index
	if err := s.SqliteCatalogService.db.Where("server_id = ?", "localhost").Find(&indexes).Error; err != nil {
		return fmt.Errorf("failed to get indexes from catalog: %v", err)
	}

	for _, index := range indexes {
		// retrieve collection
		collection, err := s.getCollection(*index.CollectionID)
		if err != nil {
			return fmt.Errorf("failed to get collection: %v", err)
		}

		// Đảm bảo file SQLite cho collection đã được tạo hoặc tồn tại
		dbPath := filepath.Join("data", "index", fmt.Sprintf("%d.db", &collection.Name))
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
		s.Connections[collection.Name] = db
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
	// Tạo schema với datatype chỉ định
	schema := models.IndexTableStruct{}
	if index.DataType == models.DataTypeInt {
		if value, ok := schema.Value.(int); ok {
			schema.Value = value
		} else {
			return fmt.Errorf("failed to assert schema value to int")
		}
	} else if index.DataType == models.DataTypeFloat {
		if value, ok := schema.Value.(float64); ok {
			schema.Value = value
		} else {
			return fmt.Errorf("failed to assert schema value to float64")
		}
	} else {
		if value, ok := schema.Value.(string); ok {
			schema.Value = value
		} else {
			return fmt.Errorf("failed to assert schema value to string")
		}
	}

	// Tạo table sử dụng model.IndexTableStruct với tên table là index.Name
	if err := db.Table(index.Name).AutoMigrate(schema); err != nil {
		return fmt.Errorf("failed to auto migrate table: %v", err)
	}
	return nil
}

// Thêm hàm tạo mới một bảng index trong một collection
func (s *SQLiteIndexService) CreateIndex(collectionID int, index *models.Index) error {
	// Lưu index vào catalog
	if err := s.SqliteCatalogService.db.Create(&index).Error; err != nil {
		return fmt.Errorf("failed to create index in catalog: %v", err)
	}

	// retrieve collection
	collection, err := s.getCollection(collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %v", err)
	}

	// Đảm bảo file SQLite cho collection đã được tạo hoặc tồn tại
	dbPath := filepath.Join("data", "index", fmt.Sprintf("%s.db", collection.Name))
	if err := s.ensureDBFile(dbPath); err != nil {
		return fmt.Errorf("failed to ensure db file for collection %d: %v", collectionID, err)
	}

	// Khởi tạo kết nối cơ sở dữ liệu bằng GORM
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Tạo schema với datatype chỉ định
	schema := models.IndexTableStruct{}
	if index.DataType == models.DataTypeInt {
		if value, ok := schema.Value.(int); ok {
			schema.Value = value
		} else {
			return fmt.Errorf("failed to assert schema value to int")
		}
	} else if index.DataType == models.DataTypeFloat {
		if value, ok := schema.Value.(float64); ok {
			schema.Value = value
		} else {
			return fmt.Errorf("failed to assert schema value to float64")
		}
	} else {
		if value, ok := schema.Value.(string); ok {
			schema.Value = value
		} else {
			return fmt.Errorf("failed to assert schema value to string")
		}
	}
	// Tạo table mới, tên table là indexName
	if err := db.Table(index.Name).AutoMigrate(schema); err != nil {
		return fmt.Errorf("failed to auto migrate table: %v", err)
	}

	// Lưu kết nối vào mảng Connections
	s.Connections[collection.Name] = db

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

func (s *SQLiteIndexService) getCollection(collectionID int) (*models.Collection, error) {
	var collection models.Collection
	if err := s.SqliteCatalogService.db.First(&collection, collectionID).Error; err != nil {
		return nil, fmt.Errorf("failed to get collection: %v", err)
	}
	return &collection, nil
}

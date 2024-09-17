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

	sqliteIndexService := SQLiteIndexService{
		SqliteCatalogService: sqliteCatalogService,
		BadgerService:        badgerService,
		Connections:          make(map[string]*gorm.DB),
	}
	err := sqliteIndexService.EnsureIndexes()
	if err != nil {
		log.Fatalf("failed to ensure db file: %v", err)
	}
	return &sqliteIndexService
}

func (s *SQLiteIndexService) EnsureIndexes() error {
	// Lấy danh sách các index từ catalog thuộc main server
	var indexes []models.Index
	if err := s.SqliteCatalogService.Db.Where("server_id = ?", 1).Find(&indexes).Error; err != nil {
		return fmt.Errorf("failed to get indexes from catalog: %v", err)
	}

	for _, index := range indexes {
		// retrieve collection
		collection, err := s.getCollection(index.CollectionID)
		if err != nil {
			return fmt.Errorf("failed to get collection: %v", err)
		}

		// Đảm bảo file SQLite cho collection đã được tạo hoặc tồn tại
		dbPath := filepath.Join("data", "index", fmt.Sprintf("%d.db", &collection.Name))
		if err := s.ensureDBFile(dbPath); err != nil {
			return fmt.Errorf("failed to ensure db file for collection %d: %v", index.CollectionID, err)
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
	// Tạo table chưa index data
	err := createIndexTable(db, index.Name, index.DataType)
	if err != nil {
		return fmt.Errorf("failed to create index table: %v", err)
	}
	return nil
}

// TẠo index trong index service, hàm này phải được gọi sau khi index đã được tạo trong catalog
func (s *SQLiteIndexService) CreateIndex(index *models.Index) error {
	// retrieve collection
	collection, err := s.getCollection(index.CollectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %v", err)
	}

	// Đảm bảo file SQLite cho collection đã được tạo hoặc tồn tại
	dbPath := filepath.Join("data", "index", fmt.Sprintf("%s.db", collection.Name))
	if err := s.ensureDBFile(dbPath); err != nil {
		return fmt.Errorf("failed to ensure db file for collection %d: %v", index.CollectionID, err)
	}

	// Khởi tạo kết nối cơ sở dữ liệu bằng GORM
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Tạo schema với datatype chỉ định
	// Tạo table mới, tên table là indexName
	if err := createIndexTable(db, index.Name, index.DataType); err != nil {
		return fmt.Errorf("failed to create index table: %v", err)
	}

	// Lưu kết nối vào mảng Connections
	s.Connections[collection.Name] = db

	return nil
}

// Trả về kết nối đến SQLite database của một collection
func (s *SQLiteIndexService) GetConnection(collectionName string) (*gorm.DB, error) {
	db, ok := s.Connections[collectionName]
	if !ok {
		return nil, fmt.Errorf("connection for collection %s not found", collectionName)
	}
	return db, nil
}

func (s *SQLiteIndexService) getCollection(collectionID int) (*models.Collection, error) {
	var collection models.Collection
	if err := s.SqliteCatalogService.Db.First(&collection, collectionID).Error; err != nil {
		return nil, fmt.Errorf("failed to get collection: %v", err)
	}
	return &collection, nil
}

// Tìm tất cả các keys trong table default
func (s *SQLiteIndexService) FindKeys(collectionName string) ([]string, error) {
	db, err := s.GetConnection(collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %v", err)
	}

	var keys []string
	if err := db.Table(fmt.Sprintf("%s_default_idx", collectionName)).Pluck("keys", &keys).Error; err != nil {
		return nil, fmt.Errorf("failed to find keys: %v", err)
	}
	return keys, nil
}

// Hàm tạo bảng Index với tên và kiểu dữ liệu tùy biến
func createIndexTable(db *gorm.DB, tableName string, valueType string) error {
	// Chuyển đổi kiểu dữ liệu đầu vào thành kiểu dữ liệu SQLite
	var sqliteType string
	switch valueType {
	case models.DataTypeString:
		sqliteType = "TEXT"
	case models.DataTypeInt:
		sqliteType = "INTEGER"
	case models.DataTypeFloat:
		sqliteType = "REAL"
	default:
		return fmt.Errorf("kiểu dữ liệu không hợp lệ: %s, chỉ chấp nhận string, int, hoặc float", valueType)
	}

	// Câu lệnh SQL để tạo bảng động
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			value %s,
			keys TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
	`, tableName, sqliteType)

	// Thực thi câu lệnh SQL để tạo bảng
	if err := db.Exec(createTableSQL).Error; err != nil {
		return fmt.Errorf("không thể tạo bảng: %v", err)
	}

	log.Printf("Bảng '%s' đã được tạo thành công với kiểu dữ liệu '%s' cho trường 'value'.\n", tableName, valueType)
	return nil
}

func (s *SQLiteIndexService) GetConnectionList() map[string]*gorm.DB {
	return s.Connections
}

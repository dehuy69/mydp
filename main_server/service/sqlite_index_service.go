package service

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dehuy69/mydp/main_server/models"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

type SQLiteIndexService struct {
	DB            *gorm.DB
	BadgerService *BadgerService
	Connections   map[string]*sql.DB
}

func NewSQLiteIndexService(db *gorm.DB, badgerService *BadgerService) *SQLiteIndexService {
	return &SQLiteIndexService{
		DB:            db,
		BadgerService: badgerService,
		Connections:   make(map[string]*sql.DB),
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

		// Mở kết nối đến SQLite database
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			return fmt.Errorf("failed to open sqlite db for collection %d: %v", *index.CollectionID, err)
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

func (s *SQLiteIndexService) ensureTable(db *sql.DB, index models.Index) error {
	// Tạo câu lệnh SQL để tạo table nếu chưa tồn tại
	fields := strings.Split(index.Fields, ",")
	fieldDefinitions := make([]string, len(fields))
	for i, field := range fields {
		fieldDefinitions[i] = fmt.Sprintf("%s TEXT", field)
	}
	// Thêm cột key để xác định vị trí data trong Badger
	fieldDefinitions = append(fieldDefinitions, "key TEXT")
	fieldDefinitionsSQL := strings.Join(fieldDefinitions, ", ")
	primaryKeySQL := strings.Join(fields, ", ")

	uniqueConstraint := ""
	if index.IsUnique {
		uniqueConstraint = fmt.Sprintf(", UNIQUE (%s)", primaryKeySQL)
	}

	createTableSQL := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            %s,
            PRIMARY KEY (%s)
            %s
        );
    `, index.Name, fieldDefinitionsSQL, primaryKeySQL, uniqueConstraint)

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %v", index.Name, err)
	}

	return nil
}

func (s *SQLiteIndexService) CloseConnections() {
	for _, db := range s.Connections {
		db.Close()
	}
}

// Thêm hàm tạo mới một index trong một collection
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

	// Mở kết nối đến SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open sqlite db for collection %d: %v", collectionID, err)
	}

	// Đảm bảo table trùng tên với index đã được tạo hoặc tồn tại
	if err := s.ensureTable(db, index); err != nil {
		return fmt.Errorf("failed to ensure table for index %s: %v", indexName, err)
	}

	// Lưu kết nối vào mảng Connections
	s.Connections[dbPath] = db

	return nil
}

// Trả về kết nối đến SQLite database của một collection
func (s *SQLiteIndexService) GetConnection(collectionID int) (*sql.DB, error) {
	dbPath := filepath.Join("data", "index", fmt.Sprintf("%d.db", collectionID))
	db, ok := s.Connections[dbPath]
	if !ok {
		return nil, fmt.Errorf("connection for collection %d not found", collectionID)
	}
	return db, nil
}

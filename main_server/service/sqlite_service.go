package service

import (
	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/main_server/models"
	"github.com/dehuy69/mydp/utils"

	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteService struct {
	db *gorm.DB
}

func NewSQLiteService(cfg *config.Config) (*SQLiteService, error) {
	// Path to SQLite database file
	path := filepath.Join(cfg.DataFolderDefault, "catalog", "sqlite3.db")

	// Create the folders if not exist
	utils.EnsureFolder(filepath.Dir(path))

	// Mở kết nối tới SQLite bằng GORM
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Tự động migrate các bảng
	err = autoMigrate(db)
	if err != nil {
		return nil, err
	}

	// Khởi tạo SQLiteManager
	manager := &SQLiteService{db: db}

	// Tạo người dùng admin mặc định nếu chưa tồn tại
	err = manager.createDefaultAdminUser()
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// autoMigrate tự động migrate các bảng trong models/catalog.go
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Workspace{},
		&models.Collection{},
		&models.Table{},
		&models.Index{},
		&models.Pipeline{},
		&models.User{}, // Thêm bảng người dùng
	)
}

// createDefaultAdminUser tạo người dùng admin mặc định nếu chưa tồn tại
func (m *SQLiteService) createDefaultAdminUser() error {
	var user models.User
	result := m.db.First(&user, "username = ?", "admin")
	if result.Error == gorm.ErrRecordNotFound {
		hashedPassword, _ := utils.HashPassword("admin_password")
		adminUser := models.User{
			Username: "admin",
			Password: hashedPassword, // Bạn nên mã hóa mật khẩu này
			Role:     "admin",
		}
		return m.db.Create(&adminUser).Error
	}
	return result.Error
}

// Close đóng kết nối cơ sở dữ liệu
func (m *SQLiteService) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AddUser thêm người dùng mới với mật khẩu đã mã hóa
func (m *SQLiteService) AddUser(username, password, role string) error {
	user := models.User{
		Username: username,
		Password: password, // Bạn nên mã hóa mật khẩu này
		Role:     role,
	}
	return m.db.Create(&user).Error
}

// GetUser lấy thông tin người dùng từ cơ sở dữ liệu
func (m *SQLiteService) GetUser(username string) (*models.User, error) {
	var user models.User
	result := m.db.First(&user, "username = ?", username)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// CreateCollection tạo một collection mới
func (m *SQLiteService) CreateCollection(collection *models.Collection) error {
	return m.db.Create(collection).Error
}

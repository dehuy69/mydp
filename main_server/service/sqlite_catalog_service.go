package service

import (
	"reflect"

	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/main_server/models"
	"github.com/dehuy69/mydp/utils"

	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteCatalogService struct {
	Db *gorm.DB
}

func NewSQLiteCatalogService(cfg *config.Config) (*SQLiteCatalogService, error) {
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
	manager := &SQLiteCatalogService{Db: db}

	// Tạo người dùng admin mặc định nếu chưa tồn tại
	err = manager.createDefaultAdminUser()
	if err != nil {
		return nil, err
	}

	// Tạo server mặc định nếu chưa tồn tại
	manager.createDefaultServer()
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
		&models.User{},   // Thêm bảng người dùng
		&models.Server{}, // Thêm bảng server
		&models.Shard{},  // Thêm bảng shard
	)
}

// createDefaultServer tạo server mặc định nếu chưa tồn tại
func (m *SQLiteCatalogService) createDefaultServer() error {
	var server models.Server
	result := m.Db.First(&server, "host = ?", "localhost")
	if result.Error == gorm.ErrRecordNotFound {
		server := models.Server{
			Host: "localhost",
		}
		return m.Db.Create(&server).Error
	}
	return result.Error
}

// createDefaultAdminUser tạo người dùng admin mặc định nếu chưa tồn tại
func (m *SQLiteCatalogService) createDefaultAdminUser() error {
	var user models.User
	result := m.Db.First(&user, "username = ?", "admin")
	if result.Error == gorm.ErrRecordNotFound {
		hashedPassword, _ := utils.HashPassword("admin_password")
		adminUser := models.User{
			Username: "admin",
			Password: hashedPassword, // Bạn nên mã hóa mật khẩu này
			Role:     "admin",
		}
		return m.Db.Create(&adminUser).Error
	}
	return result.Error
}

// Close đóng kết nối cơ sở dữ liệu
func (m *SQLiteCatalogService) Close() error {
	sqlDB, err := m.Db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AddUser thêm người dùng mới với mật khẩu đã mã hóa
func (m *SQLiteCatalogService) AddUser(username, password, role string) error {
	user := models.User{
		Username: username,
		Password: password, // Bạn nên mã hóa mật khẩu này
		Role:     role,
	}
	return m.Db.Create(&user).Error
}

// GetUser lấy thông tin người dùng từ cơ sở dữ liệu
func (m *SQLiteCatalogService) GetUser(username string) (*models.User, error) {
	var user models.User
	result := m.Db.First(&user, "username = ?", username)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// CreateCollection tạo một collection mới
func (m *SQLiteCatalogService) CreateCollection(collection *models.Collection) error {
	err := m.Db.Create(collection).Error
	if err != nil {
		return err
	}
	// Get association
	err = m.GetModelWithAllAssociations(collection, collection.ID)
	if err != nil {
		return err
	}
	return nil
}

func (m *SQLiteCatalogService) GetCollectionByName(name string) (*models.Collection, error) {
	var collection models.Collection
	result := m.Db.First(&collection, "name = ?", name)
	if result.Error != nil {
		return nil, result.Error
	}
	return &collection, nil
}

func (m *SQLiteCatalogService) GetCollectionByID(id int) (*models.Collection, error) {
	var collection models.Collection
	result := m.Db.First(&collection, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &collection, nil
}

func (m *SQLiteCatalogService) GetIndexesByServer(server string) ([]models.Index, error) {
	var indexes []models.Index
	result := m.Db.Find(&indexes, "server = ?", server)
	if result.Error != nil {
		return nil, result.Error
	}
	return indexes, nil
}

func (m *SQLiteCatalogService) CreateIndex(index *models.Index) error {
	return m.Db.Create(index).Error
}

// Tạo shard mới
func (m *SQLiteCatalogService) CreateShard(shard *models.Shard) error {
	return m.Db.Create(shard).Error
}

// Lấy Server theo host
func (m *SQLiteCatalogService) GetServerByHost(host string) (*models.Server, error) {
	var server models.Server
	result := m.Db.First(&server, "host = ?", host)
	if result.Error != nil {
		return nil, result.Error
	}
	return &server, nil
}

// Lấy workspace theo tên
func (m *SQLiteCatalogService) GetWorkspaceByName(name string) (*models.Workspace, error) {
	var workspace models.Workspace
	result := m.Db.First(&workspace, "name = ?", name)
	if result.Error != nil {
		return nil, result.Error
	}
	return &workspace, nil
}

func (m *SQLiteCatalogService) GetWorkspaceByID(id int) (*models.Workspace, error) {
	var workspace models.Workspace
	result := m.Db.First(&workspace, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &workspace, nil
}

// UpdateIndex cập nhật thông tin index
func (m *SQLiteCatalogService) UpdateIndex(index *models.Index) error {
	return m.Db.Save(index).Error
}

// CreatWorkspace
func (m *SQLiteCatalogService) CreateWorkspace(workspace *models.Workspace) error {
	return m.Db.Create(workspace).Error
}

// Hàm chung để preload tất cả các quan hệ của một model đầu vào
func (m *SQLiteCatalogService) GetModelWithAllAssociations(model interface{}, id int) error {
	query := m.Db.Model(model)

	// Sử dụng reflection để preload tất cả các liên kết
	val := reflect.ValueOf(model).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if field.Tag.Get("gorm") != "" && (field.Type.Kind() == reflect.Ptr || field.Type.Kind() == reflect.Slice) {
			associationName := field.Name
			query = query.Preload(associationName)
		}
	}

	// Truy vấn với preload tất cả các liên kết
	if err := query.First(model, id).Error; err != nil {
		return err
	}
	return nil
}

func (m *SQLiteCatalogService) GetIndexByID(id int) (*models.Index, error) {
	var index models.Index
	result := m.Db.First(&index, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &index, nil
}

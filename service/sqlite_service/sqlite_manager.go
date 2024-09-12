package sqlite_service

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/models"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// SQLiteManager struct để quản lý SQLite DB cho metadata
type SQLiteManager struct {
	db *sql.DB
}

// NewSQLiteManager khởi tạo một SQLiteManager mới với cấu hình từ config.Config
func NewSQLiteManager(cfg *config.Config) (*SQLiteManager, error) {
	// Tạo thư mục nếu chưa tồn tại
	if err := os.MkdirAll(filepath.Dir(cfg.SQLiteFile), os.ModePerm); err != nil {
		return nil, err
	}

	// Kiểm tra xem file SQLite đã tồn tại chưa
	if _, err := os.Stat(cfg.SQLiteFile); os.IsNotExist(err) {
		// Tạo file SQLite mới
		file, err := os.Create(cfg.SQLiteFile)
		if err != nil {
			return nil, err
		}
		file.Close() // Đóng file sau khi tạo
	}

	// Mở kết nối tới SQLite
	db, err := sql.Open("sqlite3", cfg.SQLiteFile)
	if err != nil {
		return nil, err
	}

	// Tạo bảng người dùng nếu chưa tồn tại
	err = createUserTable(db)
	if err != nil {
		return nil, err
	}

	// Khởi tạo SQLiteManager
	manager := &SQLiteManager{db: db}

	// Tạo người dùng admin mặc định nếu chưa tồn tại
	err = manager.createDefaultAdminUser()
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// createUserTable tạo bảng người dùng trong SQLite
func createUserTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		role TEXT NOT NULL,
		created_at TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}

// AddUser thêm người dùng mới với mật khẩu đã mã hóa
func (m *SQLiteManager) AddUser(username, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = m.db.Exec("INSERT INTO users (username, password, role, created_at) VALUES (?, ?, ?, datetime('now'))", username, string(hashedPassword), role)
	return err
}

// UserExists kiểm tra xem người dùng có tồn tại hay không
func (m *SQLiteManager) UserExists(username string) (bool, error) {
	var exists bool
	err := m.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

// createDefaultAdminUser tạo người dùng admin mặc định nếu chưa tồn tại
func (m *SQLiteManager) createDefaultAdminUser() error {
	exists, err := m.UserExists("admin")
	if err != nil {
		return err
	}

	if !exists {
		// Tạo người dùng admin mặc định
		err := m.AddUser("admin", "admin123", "Admin")
		if err != nil {
			return err
		}
		log.Println("Default admin user created.")
	} else {
		log.Println("Admin user already exists.")
	}
	return nil
}

// GetUser lấy toàn bộ thông tin của người dùng từ cơ sở dữ liệu
func (m *SQLiteManager) GetUser(username string) (*models.User, error) {
	var user models.User
	err := m.db.QueryRow("SELECT id, username, password, role, created_at FROM users WHERE username = ?", username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Close đóng kết nối SQLite
func (m *SQLiteManager) Close() {
	m.db.Close()
}

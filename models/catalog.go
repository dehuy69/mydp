package models

import (
	"gorm.io/gorm"
)

// User struct đại diện cho một người dùng trong hệ thống
type User struct {
	gorm.Model
	ID        int    `json:"id"`         // ID của người dùng
	Username  string `json:"username"`   // Tên đăng nhập của người dùng
	Password  string `json:"password"`   // Mật khẩu đã mã hóa
	Role      string `json:"role"`       // Vai trò của người dùng (Admin, User, etc.)
	CreatedAt string `json:"created_at"` // Thời gian tạo người dùng
}

// UserPermission struct đại diện cho quyền truy cập của người dùng vào một cơ sở dữ liệu hoặc bảng
type UserPermission struct {
	gorm.Model
	ID         int    `json:"id"`          // ID của quyền truy cập
	UserID     int    `json:"user_id"`     // ID của người dùng
	DatabaseID int    `json:"database_id"` // ID của cơ sở dữ liệu
	TableID    int    `json:"table_id"`    // ID của bảng (nếu áp dụng)
	Permission string `json:"permission"`  // Quyền truy cập (READ, WRITE, ADMIN, etc.)
	GrantedAt  string `json:"granted_at"`  // Thời gian cấp quyền
}

// Workspace struct đại diện cho một workspace trong catalog
type Workspace struct {
	gorm.Model
	ID        int    `json:"id"`         // ID của workspace
	Name      string `json:"name"`       // Tên của workspace
	Owner     string `json:"owner"`      // Chủ sở hữu của workspace
	CreatedAt string `json:"created_at"` // Thời gian tạo workspace
}

// Collection struct đại diện cho một bảng OLTP trong workspace
type Collection struct {
	gorm.Model
	ID          int    `json:"id"`           // ID của collection
	Name        string `json:"name"`         // Tên của collection
	WorkspaceID int    `json:"workspace_id"` // ID của workspace chứa collection này
	CreatedAt   string `json:"created_at"`   // Thời gian tạo collection
}

// Table struct đại diện cho một bảng OLAP trong workspace
type Table struct {
	gorm.Model
	ID          int    `json:"id"`           // ID của table
	Name        string `json:"name"`         // Tên của table
	WorkspaceID int    `json:"workspace_id"` // ID của workspace chứa table này
	CreatedAt   string `json:"created_at"`   // Thời gian tạo table
}

// Index struct đại diện cho một chỉ mục trong một collection hoặc table
type Index struct {
	gorm.Model
	ID        int    `json:"id"`         // ID của chỉ mục
	Name      string `json:"name"`       // Tên của chỉ mục
	TableID   int    `json:"table_id"`   // ID của collection hoặc table chứa chỉ mục này
	IndexType string `json:"index_type"` // Loại chỉ mục (ví dụ: B-Tree, Hash, Inverted Index)
	CreatedAt string `json:"created_at"` // Thời gian tạo chỉ mục
}

// Pipeline struct đại diện cho một pipeline trong workspace
type Pipeline struct {
	gorm.Model
	ID          int    `json:"id"`           // ID của pipeline
	Name        string `json:"name"`         // Tên của pipeline
	WorkspaceID int    `json:"workspace_id"` // ID của workspace chứa pipeline này
	Notebook    string `json:"notebook"`     // Đường dẫn tới Jupyter Notebook
	Schedule    string `json:"schedule"`     // Lịch trình chạy pipeline
	CreatedAt   string `json:"created_at"`   // Thời gian tạo pipeline
}

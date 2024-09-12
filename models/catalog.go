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
	Name        string       `json:"name" gorm:"unique;not null"` // Tên của workspace
	OwnerID     int          `json:"owner_id" gorm:"not null"`    // ID của chủ sở hữu workspace
	Owner       User         `gorm:"foreignKey:OwnerID"`          // Tham chiếu đến người dùng chủ sở hữu
	Collections []Collection `json:"collections"`                 // Danh sách các collections trong workspace
	Tables      []Table      `json:"tables"`                      // Danh sách các tables trong workspace
	Pipelines   []Pipeline   `json:"pipelines"`                   // Danh sách các pipelines trong workspace
}

// Collection struct đại diện cho một bảng OLTP trong workspace
type Collection struct {
	gorm.Model
	Name        string    `json:"name" gorm:"unique;not null"`  // Tên của collection
	WorkspaceID int       `json:"workspace_id" gorm:"not null"` // ID của workspace chứa collection này
	Workspace   Workspace `gorm:"foreignKey:WorkspaceID"`       // Tham chiếu đến workspace
	Indexes     []Index   `json:"indexes"`                      // Danh sách các chỉ mục trong collection
}

// Table struct đại diện cho một bảng OLAP trong workspace
type Table struct {
	gorm.Model
	Name        string    `json:"name" gorm:"unique;not null"`  // Tên của table
	WorkspaceID int       `json:"workspace_id" gorm:"not null"` // ID của workspace chứa table này
	Workspace   Workspace `gorm:"foreignKey:WorkspaceID"`       // Tham chiếu đến workspace
	Indexes     []Index   `json:"indexes"`                      // Danh sách các chỉ mục trong table
}

// Index struct đại diện cho một chỉ mục trong một collection hoặc table
type Index struct {
	gorm.Model
	Name         string      `json:"name" gorm:"not null"`       // Tên của chỉ mục
	TableID      *int        `json:"table_id"`                   // ID của table chứa chỉ mục này (nếu có)
	CollectionID *int        `json:"collection_id"`              // ID của collection chứa chỉ mục này (nếu có)
	IndexType    string      `json:"index_type" gorm:"not null"` // Loại chỉ mục (ví dụ: B-Tree, Hash, Inverted Index)
	Table        *Table      `gorm:"foreignKey:TableID"`         // Tham chiếu đến bảng
	Collection   *Collection `gorm:"foreignKey:CollectionID"`    // Tham chiếu đến collection
}

// Pipeline struct đại diện cho một pipeline trong workspace
type Pipeline struct {
	gorm.Model
	Name        string    `json:"name" gorm:"unique;not null"`  // Tên của pipeline
	WorkspaceID int       `json:"workspace_id" gorm:"not null"` // ID của workspace chứa pipeline này
	Workspace   Workspace `gorm:"foreignKey:WorkspaceID"`       // Tham chiếu đến workspace
	Notebook    string    `json:"notebook" gorm:"not null"`     // Đường dẫn tới Jupyter Notebook
	Schedule    string    `json:"schedule"`                     // Lịch trình chạy pipeline
}

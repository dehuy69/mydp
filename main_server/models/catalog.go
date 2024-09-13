package models

import (
	"gorm.io/gorm"
)

// User struct đại diện cho một người dùng trong hệ thống
type User struct {
	gorm.Model
	Username        string           `json:"username" gorm:"unique;not null"` // Tên đăng nhập của người dùng
	Password        string           `json:"password" gorm:"not null"`        // Mật khẩu đã mã hóa
	Role            string           `json:"role" gorm:"not null"`            // Vai trò của người dùng (Admin, User, etc.)
	UserPermissions []UserPermission `json:"permissions"`                     // Danh sách quyền của người dùng
}

// UserPermission struct đại diện cho quyền truy cập của người dùng vào một cơ sở dữ liệu hoặc bảng
type UserPermission struct {
	gorm.Model
	UserID      int    `json:"user_id" gorm:"not null"`      // ID của người dùng
	WorkspaceID int    `json:"workspace_id" gorm:"not null"` // ID của workspace
	Permission  string `json:"permission" gorm:"not null"`   // Quyền truy cập (READ, WRITE, ADMIN, etc.)
	User        User   `gorm:"foreignKey:UserID"`            // Tham chiếu đến người dùng
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

// Collection struct đại diện cho một bảng OLTP trong workspace với hỗ trợ sharding
type Collection struct {
	gorm.Model
	Name          string    `json:"name" gorm:"unique;not null"`    // Tên của collection
	WorkspaceID   int       `json:"workspace_id" gorm:"not null"`   // ID của workspace chứa collection này
	Workspace     Workspace `gorm:"foreignKey:WorkspaceID"`         // Tham chiếu đến workspace
	ShardKey      string    `json:"shard_key" gorm:"not null"`      // Khóa để thực hiện sharding
	ShardStrategy string    `json:"shard_strategy" gorm:"not null"` // Chiến lược sharding (range, hash, list, etc.)
	Shards        []Shard   `json:"shards"`                         // Danh sách các shards
	Indexes       []Index   `json:"indexes"`                        // Danh sách các chỉ mục trong collection
}

// Server struct đại diện cho một máy chủ nơi các shards được triển khai
type Server struct {
	gorm.Model
	Host        string  `json:"host" gorm:"unique;not null"`  // Địa chỉ host hoặc IP của server
	IsLocalhost bool    `json:"is_localhost" gorm:"not null"` // Đánh dấu nếu đây là localhost
	Shards      []Shard `json:"shards"`                       // Danh sách các shards được triển khai trên server này
}

// Shard struct đại diện cho thông tin về một shard trong Collection
type Shard struct {
	gorm.Model
	CollectionID int    `json:"collection_id" gorm:"not null"` // ID của collection chứa shard này
	ShardNumber  int    `json:"shard_number" gorm:"not null"`  // Số thứ tự của shard trong collection
	ShardKey     string `json:"shard_key" gorm:"not null"`     // Giá trị của shard key
	ServerID     int    `json:"server_id" gorm:"not null"`     // ID của server chứa shard này
	Server       Server `gorm:"foreignKey:ServerID"`           // Tham chiếu đến server
	Status       string `json:"status" gorm:"not null"`        // Trạng thái của shard (active, inactive, etc.)
	Size         int64  `json:"size"`                          // Kích thước của shard
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

package models

// Database struct đại diện cho một cơ sở dữ liệu trong catalog
type Database struct {
	ID        int    `json:"id"`         // ID của database
	Name      string `json:"name"`       // Tên của database
	Owner     string `json:"owner"`      // Chủ sở hữu của database
	CreatedAt string `json:"created_at"` // Thời gian tạo database
}

// Table struct đại diện cho một bảng trong cơ sở dữ liệu
type Table struct {
	ID         int    `json:"id"`          // ID của bảng
	Name       string `json:"name"`        // Tên của bảng
	DatabaseID int    `json:"database_id"` // ID của database chứa bảng này
	TableType  string `json:"table_type"`  // Loại bảng (OLTP hoặc OLAP)
	CreatedAt  string `json:"created_at"`  // Thời gian tạo bảng
}

// Index struct đại diện cho một chỉ mục trong một bảng
type Index struct {
	ID        int    `json:"id"`         // ID của chỉ mục
	Name      string `json:"name"`       // Tên của chỉ mục
	TableID   int    `json:"table_id"`   // ID của bảng chứa chỉ mục này
	IndexType string `json:"index_type"` // Loại chỉ mục (ví dụ: B-Tree, Hash, Inverted Index)
	CreatedAt string `json:"created_at"` // Thời gian tạo chỉ mục
}

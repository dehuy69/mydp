package models

// User struct đại diện cho một người dùng trong hệ thống
type User struct {
	ID        int    `json:"id"`         // ID của người dùng
	Username  string `json:"username"`   // Tên đăng nhập của người dùng
	Password  string `json:"password"`   // Mật khẩu đã mã hóa
	Role      string `json:"role"`       // Vai trò của người dùng (Admin, User, etc.)
	CreatedAt string `json:"created_at"` // Thời gian tạo người dùng
}

// UserPermission struct đại diện cho quyền truy cập của người dùng vào một cơ sở dữ liệu hoặc bảng
type UserPermission struct {
	ID         int    `json:"id"`          // ID của quyền truy cập
	UserID     int    `json:"user_id"`     // ID của người dùng
	DatabaseID int    `json:"database_id"` // ID của cơ sở dữ liệu
	TableID    int    `json:"table_id"`    // ID của bảng (nếu áp dụng)
	Permission string `json:"permission"`  // Quyền truy cập (READ, WRITE, ADMIN, etc.)
	GrantedAt  string `json:"granted_at"`  // Thời gian cấp quyền
}

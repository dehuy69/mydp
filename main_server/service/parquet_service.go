package service

// ParquetManager quản lý các hoạt động liên quan đến tệp Parquet
type ParquetService struct{}

// NewParquetManager tạo một instance mới của ParquetManager
func NewParquetService() (*ParquetService, error) {
	return &ParquetService{}, nil
}

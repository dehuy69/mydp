package models

// Bảng index dynamic có 2 field cố định là value và keys
type IndexTableStruct struct {
	Value interface{} `json:"value" gorm:"index:idx_value,unique"`
	Keys  string      `json:"keys"`
}

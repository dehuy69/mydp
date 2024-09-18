package models

import "gorm.io/gorm"

type IndexTableStruct struct {
	gorm.Model
	Value interface{} `json:"value" gorm:"index:idx_value,unique"` // Sử dụng kiểu `string` thay vì `interface{}`
	Keys  string      `json:"keys"`
}

// Bảng index dynamic có 2 field cố định là value và keys
type IndexTableStructInt struct {
	gorm.Model
	Value int    `json:"value" gorm:"index:idx_value,unique"`
	Keys  string `json:"keys"`
}

// Bảng index dynamic có 2 field cố định là value và keys
type IndexTableStructString struct {
	gorm.Model
	Value string `json:"value" gorm:"index:idx_value,unique"`
	Keys  string `json:"keys"`
}

// Bảng index dynamic có 2 field cố định là value và keys
type IndexTableStructFloat struct {
	gorm.Model
	Value float64 `json:"value" gorm:"index:idx_value,unique"`
	Keys  string  `json:"keys"`
}

var IndexBboltValue []string

package models

import "gorm.io/gorm"

type Animal struct {
	gorm.Model
	Name        string `json:"name"`
	Type        int    `json:"type"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive" gorm:"default:true"`
}

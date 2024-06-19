package models

import (
	"time"
)

type Animal struct {
	ID          uint `gorm:"primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string `json:"name"`
	Type        int    `json:"type"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive" gorm:"default:true"`
}

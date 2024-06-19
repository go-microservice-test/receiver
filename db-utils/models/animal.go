package models

import (
	"time"
)

type Animal struct {
	ID          uint `gorm:"primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string
	Type        int
	Description string
	IsActive    bool `gorm:"default:true"`
}

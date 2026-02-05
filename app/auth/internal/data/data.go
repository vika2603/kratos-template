package data

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	CreatedAt    int64
	UpdatedAt    int64
}

type Data struct {
	db  *gorm.DB
	log *zap.Logger
}

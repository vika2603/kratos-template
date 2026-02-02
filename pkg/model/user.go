package model

import "time"

type User struct {
	ID           string `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()"`
	Username     string `gorm:"column:username;type:varchar(255);uniqueIndex;not null"`
	Email        string `gorm:"column:email;type:varchar(255);uniqueIndex;not null"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string {
	return "users"
}

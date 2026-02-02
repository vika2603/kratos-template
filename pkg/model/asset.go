package model

import "time"

type Asset struct {
	ID          string  `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name        string  `gorm:"column:name;type:varchar(255);not null"`
	Description string  `gorm:"column:description;type:text"`
	OwnerID     string  `gorm:"column:owner_id;type:uuid;not null;index"`
	Value       float64 `gorm:"column:value;type:decimal(15,2);default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Asset) TableName() string {
	return "assets"
}

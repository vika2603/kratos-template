package data

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Data struct {
	db  *gorm.DB
	log *zap.Logger
}

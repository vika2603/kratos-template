package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"kratos-template/app/user/internal/data/query"
)

type Data struct {
	db  *gorm.DB
	q   *query.Query
	log *log.Helper
}

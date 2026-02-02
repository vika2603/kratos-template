package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gorm"

	"kratos-template/app/asset/internal/data/query"
	"kratos-template/pkg/model"
)

type Data struct {
	db  *gorm.DB
	q   *query.Query
	log *log.Helper
}

func GenerateCode(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./app/asset/internal/data/query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.UseDB(db)
	g.ApplyBasic(model.Asset{})
	g.Execute()
}

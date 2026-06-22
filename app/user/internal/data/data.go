package data

import (
	"cmp"
	"fmt"
	"os"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"kratos-template/app/user/internal/conf"
	"kratos-template/app/user/internal/data/query"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
)

type Data struct {
	db *gorm.DB
	q  *query.Query
}

func NewDB(cfg *conf.Bootstrap, logger *zap.Logger) (*gorm.DB, *query.Query, error) {
	dsn := cmp.Or(os.Getenv("DB_DSN"), cfg.Data.Database.Source)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: adapter.NewGormAdapter(logger),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	query.SetDefault(db)
	q := query.Use(db)

	return db, q, nil
}

func NewData(db *gorm.DB, q *query.Query) (*Data, func(), error) {
	d := &Data{
		db: db,
		q:  q,
	}

	cleanup := func() {
		log.Info("closing data resources")
		if sqlDB, err := d.db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Errorf("failed to close db: %v", err)
			}
		}
	}

	return d, cleanup, nil
}

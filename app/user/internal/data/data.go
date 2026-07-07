package data

import (
	"cmp"
	"fmt"
	"kratos-template/app/user/internal/conf"
	"kratos-template/app/user/internal/data/query"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	if err := configurePool(db, cfg.GetData().GetDatabase()); err != nil {
		return nil, nil, err
	}

	q := query.Use(db)

	return db, q, nil
}

// configurePool applies pool limits; zero config falls back to template
// defaults rather than Go's unlimited open connections.
func configurePool(db *gorm.DB, cfg *conf.Data_Database) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed getting sql.DB for pool config: %w", err)
	}

	maxOpen := int(cfg.GetMaxOpenConns())
	if maxOpen <= 0 {
		maxOpen = 25
	}
	maxIdle := int(cfg.GetMaxIdleConns())
	if maxIdle <= 0 {
		maxIdle = 5
	}
	maxLifetime := cfg.GetConnMaxLifetime().AsDuration()
	if maxLifetime <= 0 {
		// Below common LB/pgbouncer idle timeouts so we recycle first.
		maxLifetime = 30 * time.Minute
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(maxLifetime)
	return nil
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

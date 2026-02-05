package data

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"kratos-template/app/asset/internal/biz"
	"kratos-template/app/asset/internal/conf"
	"kratos-template/app/asset/internal/data/query"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
)

func NewDB(cfg *conf.Bootstrap, logger *zap.Logger) (*gorm.DB, *query.Query, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = cfg.Data.Database.Source
	}

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

func NewData(db *gorm.DB, q *query.Query, logger *zap.Logger) (*Data, func(), error) {
	logger = logger.With(log.String("module", "asset/data"))

	d := &Data{
		db:  db,
		q:   q,
		log: logger,
	}

	cleanup := func() {
		logger.Info("closing data resources")
		if sqlDB, err := d.db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.Sugar().Errorf("failed to close db: %v", err)
			}
		}
	}

	return d, cleanup, nil
}

func NewAssetRepo(data *Data) biz.AssetRepo {
	return &AssetRepo{data: data}
}

func registerLifecycle(lc fx.Lifecycle, cleanup func()) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cleanup()
			return nil
		},
	})
}

var Module = fx.Module("asset.data",
	fx.Provide(NewDB),
	fx.Provide(NewData),
	fx.Provide(NewAssetRepo),
	fx.Invoke(registerLifecycle),
)

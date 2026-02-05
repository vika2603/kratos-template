package data

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"kratos-template/app/auth/internal/biz"
	"kratos-template/app/auth/internal/conf"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
)

func NewDB(cfg *conf.Bootstrap, logger *zap.Logger) (*gorm.DB, error) {
	helper := logger.With(log.String("module", "auth/data/gorm"))

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = cfg.Data.Database.Source
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: adapter.NewGormAdapter(logger),
	})
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		defaultUser := &User{
			Username:     "admin",
			PasswordHash: "$2a$10$64hf5Zc1MygJgdsCF27.zuW3BQrvtf5JvTC1Eei6qW93A7y279a1m",
		}
		if err := db.Create(defaultUser).Error; err != nil {
			helper.Sugar().Errorf("failed to create default user: %v", err)
		}
	}

	return db, nil
}

func NewData(db *gorm.DB, logger *zap.Logger) (*Data, func(), error) {
	helper := logger.With(log.String("module", "auth/data"))

	d := &Data{
		db:  db,
		log: helper,
	}

	cleanup := func() {
		helper.Info("closing data resources")
		if sqlDB, err := d.db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				helper.Sugar().Errorf("failed to close db: %v", err)
			}
		}
	}

	return d, cleanup, nil
}

func NewAuthUserRepo(data *Data) biz.AuthUserRepo {
	return &authUserRepo{data: data}
}

func registerLifecycle(lc fx.Lifecycle, cleanup func()) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cleanup()
			return nil
		},
	})
}

var Module = fx.Module("auth.data",
	fx.Provide(NewDB),
	fx.Provide(NewData),
	fx.Provide(NewAuthUserRepo),
	fx.Invoke(registerLifecycle),
)

package data

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"kratos-template/app/user/internal/biz"
	"kratos-template/app/user/internal/conf"
	"kratos-template/app/user/internal/data/query"
)

func NewDB(cfg *conf.Bootstrap, logger log.Logger) (*gorm.DB, *query.Query, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = cfg.Data.Database.Source
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	query.SetDefault(db)
	q := query.Use(db)

	return db, q, nil
}

func NewData(db *gorm.DB, q *query.Query, logger log.Logger) (*Data, func(), error) {
	helper := log.NewHelper(log.With(logger, "module", "user/data"))

	d := &Data{
		db:  db,
		q:   q,
		log: helper,
	}

	cleanup := func() {
		helper.Info("closing data resources")
		if sqlDB, err := d.db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				helper.Errorf("failed to close db: %v", err)
			}
		}
	}

	return d, cleanup, nil
}

func NewUserRepo(data *Data) biz.UserRepo {
	return &UserRepo{data: data}
}

func registerLifecycle(lc fx.Lifecycle, cleanup func()) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cleanup()
			return nil
		},
	})
}

var Module = fx.Module("user.data",
	fx.Provide(NewDB),
	fx.Provide(NewData),
	fx.Provide(NewUserRepo),
	fx.Invoke(registerLifecycle),
)

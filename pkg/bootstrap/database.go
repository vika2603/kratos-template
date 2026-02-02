package bootstrap

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewGormDB creates a new GORM database connection.
// Auto-migration should be done by each service after calling this function.
func NewGormDB(dsn string, logger log.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to postgres: %w", err)
	}

	return db, nil
}

// RegisterDataCleanup registers a cleanup function for database resources.
func RegisterDataCleanup(lc fx.Lifecycle, cleanup func()) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cleanup()
			return nil
		},
	})
}

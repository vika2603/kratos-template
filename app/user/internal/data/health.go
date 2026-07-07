package data

import (
	"context"
	"kratos-template/pkg/bootstrap"

	"gorm.io/gorm"
)

func NewDBHealthChecker(db *gorm.DB) bootstrap.HealthChecker {
	return bootstrap.HealthChecker{
		Name: "postgres",
		Check: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.PingContext(ctx)
		},
	}
}

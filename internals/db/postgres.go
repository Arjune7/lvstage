package db

import (
	// "lystage-proj/internals/ads"

	"lystage-proj/internals/observability"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var GormDB *gorm.DB

func InitPostgres(dsn string) *gorm.DB {
	var err error
	GormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		observability.Logger.Fatal("Failed to connect to Postgres using GORM", zap.Error(err))
	}

	observability.Logger.Info("Connected to Postgres via GORM")
	// if err := GormDB.AutoMigrate(
	// 	&models.Click{}, &models.Ad{}, &models.AdAnalytics{}, &models.Impression{},
	// ); err != nil {
	// 	observability.Logger.Fatal("AutoMigrate failed", zap.Error(err))
	// }

	return GormDB
}

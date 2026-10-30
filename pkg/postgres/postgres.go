package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func Open(
	ctx context.Context,
	dsn string,
	cfg PoolConfig,
) (*gorm.DB, error) {
	db, err := gorm.Open(
		postgresDriver.Open(dsn),
		&gorm.Config{
			DisableAutomaticPing: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		defer func() {
			if err := sqlDB.Close(); err != nil {
				log.Printf("failed to close db connection: %v", err)
			}
		}()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

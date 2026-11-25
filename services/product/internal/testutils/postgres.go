package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/pressly/goose/v3"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresTestContainer(t *testing.T, migrationsDir string) *gorm.DB {
	t.Helper()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancel()

	container, err := tcpostgres.Run(
		ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("tekora_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// Register container cleanup first
	// Cleanup functions execute in reverse order
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}

	// Close the database before terminating the container
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Logf("failed to close database: %v", err)
		}
	})

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}

	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

// Package database manages the PostgreSQL connection via GORM with
// automatic migration and index creation.
package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
)

// Connect opens a connection to PostgreSQL, runs auto-migration,
// and returns the *gorm.DB instance.
func Connect(dsn string) (*gorm.DB, error) {
	// Use Warn level for runtime: only slow queries (>200 ms) and errors are printed.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	// Connection pool tuning
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Run migrations on a silent session so the hundreds of pg_catalog
	// schema-inspection queries that AutoMigrate emits do not flood stdout.
	if err := autoMigrate(db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})); err != nil {
		return nil, err
	}

	log.Println("[database] connected and migrated")
	return db, nil
}

// autoMigrate runs GORM's AutoMigrate for every domain entity.
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.User{},
		&domain.Session{},
		&domain.Expense{},
		&domain.Income{},
		&domain.Budget{},
		&domain.Subscription{},
		&domain.Goal{},
		&domain.Tag{},
		&domain.Notification{},
		&domain.AuditLog{},
		&domain.RecurringExpense{},
		&domain.MerchantMapping{},
		&domain.FinancialHealthScore{},
	)
}

// Close gracefully closes the underlying sql.DB.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

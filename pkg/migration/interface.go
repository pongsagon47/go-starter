package migration

import (
	"gorm.io/gorm"
)

// Migration interface ที่แต่ละ migration file ต้อง implement
type Migration interface {
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
	Version() string
	Description() string
}

// MigrationEngine interface สำหรับจัดการ migrations
type MigrationEngine interface {
	// Registration
	RegisterMigration(migration Migration)
	GetRegisteredMigrations() []Migration

	// Execution
	RunMigrations() error
	RollbackMigrations(count string) error
	GetMigrationStatus() error

	// Information
	GetAppliedMigrations() ([]MigrationRecord, error)
	GetPendingMigrations() ([]Migration, error)
	IsMigrationApplied(version string) (bool, error)
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	ID          uint   `gorm:"primaryKey"`
	Version     string `gorm:"type:varchar(255);uniqueIndex;not null"`
	Description string `gorm:"type:varchar(500);not null"`
	AppliedAt   string `gorm:"not null"` // Use string for cross-database compatibility
}

// TableName returns the table name for GORM
func (MigrationRecord) TableName() string {
	return "migrations"
}

// MigrationConfig configuration for migration engine
type MigrationConfig struct {
	TableName string // Custom migration table name (default: "migrations")
	AutoRun   bool   // Auto run migrations on startup
}

// DefaultMigrationConfig returns default configuration
func DefaultMigrationConfig() *MigrationConfig {
	return &MigrationConfig{
		TableName: "migrations",
		AutoRun:   false,
	}
}

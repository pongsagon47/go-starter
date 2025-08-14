package migration

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"go-starter/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Manager implements MigrationEngine interface
type Manager struct {
	db         *gorm.DB
	config     *MigrationConfig
	migrations map[string]Migration
}

// NewManager creates a new migration manager
func NewManager(db *gorm.DB, config *MigrationConfig) *Manager {
	if config == nil {
		config = DefaultMigrationConfig()
	}

	return &Manager{
		db:         db,
		config:     config,
		migrations: make(map[string]Migration),
	}
}

// RegisterMigration registers a migration
func (m *Manager) RegisterMigration(migration Migration) {
	m.migrations[migration.Version()] = migration
}

// GetRegisteredMigrations returns all registered migrations
func (m *Manager) GetRegisteredMigrations() []Migration {
	migrations := make([]Migration, 0, len(m.migrations))
	for _, migration := range m.migrations {
		migrations = append(migrations, migration)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version() < migrations[j].Version()
	})

	return migrations
}

// RunMigrations runs all pending migrations
func (m *Manager) RunMigrations() error {
	// Create migrations table if not exists
	if err := m.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedRecords, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, record := range appliedRecords {
		appliedMap[record.Version] = true
	}

	// Get pending migrations
	pendingMigrations, err := m.GetPendingMigrations()
	if err != nil {
		return fmt.Errorf("failed to get pending migrations: %w", err)
	}

	if len(pendingMigrations) == 0 {
		logger.Info("No pending migrations found")
		return nil
	}

	// Run pending migrations
	for _, migration := range pendingMigrations {
		logger.Info("Running migration",
			zap.String("version", migration.Version()),
			zap.String("description", migration.Description()))

		if err := m.runSingleMigration(migration); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Version(), err)
		}

		logger.Info("Migration completed",
			zap.String("version", migration.Version()))
	}

	logger.Info("All migrations completed successfully",
		zap.Int("count", len(pendingMigrations)))

	return nil
}

// RollbackMigrations rolls back specified number of migrations
func (m *Manager) RollbackMigrations(count string) error {
	countInt, err := strconv.Atoi(count)
	if err != nil && count != "all" {
		return fmt.Errorf("invalid count value: %w", err)
	}

	if countInt <= 0 && count != "all" {
		return fmt.Errorf("rollback count must be greater than 0")
	}

	// Get applied migrations in reverse order
	var appliedRecords []MigrationRecord
	query := m.db.Order("applied_at DESC")
	if countInt > 0 {
		query = query.Limit(countInt)
	}
	if err := query.Find(&appliedRecords).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(appliedRecords) == 0 {
		logger.Info("No migrations to rollback")
		return nil
	}

	if len(appliedRecords) < countInt {
		logger.Warn("Only found migrations to rollback",
			zap.Int("requested", countInt),
			zap.Int("available", len(appliedRecords)))
	}

	// Rollback each migration
	for _, record := range appliedRecords {
		migration, exists := m.migrations[record.Version]
		if !exists {
			return fmt.Errorf("migration %s not found in registered migrations", record.Version)
		}

		logger.Info("Rolling back migration",
			zap.String("version", record.Version),
			zap.String("description", record.Description))

		if err := m.rollbackSingleMigration(migration, record); err != nil {
			return fmt.Errorf("rollback failed for migration %s: %w", record.Version, err)
		}

		logger.Info("Migration rolled back successfully",
			zap.String("version", record.Version))
	}

	logger.Info("Rollback completed successfully",
		zap.Int("count", len(appliedRecords)))
	return nil
}

// GetMigrationStatus shows migration status
func (m *Manager) GetMigrationStatus() error {
	// Create migrations table if not exists
	if err := m.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedRecords, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedMap := make(map[string]MigrationRecord)
	for _, record := range appliedRecords {
		appliedMap[record.Version] = record
	}

	// Sort all migrations by version
	allMigrations := m.GetRegisteredMigrations()
	appliedCount := 0
	pendingCount := 0

	logger.Info("Migration Status:")
	logger.Info("================")

	for _, migration := range allMigrations {
		if record, applied := appliedMap[migration.Version()]; applied {
			appliedCount++
			logger.Info("✅ APPLIED",
				zap.String("version", migration.Version()),
				zap.String("description", migration.Description()),
				zap.String("applied_at", record.AppliedAt))
		} else {
			pendingCount++
			logger.Info("⏳ PENDING",
				zap.String("version", migration.Version()),
				zap.String("description", migration.Description()))
		}
	}

	logger.Info("==================")
	logger.Info("Summary",
		zap.Int("applied", appliedCount),
		zap.Int("pending", pendingCount),
		zap.Int("total", len(allMigrations)))

	return nil
}

// GetAppliedMigrations returns all applied migrations
func (m *Manager) GetAppliedMigrations() ([]MigrationRecord, error) {
	var records []MigrationRecord
	if err := m.db.Order("applied_at ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// GetPendingMigrations returns all pending migrations
func (m *Manager) GetPendingMigrations() ([]Migration, error) {
	appliedRecords, err := m.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	appliedMap := make(map[string]bool)
	for _, record := range appliedRecords {
		appliedMap[record.Version] = true
	}

	var pendingMigrations []Migration
	for _, migration := range m.GetRegisteredMigrations() {
		if !appliedMap[migration.Version()] {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	return pendingMigrations, nil
}

// IsMigrationApplied checks if a migration is applied
func (m *Manager) IsMigrationApplied(version string) (bool, error) {
	var count int64
	err := m.db.Model(&MigrationRecord{}).Where("version = ?", version).Count(&count).Error
	return count > 0, err
}

// Private methods

func (m *Manager) ensureMigrationsTable() error {
	return m.db.AutoMigrate(&MigrationRecord{})
}

func (m *Manager) runSingleMigration(migration Migration) error {
	// Start transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Run migration
	if err := migration.Up(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("migration failed: %w", err)
	}

	// Record migration
	record := MigrationRecord{
		Version:     migration.Version(),
		Description: migration.Description(),
		AppliedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

func (m *Manager) rollbackSingleMigration(migration Migration, record MigrationRecord) error {
	// Start transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Run rollback
	if err := migration.Down(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("rollback failed: %w", err)
	}

	// Remove migration record
	if err := tx.Delete(&record).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	return nil
}

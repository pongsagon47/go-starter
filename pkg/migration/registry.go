package migration

import (
	"sync"

	"gorm.io/gorm"
)

// GlobalRegistry เป็น global registry สำหรับ migrations
type GlobalRegistry struct {
	migrations []Migration
	mu         sync.RWMutex
}

// Global instance
var globalRegistry = &GlobalRegistry{
	migrations: make([]Migration, 0),
}

// Register registers a migration globally (called from migration files' init())
func Register(migration Migration) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.migrations = append(globalRegistry.migrations, migration)
}

// GetRegisteredMigrations returns all globally registered migrations
func GetRegisteredMigrations() []Migration {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	// Return a copy to avoid race conditions
	migrations := make([]Migration, len(globalRegistry.migrations))
	copy(migrations, globalRegistry.migrations)
	return migrations
}

// ClearRegistry clears the global registry (useful for testing)
func ClearRegistry() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.migrations = globalRegistry.migrations[:0]
}

// RegisterMigrationsToManager registers all global migrations to a manager
func RegisterMigrationsToManager(manager *Manager) {
	migrations := GetRegisteredMigrations()
	for _, migration := range migrations {
		manager.RegisterMigration(migration)
	}
}

// NewManagerWithGlobalMigrations creates a new manager with all global migrations
func NewManagerWithGlobalMigrations(db interface{}, config *MigrationConfig) *Manager {
	// Type assertion for GORM DB
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		// Try to get GORM DB from interface
		if dbInterface, ok := db.(interface{ GetDB() *gorm.DB }); ok {
			gormDB = dbInterface.GetDB()
		} else {
			panic("database must be *gorm.DB or implement GetDB() method")
		}
	}

	manager := NewManager(gormDB, config)
	RegisterMigrationsToManager(manager)
	return manager
}

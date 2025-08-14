package seeder

import (
	"sync"

	"gorm.io/gorm"
)

// GlobalRegistry เป็น global registry สำหรับ seeders
type GlobalRegistry struct {
	seeders []Seeder
	mu      sync.RWMutex
}

// Global instance
var globalRegistry = &GlobalRegistry{
	seeders: make([]Seeder, 0),
}

// Register registers a seeder globally (called from seeder files' init())
func Register(seeder Seeder) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.seeders = append(globalRegistry.seeders, seeder)
}

// GetRegisteredSeeders returns all globally registered seeders
func GetRegisteredSeeders() []Seeder {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	// Return a copy to avoid race conditions
	seeders := make([]Seeder, len(globalRegistry.seeders))
	copy(seeders, globalRegistry.seeders)
	return seeders
}

// ClearRegistry clears the global registry (useful for testing)
func ClearRegistry() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.seeders = globalRegistry.seeders[:0]
}

// RegisterSeedersToManager registers all global seeders to a manager
func RegisterSeedersToManager(manager *Manager) {
	seeders := GetRegisteredSeeders()
	for _, seeder := range seeders {
		manager.RegisterSeeder(seeder)
	}
}

// NewManagerWithGlobalSeeders creates a new manager with all global seeders
func NewManagerWithGlobalSeeders(db interface{}, config *SeederConfig) *Manager {
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
	RegisterSeedersToManager(manager)
	return manager
}

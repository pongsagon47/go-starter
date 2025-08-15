// internal/migrations/manager.go - Migration Manager สำหรับระบบ Laravel-style
// This file is kept for backward compatibility and will delegate to pkg/migration
package migrations

import (
	"go-starter/pkg/migration"

	"gorm.io/gorm"
)

// Migration interface ที่แต่ละไฟล์ต้อง implement (backward compatibility)
type Migration = migration.Migration

// Migrations represents migration history in database (backward compatibility)
type Migrations = migration.MigrationRecord

// MigrationManager จัดการ migrations (backward compatibility)
type MigrationManager = migration.Manager

// Global migration manager instance (backward compatibility)
var globalManager *migration.Manager

// NewMigrationManager สร้าง manager ใหม่ (backward compatibility)
func NewMigrationManager(db *gorm.DB) *migration.Manager {
	config := migration.DefaultMigrationConfig()
	manager := migration.NewManagerWithGlobalMigrations(db, config)
	return manager
}

// SetGlobalManager ตั้งค่า global manager (backward compatibility)
func SetGlobalManager(manager *migration.Manager) {
	globalManager = manager
}

// Register ฟังก์ชันสำหรับให้แต่ละไฟล์เรียกใช้ใน init() (backward compatibility)
func Register(migrationInstance Migration) {
	migration.Register(migrationInstance)
}

// ลงทะเบียน migrations ทั้งหมดใน package นี้
func init() {
	// Manual registration will be handled by generated migrations_generated.go
	// migration.Register(&CreateUsersTable{})
}

// internal/seeders/manager.go - Enhanced Seeder Manager with Dependencies
// This file is kept for backward compatibility and will delegate to pkg/seeder
package seeders

import (
	"flex-service/pkg/seeder"

	"gorm.io/gorm"
)

// Seeder interface ที่แต่ละไฟล์ต้อง implement (backward compatibility)
type Seeder = seeder.Seeder

// SeederManager จัดการ seeders (backward compatibility)
type SeederManager = seeder.Manager

// Global seeder manager instance (backward compatibility)
var globalSeederManager *seeder.Manager

// NewSeederManager สร้าง seeder manager ใหม่ (backward compatibility)
func NewSeederManager(db *gorm.DB) *seeder.Manager {
	config := seeder.DefaultSeederConfig()
	manager := seeder.NewManagerWithGlobalSeeders(db, config)
	return manager
}

// SetGlobalSeederManager ตั้งค่า global seeder manager (backward compatibility)
func SetGlobalSeederManager(manager *seeder.Manager) {
	globalSeederManager = manager
}

// Register ฟังก์ชันสำหรับให้แต่ละไฟล์เรียกใช้ใน init() (backward compatibility)
func Register(seederInstance Seeder) {
	seeder.Register(seederInstance)
}

package seeder

import (
	"gorm.io/gorm"
)

// Seeder interface ที่แต่ละ seeder file ต้อง implement
type Seeder interface {
	Run(db *gorm.DB) error
	Name() string
	Dependencies() []string
}

// SeederEngine interface สำหรับจัดการ seeders
type SeederEngine interface {
	// Registration
	RegisterSeeder(seeder Seeder)
	GetRegisteredSeeders() []Seeder

	// Execution
	RunSeeders(seederName string) error
	RunSpecificSeeder(seederName string) error
	ListSeeders() error

	// Dependencies
	ResolveDependencies() ([]Seeder, error)
	ResolveDependenciesFor(seederName string) ([]Seeder, error)
}

// SeederConfig configuration for seeder engine
type SeederConfig struct {
	AutoRun        bool     // Auto run seeders on startup
	DefaultSeeders []string // Default seeders to run
	SkipIfExists   bool     // Skip seeder if data already exists
	FailOnError    bool     // Fail on first error or continue
}

// DefaultSeederConfig returns default configuration
func DefaultSeederConfig() *SeederConfig {
	return &SeederConfig{
		AutoRun:        false,
		DefaultSeeders: []string{},
		SkipIfExists:   true,
		FailOnError:    true,
	}
}

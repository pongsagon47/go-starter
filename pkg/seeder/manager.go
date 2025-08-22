package seeder

import (
	"fmt"
	"strings"

	"flex-service/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Manager implements SeederEngine interface
type Manager struct {
	db      *gorm.DB
	config  *SeederConfig
	seeders []Seeder
}

// NewManager creates a new seeder manager
func NewManager(db *gorm.DB, config *SeederConfig) *Manager {
	if config == nil {
		config = DefaultSeederConfig()
	}

	return &Manager{
		db:      db,
		config:  config,
		seeders: make([]Seeder, 0),
	}
}

// RegisterSeeder registers a seeder
func (sm *Manager) RegisterSeeder(seeder Seeder) {
	sm.seeders = append(sm.seeders, seeder)
}

// GetRegisteredSeeders returns all registered seeders
func (sm *Manager) GetRegisteredSeeders() []Seeder {
	// Return a copy to avoid modification
	seeders := make([]Seeder, len(sm.seeders))
	copy(seeders, sm.seeders)
	return seeders
}

// RunSeeders runs seeders (all or specific)
func (sm *Manager) RunSeeders(seederName string) error {
	if len(sm.seeders) == 0 {
		logger.Info("No seeders found")
		return nil
	}

	logger.Info("Starting database seeding...",
		zap.Int("total_seeders", len(sm.seeders)))

	if seederName != "" {
		if !strings.HasSuffix(seederName, "Seeder") {
			seederName += "Seeder"
		}

		if err := sm.RunSpecificSeeder(seederName); err != nil {
			logger.Error("Seeder failed",
				zap.String("name", seederName),
				zap.Error(err))
			return fmt.Errorf("seeder %s failed: %w", seederName, err)
		}

		logger.Info("Seeder completed successfully", zap.String("name", seederName))
		return nil
	}

	// Run all seeders with dependency resolution
	orderedSeeders, err := sm.ResolveDependencies()
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	successCount := 0
	for _, seeder := range orderedSeeders {
		logger.Info("Running seeder", zap.String("name", seeder.Name()))

		if err := seeder.Run(sm.db); err != nil {
			logger.Error("Seeder failed",
				zap.String("name", seeder.Name()),
				zap.Error(err))

			if sm.config.FailOnError {
				return fmt.Errorf("seeder %s failed: %w", seeder.Name(), err)
			} else {
				logger.Warn("Continuing despite seeder failure",
					zap.String("name", seeder.Name()))
				continue
			}
		}

		successCount++
		logger.Info("Seeder completed successfully", zap.String("name", seeder.Name()))
	}

	logger.Info("All seeders completed successfully", zap.Int("count", successCount))
	return nil
}

// RunSpecificSeeder runs a specific seeder with its dependencies
func (sm *Manager) RunSpecificSeeder(seederName string) error {
	// Find target seeder
	var targetSeeder Seeder
	for _, seeder := range sm.seeders {
		if seeder.Name() == seederName {
			targetSeeder = seeder
			break
		}
	}

	if targetSeeder == nil {
		return fmt.Errorf("seeder %s not found", seederName)
	}

	// Resolve dependencies for this seeder
	toRun, err := sm.ResolveDependenciesFor(seederName)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies for %s: %w", seederName, err)
	}

	// Run seeders in dependency order
	for _, seeder := range toRun {
		logger.Info("Running seeder", zap.String("name", seeder.Name()))

		if err := seeder.Run(sm.db); err != nil {
			logger.Error("Seeder failed",
				zap.String("name", seeder.Name()),
				zap.Error(err))
			return fmt.Errorf("seeder %s failed: %w", seeder.Name(), err)
		}

		logger.Info("Seeder completed successfully", zap.String("name", seeder.Name()))
	}

	return nil
}

// ListSeeders lists all seeders with their dependencies
func (sm *Manager) ListSeeders() error {
	logger.Info("Registered Seeders:")
	logger.Info("==================")

	if len(sm.seeders) == 0 {
		logger.Info("No seeders registered")
		return nil
	}

	// Try to resolve dependencies for display
	orderedSeeders, err := sm.ResolveDependencies()
	if err != nil {
		logger.Error("Failed to resolve dependencies", zap.Error(err))
		// Fallback to original order
		orderedSeeders = sm.seeders
	}

	for i, seeder := range orderedSeeders {
		deps := seeder.Dependencies()
		if len(deps) > 0 {
			logger.Info(fmt.Sprintf("%d. %s (depends on: %s)",
				i+1, seeder.Name(), strings.Join(deps, ", ")))
		} else {
			logger.Info(fmt.Sprintf("%d. %s", i+1, seeder.Name()))
		}
	}

	logger.Info("==================")
	logger.Info("Total seeders", zap.Int("count", len(sm.seeders)))
	return nil
}

// ResolveDependencies resolves all seeder dependencies using topological sort
func (sm *Manager) ResolveDependencies() ([]Seeder, error) {
	// Create map for seeder lookup
	seederMap := make(map[string]Seeder)
	for _, seeder := range sm.seeders {
		seederMap[seeder.Name()] = seeder
	}

	// Validate all dependencies exist
	for _, seeder := range sm.seeders {
		for _, dep := range seeder.Dependencies() {
			if _, exists := seederMap[dep]; !exists {
				return nil, fmt.Errorf("seeder %s depends on %s but %s not found",
					seeder.Name(), dep, dep)
			}
		}
	}

	// Topological sort using Kahn's algorithm
	return sm.topologicalSort(seederMap)
}

// ResolveDependenciesFor resolves dependencies for a specific seeder
func (sm *Manager) ResolveDependenciesFor(seederName string) ([]Seeder, error) {
	seederMap := make(map[string]Seeder)
	for _, seeder := range sm.seeders {
		seederMap[seeder.Name()] = seeder
	}

	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var result []Seeder

	var visit func(string) error
	visit = func(name string) error {
		if visiting[name] {
			return fmt.Errorf("circular dependency detected involving %s", name)
		}
		if visited[name] {
			return nil
		}

		seeder, exists := seederMap[name]
		if !exists {
			return fmt.Errorf("seeder %s not found", name)
		}

		visiting[name] = true

		// Visit dependencies first
		for _, dep := range seeder.Dependencies() {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true
		result = append(result, seeder)
		return nil
	}

	if err := visit(seederName); err != nil {
		return nil, err
	}

	return result, nil
}

// Private methods

// topologicalSort implements Kahn's algorithm for topological sorting
func (sm *Manager) topologicalSort(seederMap map[string]Seeder) ([]Seeder, error) {
	// Create adjacency list and in-degree count
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize
	for name := range seederMap {
		graph[name] = []string{}
		inDegree[name] = 0
	}

	// Build graph and count in-degrees
	for name, seeder := range seederMap {
		for _, dep := range seeder.Dependencies() {
			graph[dep] = append(graph[dep], name)
			inDegree[name]++
		}
	}

	// Queue for nodes with no dependencies
	var queue []string
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	var result []Seeder
	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]

		result = append(result, seederMap[current])

		// Reduce in-degree of neighbors
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for circular dependency
	if len(result) != len(seederMap) {
		return nil, fmt.Errorf("circular dependency detected in seeders")
	}

	return result, nil
}

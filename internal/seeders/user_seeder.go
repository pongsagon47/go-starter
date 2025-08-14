package seeders

import (
	"fmt"
	"time"

	"go-starter/internal/entity"
	"go-starter/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserSeeder seeds the users table with sample data
type UserSeeder struct{}

// Run executes the seeder
func (s *UserSeeder) Run(db *gorm.DB) error {
	logger.Info("Running UserSeeder...")

	// Check if data already exists
	var count int64
	if err := db.Model(&entity.User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if count > 0 {
		logger.Info("Users already exist, skipping UserSeeder")
		return nil
	}

	// Hash password for demo users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()

	// Sample users data
	users := []entity.User{
		{
			UUID:      uuid.New(),
			Name:      "John Doe",
			Email:     "john@example.com",
			Password:  string(hashedPassword),
			Active:    true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			UUID:      uuid.New(),
			Name:      "Jane Smith",
			Email:     "jane@example.com",
			Password:  string(hashedPassword),
			Active:    true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			UUID:      uuid.New(),
			Name:      "Test User",
			Email:     "test@example.com",
			Password:  string(hashedPassword),
			Active:    false,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// Insert users
	if err := db.Create(&users).Error; err != nil {
		return fmt.Errorf("failed to create users: %w", err)
	}

	logger.Info("UserSeeder completed successfully", zap.Int("users_created", len(users)))
	return nil
}

// Name returns seeder name
func (s *UserSeeder) Name() string {
	return "UserSeeder"
}

// Dependencies returns list of seeders that must run before this seeder
func (s *UserSeeder) Dependencies() []string {
	return []string{} // No dependencies for basic user seeder
}

// Auto-register seeder
func init() {
	Register(&UserSeeder{})
}

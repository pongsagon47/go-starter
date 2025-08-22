package seeders

import (
	"flex-service/internal/entity"
	"flex-service/pkg/logger"
	"flex-service/pkg/utils"
	"time"

	"gorm.io/gorm"
)

// UserSeeder seeds the tb_user table
type UserSeeder struct{}

// Run executes the seeder
func (s *UserSeeder) Run(db *gorm.DB) error {
	logger.Info("Running UserSeeder...")

	// Check if data already exists
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM tb_user").Scan(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		logger.Info("tb_user already exist, skipping UserSeeder")
		return nil
	}

	password, err := utils.HashPassword("test_1234")
	if err != nil {
		return err
	}

	birthDate := time.Date(1996, 10, 28, 0, 0, 0, 0, time.UTC)
	phone := "081234567890"
	email := "test_user@example.com"
	profilePicture := "https://via.placeholder.com/150"
	title := "Mr."
	users := []entity.User{
		{
			MemberNo:       "flex182726151224",
			Username:       "test_user",
			Password:       &password,
			Title:          &title,
			FirstName:      "Test",
			LastName:       "User",
			Gender:         "male",
			BirthDate:      &birthDate,
			Phone:          &phone,
			Email:          &email,
			Active:         entity.UserActive,
			ProfilePicture: &profilePicture,
		},
	}

	if err := db.Create(&users).Error; err != nil {
		return err
	}

	logger.Info("UserSeeder completed successfully")
	return nil
}

// Name returns seeder name
func (s *UserSeeder) Name() string {
	return "UserSeeder"
}

// Dependencies returns list of seeders that must run before this seeder
func (s *UserSeeder) Dependencies() []string {
	return []string{} // No dependencies
}

// Auto-register seeder
func init() {
	Register(&UserSeeder{})
}

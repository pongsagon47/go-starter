package migrations

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User entity struct for migration
type User struct {
	ID        int            `gorm:"primaryKey"`
	UUID      uuid.UUID      `gorm:"type:varchar(36);unique;index;not null"`
	Name      string         `gorm:"type:varchar(255);not null"`
	Email     string         `gorm:"type:varchar(255);unique;not null"`
	Password  string         `gorm:"type:varchar(255);not null"`
	Active    bool           `gorm:"default:true"`
	CreatedAt time.Time      `gorm:"autoCreateTime;not null"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for GORM
func (User) TableName() string {
	return "users"
}

// CreateUsersTable migration - Create users table
type CreateUsersTable struct{}

// Up creates the users table using the User struct
func (m *CreateUsersTable) Up(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

// Down drops the users table
func (m *CreateUsersTable) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&User{})
}

// Description returns migration description
func (m *CreateUsersTable) Description() string {
	return "Create users table"
}

// Version returns migration version
func (m *CreateUsersTable) Version() string {
	return "2025_01_01_000001_create_users_table"
}

// Auto-register migration
func init() {
	Register(&CreateUsersTable{})
}

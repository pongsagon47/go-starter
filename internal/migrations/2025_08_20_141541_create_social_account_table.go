package migrations

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SocialAccount entity struct for migration (MySQL compatible)
type SocialAccount struct {
	ID           int             `gorm:"primaryKey"`
	UUID         uuid.UUID       `gorm:"type:varchar(36);unique;not null"`
	UserID       int             `gorm:"not null;index"`
	Provider     string          `gorm:"type:varchar(255);not null"`
	ProviderId   string          `gorm:"type:varchar(255);not null"`
	ProviderData json.RawMessage `gorm:"type:json;"`
	User         User            `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	CreatedAt    time.Time       `gorm:"autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt  `gorm:"index"`
}

// TableName returns the table name for GORM
func (SocialAccount) TableName() string {
	return "tb_social_account"
}

// CreateSocialAccountTable migration - Create tb_social_account table (MySQL)
type CreateSocialAccountTable struct{}

// Up creates the tb_social_account table using the SocialAccount struct
func (m *CreateSocialAccountTable) Up(db *gorm.DB) error {
	return db.AutoMigrate(&SocialAccount{})
}

// Down drops the tb_social_account table
func (m *CreateSocialAccountTable) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&SocialAccount{})
}

// Description returns migration description
func (m *CreateSocialAccountTable) Description() string {
	return "Create tb_social_account table"
}

// Version returns migration version
func (m *CreateSocialAccountTable) Version() string {
	return "2025_08_20_141541_create_social_account_table"
}

// Auto-register migration
func init() {
	Register(&CreateSocialAccountTable{})
}

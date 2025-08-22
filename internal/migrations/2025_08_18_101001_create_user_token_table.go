package migrations

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserTokenStatus string

const (
	UserTokenActive   UserTokenStatus = "active"
	UserTokenInactive UserTokenStatus = "inactive"
)

// UserToken entity struct for migration (MySQL compatible)
type UserToken struct {
	ID          int             `gorm:"primaryKey"`
	UUID        uuid.UUID       `gorm:"type:varchar(36);unique;not null;index"` // Session ID
	UserID      int             `gorm:"not null;index"`
	User        User            `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	TokenStatus UserTokenStatus `gorm:"type:enum('active', 'inactive');not null;default:inactive;index"`
	AccessJti   string          `gorm:"type:varchar(36);not null;index"` // JTI for Access Token
	RefreshJti  string          `gorm:"type:varchar(36);not null;index"` // JTI for Refresh Token
	RevokedAt   *time.Time      `gorm:"type:datetime;index"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt  `gorm:"index"`
}

// TableName returns the table name for GORM
func (UserToken) TableName() string {
	return "tb_user_token"
}

// CreateUserTokenTable migration - Create tb_user_token table (MySQL)
type CreateUserTokenTable struct{}

// Up creates the tb_user_token table using the UserToken struct
func (m *CreateUserTokenTable) Up(db *gorm.DB) error {
	return db.AutoMigrate(&UserToken{})
}

// Down drops the tb_user_token table
func (m *CreateUserTokenTable) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&UserToken{})
}

// Description returns migration description
func (m *CreateUserTokenTable) Description() string {
	return "Create tb_user_token table"
}

// Version returns migration version
func (m *CreateUserTokenTable) Version() string {
	return "2025_08_18_101001_create_user_token_table"
}

// Auto-register migration
func init() {
	Register(&CreateUserTokenTable{})
}

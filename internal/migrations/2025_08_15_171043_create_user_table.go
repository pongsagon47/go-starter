package migrations

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStatus string

const (
	UserActive   UserStatus = "active"
	UserInactive UserStatus = "inactive"
)

// User entity struct for migration (MySQL compatible)
type User struct {
	ID             int             `gorm:"primaryKey"`
	UUID           uuid.UUID       `gorm:"type:varchar(36);unique;not null;index"`
	MemberNo       string          `gorm:"type:varchar(100);unique;not null;index"`
	Username       string          `gorm:"type:varchar(100);unique;index"`
	Password       *string         `gorm:"type:varchar(100);"`
	Title          *string         `gorm:"type:varchar(100);index"`
	FirstName      string          `gorm:"type:varchar(100);not null;index:idx_full_name"`
	LastName       string          `gorm:"type:varchar(100);not null;index:idx_full_name"`
	Gender         string          `gorm:"type:enum('male', 'female');not null;default:male;index"`
	BirthDate      *time.Time      `gorm:"type:date;index"`
	ProfilePicture *string         `gorm:"type:varchar(255)"`
	Phone          *string         `gorm:"type:varchar(100);index"`
	Email          *string         `gorm:"type:varchar(100);unique;index"`
	Active         UserStatus      `gorm:"type:enum('active', 'inactive');not null;default:inactive;index"`
	CreatedAt      time.Time       `gorm:"autoCreateTime"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt  `gorm:"index"`
	UserTokens     []UserToken     `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	SocialAccounts []SocialAccount `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for GORM
func (User) TableName() string {
	return "tb_user"
}

// CreateUserTable migration - Create tb_user table (MySQL)
type CreateUserTable struct{}

// Up creates the tb_user table using the User struct
func (m *CreateUserTable) Up(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

// Down drops the tb_user table
func (m *CreateUserTable) Down(db *gorm.DB) error {
	return db.Migrator().DropTable(&User{})
}

// Description returns migration description
func (m *CreateUserTable) Description() string {
	return "Create tb_user table"
}

// Version returns migration version
func (m *CreateUserTable) Version() string {
	return "2025_08_15_171043_create_user_table"
}

// Auto-register migration
func init() {
	Register(&CreateUserTable{})
}

// make make-migration NAME=create_admin_token_table CREATE=true TABLE=tb_admin_token STRATEGY=dual FIELDS="admin_id:int|fk:tb_admin,token_status:string,access_jti:string,refresh_jti:string,expired_at:string"

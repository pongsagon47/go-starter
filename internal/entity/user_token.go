package entity

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

// UserToken represents a UserToken entity
type UserToken struct {
	ID          int             `json:"-" gorm:"primaryKey"`
	UUID        uuid.UUID       `json:"uuid" gorm:"type:varchar(36);unique;not null;index"`
	UserID      int             `json:"user_id" gorm:"not null;index"`
	User        User            `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	TokenStatus UserTokenStatus `json:"token_status" gorm:"type:enum('active', 'inactive');not null;default:inactive;index"`
	AccessJti   string          `json:"access_jti" gorm:"type:varchar(36);not null;index"`
	RefreshJti  string          `json:"refresh_jti" gorm:"type:varchar(36);not null;index"`
	RevokedAt   *time.Time      `json:"revoked_at" gorm:"type:datetime;index"`
	CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt  `json:"-" gorm:"index"`
}

// TableName returns the table name for GORM
func (UserToken) TableName() string {
	return "tb_user_token"
}

// BeforeCreate is a hook that runs before creating a UserToken
func (e *UserToken) BeforeCreate(tx *gorm.DB) (err error) {
	e.UUID = uuid.New()
	return
}

// CreateUserTokenRequest represents a request to create a UserToken
type CreateUserTokenRequest struct {
	UserId int `json:"user_id" validate:"required,min=0"`
}

// UpdateUserTokenRequest represents a request to update a UserToken
type UpdateUserTokenRequest struct {
	UserId *int `json:"user_id,omitempty" validate:"omitempty,min=0"`
}

// UserTokenFilter represents filters for UserToken queries
type UserTokenFilter struct {
	Search string `form:"search"`
	Page   int    `form:"page" validate:"min=1"`
	Limit  int    `form:"limit" validate:"min=1,max=100"`
}

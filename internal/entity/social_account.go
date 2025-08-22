package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SocialAccount represents a SocialAccount entity
type SocialAccount struct {
	ID           int             `json:"-" gorm:"primaryKey;autoIncrement"`
	UUID         uuid.UUID       `json:"id" gorm:"type:varchar(36);index;unique;not null"`
	UserID       int             `json:"user_id" gorm:"not null;index"`
	Provider     string          `json:"provider" gorm:"type:varchar(255);not null"`
	ProviderID   string          `json:"provider_id" gorm:"type:varchar(255);not null"`
	ProviderData json.RawMessage `json:"provider_data,omitempty" gorm:"type:json;"`
	User         User            `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	CreatedAt    time.Time       `json:"created_at" gorm:"autoCreateTime;not null"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"autoUpdateTime;not null"`
	DeletedAt    gorm.DeletedAt  `json:"-" gorm:"index"`
}

// TableName returns the table name for GORM
func (SocialAccount) TableName() string {
	return "tb_social_account"
}

// BeforeCreate is a hook that runs before creating a SocialAccount
func (e *SocialAccount) BeforeCreate(tx *gorm.DB) (err error) {
	e.UUID = uuid.New()
	return
}

// CreateSocialAccountRequest represents a request to create a SocialAccount
type CreateSocialAccountRequest struct {
	UserID       int    `json:"user_id" validate:"required,min=0"`
	Provider     string `json:"provider" validate:"required,min=1,max=255"`
	ProviderID   string `json:"provider_id" validate:"required,min=1,max=255"`
	ProviderData string `json:"provider_data" validate:"omitempty,json"`
}

// UpdateSocialAccountRequest represents a request to update a SocialAccount
type UpdateSocialAccountRequest struct {
	UserID       *int    `json:"user_id,omitempty" validate:"omitempty,min=0"`
	Provider     *string `json:"provider,omitempty" validate:"omitempty,min=1,max=255"`
	ProviderID   *string `json:"provider_id,omitempty" validate:"omitempty,min=1,max=255"`
	ProviderData *string `json:"provider_data,omitempty" validate:"omitempty,json"`
}

// SocialAccountFilter represents filters for SocialAccount queries
type SocialAccountFilter struct {
	Provider   string `form:"provider"`
	ProviderID string `form:"provider_id"`
	Search     string `form:"search"`
	Page       int    `form:"page" validate:"min=1"`
	Limit      int    `form:"limit" validate:"min=1,max=100"`
}

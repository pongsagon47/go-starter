package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a basic User entity - example for starter project
type User struct {
	ID        int            `json:"-" gorm:"primaryKey"`
	UUID      uuid.UUID      `json:"uuid" gorm:"type:varchar(36);unique;index;not null"`
	Name      string         `json:"name" gorm:"type:varchar(255);not null"`
	Email     string         `json:"email" gorm:"type:varchar(255);unique;not null"`
	Password  string         `json:"-" gorm:"type:varchar(255);not null"` // Hidden from JSON
	Active    bool           `json:"active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;default:CURRENT_TIMESTAMP(3);not null"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for GORM
func (User) TableName() string {
	return "users"
}

// BeforeCreate is a hook that runs before creating a User
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.UUID = uuid.New()
	return
}

// CreateUserRequest represents a request to create a User
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateUserRequest represents a request to update a User
type UpdateUserRequest struct {
	Name   *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Email  *string `json:"email,omitempty" validate:"omitempty,email"`
	Active *bool   `json:"active,omitempty"`
}

// UserFilter represents filters for User queries
type UserFilter struct {
	Name   string `form:"name"`
	Email  string `form:"email"`
	Active *bool  `form:"active"`
	Search string `form:"search"`
	Page   int    `form:"page" validate:"min=1"`
	Limit  int    `form:"limit" validate:"min=1,max=100"`
}

// UserResponse represents a User response (without sensitive data)
type UserResponse struct {
	UUID      uuid.UUID `json:"uuid"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

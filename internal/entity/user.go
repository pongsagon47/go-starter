package entity

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

type UserGender string

const (
	UserMale   UserGender = "male"
	UserFemale UserGender = "female"
)

// User represents a User entity
type User struct {
	ID             int             `json:"-" gorm:"primaryKey"`
	UUID           uuid.UUID       `json:"uuid" gorm:"type:varchar(36);unique;not null;index"`
	MemberNo       string          `json:"member_no" gorm:"type:varchar(100);unique;not null;index"`
	Username       string          `json:"username" gorm:"type:varchar(100);unique;index"`
	Password       *string         `json:"-" gorm:"type:varchar(100);"`
	Title          *string         `json:"title" gorm:"type:varchar(100);index"`
	FirstName      string          `json:"first_name" gorm:"type:varchar(100);not null;index:idx_full_name"`
	LastName       string          `json:"last_name" gorm:"type:varchar(100);not null;index:idx_full_name"`
	Gender         UserGender      `json:"gender" gorm:"type:enum('male', 'female');not null;default:male;index"`
	BirthDate      *time.Time      `json:"birth_date" gorm:"type:date;index"`
	ProfilePicture *string         `json:"profile_picture" gorm:"type:varchar(255)"`
	Phone          *string         `json:"phone" gorm:"type:varchar(100);index"`
	Email          *string         `json:"email" gorm:"type:varchar(100);unique;index"`
	Active         UserStatus      `json:"active" gorm:"type:enum('active', 'inactive');not null;default:inactive;index"`
	CreatedAt      time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt  `json:"-" gorm:"index"`
	SocialAccounts []SocialAccount `json:"-" gorm:"foreignKey:UserID;references:ID"`
}

// TableName returns the table name for GORM
func (User) TableName() string {
	return "tb_user"
}

// BeforeCreate is a hook that runs before creating a User
func (e *User) BeforeCreate(tx *gorm.DB) (err error) {
	e.UUID = uuid.New()
	return
}

// CreateUserRequest represents a request to create a User
type CreateUserRequest struct {
	Username       string `json:"username" validate:"required,min=3,max=255"`
	Password       string `json:"password" validate:"required,min=8,max=255"`
	Title          string `json:"title" validate:"required,min=1,max=255"`
	FirstName      string `json:"first_name" validate:"required,min=1,max=255"`
	LastName       string `json:"last_name" validate:"required,min=1,max=255"`
	Gender         string `json:"gender" validate:"required,min=1,max=255"`
	BirthDate      string `json:"birth_date" validate:"required,datetime=2006-01-02"`
	ProfilePicture string `json:"profile_picture,omitempty" validate:"omitempty,min=1,max=255"`
	Email          string `json:"email" validate:"required,min=1,max=255"`
	Phone          string `json:"phone" validate:"required,min=1,max=255"`
}

// UpdateUserRequest represents a request to update a User
type UpdateUserRequest struct {
	Username       *string `json:"username,omitempty" validate:"omitempty,min=1,max=255"`
	Password       *string `json:"password,omitempty" validate:"omitempty,min=1,max=255"`
	Title          *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	FirstName      *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=255"`
	LastName       *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=255"`
	Gender         *string `json:"gender,omitempty" validate:"omitempty,min=1,max=255"`
	BirthDate      *string `json:"birth_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	ProfilePicture *string `json:"profile_picture,omitempty" validate:"omitempty,min=1,max=255"`
	Email          *string `json:"email,omitempty" validate:"omitempty,min=1,max=255"`
	Phone          *string `json:"phone,omitempty" validate:"omitempty,min=1,max=255"`
}

// UserFilter represents filters for User queries
type UserFilter struct {
	MemberNo       string     `form:"member_no"`
	Username       string     `form:"username"`
	Title          string     `form:"title"`
	FirstName      string     `form:"first_name"`
	LastName       string     `form:"last_name"`
	Gender         string     `form:"gender"`
	BirthDate      time.Time  `form:"birth_date"`
	ProfilePicture string     `form:"profile_picture"`
	Phone          string     `form:"phone"`
	Active         UserStatus `form:"active"`
	Search         string     `form:"search"`
	Page           int        `form:"page" validate:"min=1"`
	Limit          int        `form:"limit" validate:"min=1,max=1000"`
}

func (u *User) IsActive() bool {
	return u.Active == UserActive
}

// GetFullName returns the full name of the user
func (u *User) GetFullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	return u.FirstName + " " + u.LastName
}

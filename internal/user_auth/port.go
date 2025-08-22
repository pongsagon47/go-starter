package user_auth

import (
	"context"
	"flex-service/internal/entity"

	"github.com/google/uuid"
)

// AuthRequest structures
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

// AuthResponse structures
type AuthResponse struct {
	User         *entity.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

type GenerateTokensResponse struct {
	AccessJti    string
	RefreshJti   string
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 `json:"expires_in"`
}

type UpdateUserTokenRequest struct {
	UserID        int
	OldRefreshJti string
	AccessJti     string
	RefreshJti    string
}

type ValidateTokenResponse struct {
	User       *entity.User
	UserClaims *UserClaims
}

type LoginWithSocialAccountRequest struct {
	Provider   string `json:"provider" validate:"required"`
	ProviderID string `json:"provider_id" validate:"required"`
}

type RegisterWithSocialAccountRequest struct {
	Provider     string `json:"provider" validate:"required,oneof=google facebook apple"`
	ProviderID   string `json:"provider_id" validate:"required"`
	ProviderData string `json:"provider_data" validate:"omitempty,json"`
	FirstName    string `json:"first_name" validate:"required,min=3"`
	LastName     string `json:"last_name" validate:"required,min=3"`
	Email        string `json:"email" validate:"required,email"`
	Phone        string `json:"phone" validate:"omitempty,min=10,max=10"`
	BirthDate    string `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
}

// AuthUsecase defines the business logic interface for auth
type UserAuthUsecase interface {
	Register(ctx context.Context, req *entity.CreateUserRequest) (*AuthResponse, error)
	RegisterWithSocialAccount(ctx context.Context, req *RegisterWithSocialAccountRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	LoginWithSocialAccount(ctx context.Context, req *LoginWithSocialAccountRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*AuthResponse, error)
	Logout(ctx context.Context, token string, userID int) error
	GetUserByID(ctx context.Context, userID int) (*entity.User, error)
	GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*entity.User, error)
	ValidateToken(ctx context.Context, token string) (*ValidateTokenResponse, error)
	GetUserProfile(ctx context.Context, userID int) (*entity.User, error)
	// TODO: Add password reset methods
	// ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error
	// ResetPassword(ctx context.Context, req *ResetPasswordRequest) error
}

// AuthRepository defines the data access interface for auth
type UserAuthRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	CreateUserToken(ctx context.Context, userID int, accessJti string, refreshJti string) error
	UpdateUserToken(ctx context.Context, req *UpdateUserTokenRequest) error
	RevokeAccessTokenByJTI(ctx context.Context, jti string) error
	GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*entity.User, error)
	GetUserTokenByAccessJti(ctx context.Context, accessJti string) (*entity.UserToken, error)
	GetUserBySocialAccount(ctx context.Context, provider, providerID string) (*entity.User, error)
	CreateSocialAccount(ctx context.Context, req *RegisterWithSocialAccountRequest) (*entity.User, error)
}

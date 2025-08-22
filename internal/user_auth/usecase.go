package user_auth

import (
	"context"
	"flex-service/internal/entity"
	"fmt"
	"time"

	"flex-service/pkg/cache"
	"flex-service/pkg/errors"
	"flex-service/pkg/logger"
	"flex-service/pkg/utils"

	"github.com/google/uuid"

	"go.uber.org/zap"
)

type userAuthUsecase struct {
	repo  UserAuthRepository
	jwt   *UserJWT
	cache cache.Cache
}

func NewUserAuthUsecase(repo UserAuthRepository, jwt *UserJWT, cache cache.Cache) UserAuthUsecase {
	return &userAuthUsecase{
		repo:  repo,
		jwt:   jwt,
		cache: cache,
	}
}

func (u *userAuthUsecase) Register(ctx context.Context, req *entity.CreateUserRequest) (*AuthResponse, error) {
	logger.Info("Register attempt", zap.String("email", req.Email), zap.String("username", req.Username))

	if _, err := u.repo.GetUserByUsername(ctx, req.Username); err == nil {
		return nil, errors.UserExists("Username")
	}

	fmt.Println("req.Password", req.Password)
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.WrapInternal(err, "failed to hash password")
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return nil, errors.WrapInternal(err, "Invalid birth date format")
	}

	memberNo, err := GenerateMemberNo()
	if err != nil {
		return nil, errors.WrapInternal(err, "failed to generate member no")
	}

	user := &entity.User{
		UUID:           uuid.New(),
		Email:          &req.Email,
		Username:       req.Username,
		Password:       &hashedPassword,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Title:          &req.Title,
		Gender:         entity.UserGender(req.Gender),
		BirthDate:      &birthDate,
		ProfilePicture: &req.ProfilePicture,
		Phone:          &req.Phone,
		Active:         entity.UserActive,
		MemberNo:       memberNo,
	}

	if err := u.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	token, err := u.generateTokens(ctx, user)
	if err != nil {
		return nil, errors.WrapTokenError(err, "failed to generate tokens")
	}

	logger.Info("User registered successfully", zap.String("user_id", user.UUID.String()))

	if err := u.repo.CreateUserToken(ctx, user.ID, token.AccessJti, token.RefreshJti); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
	}, nil
}

func (u *userAuthUsecase) RegisterWithSocialAccount(ctx context.Context, req *RegisterWithSocialAccountRequest) (*AuthResponse, error) {
	logger.Info("Register with social account attempt", zap.String("provider", req.Provider), zap.String("provider_id", req.ProviderID))

	user, err := u.repo.GetUserBySocialAccount(ctx, req.Provider, req.ProviderID)
	if err != nil {
		return nil, err
	}

	fmt.Println("user", user)

	if user != nil {
		return nil, errors.UserExists("Provider")
	}

	if user, err = u.repo.CreateSocialAccount(ctx, req); err != nil {
		return nil, err
	}

	logger.Info("Social account created successfully",
		zap.Int("user_id", user.ID),
		zap.String("user_uuid", user.UUID.String()))

	token, err := u.generateTokens(ctx, user)
	if err != nil {
		return nil, errors.WrapTokenError(err, "failed to generate tokens")
	}

	if err := u.repo.CreateUserToken(ctx, user.ID, token.AccessJti, token.RefreshJti); err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
	}, nil
}

func (u *userAuthUsecase) GetUserProfile(ctx context.Context, userID int) (*entity.User, error) {
	cacheKey := fmt.Sprintf("user:profile:%d", userID)

	if u.cache != nil {
		// u.cache.Del(ctx, cacheKey)
		var user entity.User
		if err := u.cache.GetJSON(ctx, cacheKey, &user); err == nil {
			return &user, nil
		}
	}

	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if u.cache != nil {
		u.cache.SetJSON(ctx, cacheKey, user, 30*time.Minute)
	}

	return user, nil
}

func (u *userAuthUsecase) InvalidateUserCache(ctx context.Context, userID int) error {
	if u.cache == nil {
		return nil
	}

	return u.cache.Del(ctx, fmt.Sprintf("user:profile:%d", userID))
}

func (u *userAuthUsecase) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	logger.Info("Login attempt", zap.String("identifier", req.Username))

	var user *entity.User
	var err error

	if user, err = u.repo.GetUserByUsername(ctx, req.Username); err != nil {
		return nil, errors.InvalidCredentials()
	}

	if !user.IsActive() {
		return nil, errors.AccountDisabled()
	}

	if !utils.VerifyPassword(req.Password, *user.Password) {
		return nil, errors.InvalidCredentials()
	}

	token, err := u.generateTokens(ctx, user)
	if err != nil {
		return nil, errors.WrapTokenError(err, "failed to generate tokens")
	}

	if err := u.repo.CreateUserToken(ctx, user.ID, token.AccessJti, token.RefreshJti); err != nil {
		return nil, err
	}

	logger.Info("User logged in successfully", zap.String("user_id", user.UUID.String()))

	return &AuthResponse{
		User:         user,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
	}, nil
}

func (u *userAuthUsecase) LoginWithSocialAccount(ctx context.Context, req *LoginWithSocialAccountRequest) (*AuthResponse, error) {
	logger.Info("Login with social account attempt", zap.String("provider", req.Provider), zap.String("provider_id", req.ProviderID))

	user, err := u.repo.GetUserBySocialAccount(ctx, req.Provider, req.ProviderID)
	if err != nil {
		return nil, errors.WrapDatabase(err, "failed to get user by social account")
	}

	fmt.Println("user", user)

	if !user.IsActive() {
		return nil, errors.AccountDisabled()
	}

	token, err := u.generateTokens(ctx, user)
	if err != nil {
		return nil, errors.WrapTokenError(err, "failed to generate tokens")
	}

	if err := u.repo.CreateUserToken(ctx, user.ID, token.AccessJti, token.RefreshJti); err != nil {
		return nil, err
	}

	logger.Info("User logged in successfully", zap.String("user_id", user.UUID.String()))

	return &AuthResponse{
		User:         user,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
	}, nil
}

func (u *userAuthUsecase) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*AuthResponse, error) {
	logger.Info("Refresh token attempt", zap.String("refresh_token", req.RefreshToken))
	claims, err := u.jwt.ValidateUserToken(req.RefreshToken)
	if err != nil {
		return nil, errors.TokenInvalid()
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.TokenInvalid()
	}

	user, err := u.repo.GetUserByUUID(ctx, uuid.MustParse(claims.UUID))
	if err != nil {
		return nil, err
	}

	if !user.IsActive() {
		return nil, errors.AccountDisabled()
	}

	oldRefreshJti := claims.ID

	token, err := u.generateTokens(ctx, user)
	if err != nil {
		return nil, errors.WrapTokenError(err, "failed to generate tokens")
	}

	if err := u.repo.UpdateUserToken(ctx, &UpdateUserTokenRequest{
		UserID:        user.ID,
		OldRefreshJti: oldRefreshJti,
		AccessJti:     token.AccessJti,
		RefreshJti:    token.RefreshJti,
	}); err != nil {
		return nil, err
	}

	logger.Info("Token refreshed successfully", zap.String("user_id", user.UUID.String()))

	return &AuthResponse{
		User:         user,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
	}, nil
}

func (u *userAuthUsecase) Logout(ctx context.Context, token string, userID int) error {
	logger.Info("User logged out", zap.Int("user_id", userID))

	claims, err := u.jwt.ValidateUserToken(token)
	if err != nil {
		return errors.TokenInvalid()
	}

	if claims.TokenType != TokenTypeAccess {
		return errors.TokenInvalid()
	}

	accessJti := claims.ID

	if u.cache != nil {
		blacklistKey := fmt.Sprintf("token:blacklist:%s", token)
		u.cache.Set(ctx, blacklistKey, "revoked", 24*time.Hour)
	}

	if err := u.repo.RevokeAccessTokenByJTI(ctx, accessJti); err != nil {
		return errors.WrapDatabase(err, "failed to update user token")
	}

	return nil
}

func (u *userAuthUsecase) GetUserByID(ctx context.Context, userID int) (*entity.User, error) {
	return u.repo.GetUserByID(ctx, userID)
}

func (u *userAuthUsecase) GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*entity.User, error) {
	return u.repo.GetUserByUUID(ctx, userUUID)
}

func (u *userAuthUsecase) generateTokens(ctx context.Context, user *entity.User) (*GenerateTokensResponse, error) {

	accessJti := utils.GenerateUUID().String()
	refreshJti := utils.GenerateUUID().String()

	accessToken, accessJti, err := u.jwt.GenerateUserToken(user.UUID.String(), *user.Email, TokenTypeAccess, accessJti)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshJti, err := u.jwt.GenerateUserToken(user.UUID.String(), *user.Email, TokenTypeRefresh, refreshJti)
	if err != nil {
		return nil, err
	}

	return &GenerateTokensResponse{
		AccessJti:    accessJti,
		RefreshJti:   refreshJti,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(u.jwt.accessTokenTTL.Seconds()),
	}, nil
}

func (u *userAuthUsecase) ValidateToken(ctx context.Context, token string) (*ValidateTokenResponse, error) {
	if u.cache != nil {
		blacklistKey := fmt.Sprintf("token:blacklist:%s", token)
		if exists, _ := u.cache.Exists(ctx, blacklistKey); exists > 0 {
			return nil, errors.TokenInvalid()
		}
	}

	claims, err := u.jwt.ValidateUserToken(token)
	if err != nil {
		return nil, errors.TokenInvalid()
	}

	accessJti := claims.ID
	if err != nil {
		return nil, errors.TokenInvalid()
	}

	userToken, err := u.repo.GetUserTokenByAccessJti(ctx, accessJti)
	if err != nil {
		return nil, errors.TokenInvalid()
	}

	return &ValidateTokenResponse{
		User:       &userToken.User,
		UserClaims: claims,
	}, nil
}

func GenerateMemberNo() (string, error) {
	randomString, err := utils.GenerateRandomString(12)
	if err != nil {
		return "", errors.WrapInternal(err, "failed to generate random string")
	}
	memberNo := fmt.Sprintf("%s%s", "flex", randomString)
	return memberNo, nil
}

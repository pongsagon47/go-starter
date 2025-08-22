package user_auth

import (
	"context"
	"encoding/json"
	"flex-service/internal/entity"
	"flex-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userAuthRepository struct {
	db *gorm.DB
}

func NewUserAuthRepository(db *gorm.DB) UserAuthRepository {
	return &userAuthRepository{
		db: db,
	}
}

func (r *userAuthRepository) CreateUser(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return errors.WrapDatabase(err, "failed to create user")
	}
	return nil
}

func (r *userAuthRepository) CreateSocialAccount(ctx context.Context, req *RegisterWithSocialAccountRequest) (*entity.User, error) {

	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	memberNo, err := GenerateMemberNo()
	if err != nil {
		return nil, errors.WrapInternal(err, "failed to generate member no")
	}

	user := &entity.User{
		UUID:      uuid.New(),
		Email:     &req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		MemberNo:  memberNo,
	}

	if req.Email != "" {
		user.Username = req.Email
	} else if req.Phone != "" {
		user.Username = req.Phone
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, errors.WrapDatabase(err, "failed to create user")
	}

	socialAccount := &entity.SocialAccount{
		UserID:     user.ID,
		Provider:   req.Provider,
		ProviderID: req.ProviderID,
	}

	if req.ProviderData != "" {
		socialAccount.ProviderData = json.RawMessage(req.ProviderData)
	}

	if err := tx.Create(socialAccount).Error; err != nil {
		tx.Rollback()
		return nil, errors.WrapDatabase(err, "failed to create social account")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WrapDatabase(err, "failed to commit transaction")
	}

	return user, nil
}

func (r *userAuthRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.UserNotFound()
		}
		return nil, errors.WrapDatabase(err, "failed to get user by email")
	}
	return &user, nil
}

func (r *userAuthRepository) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.UserNotFound()
		}
		return nil, errors.WrapDatabase(err, "failed to get user by username")
	}
	return &user, nil
}

func (r *userAuthRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.UserNotFound()
		}
		return nil, errors.WrapDatabase(err, "failed to get user by id")
	}
	return &user, nil
}

func (r *userAuthRepository) GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return nil, errors.WrapDatabase(err, "failed to get user by uuid")
	}
	return &user, nil
}

func (r *userAuthRepository) GetUserBySocialAccount(ctx context.Context, provider, providerID string) (*entity.User, error) {
	var socialAccount entity.SocialAccount
	if err := r.db.WithContext(ctx).Preload("User").Where("provider = ? AND provider_id = ?", provider, providerID).First(&socialAccount).Error; err != nil {
		return nil, err
	}

	return &socialAccount.User, nil
}

func (r *userAuthRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return errors.WrapDatabase(err, "failed to update user")
	}
	return nil
}

func (r *userAuthRepository) CreateUserToken(ctx context.Context, userID int, accessJti string, refreshJti string) error {

	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	userToken := &entity.UserToken{
		UserID:      userID,
		AccessJti:   accessJti,
		RefreshJti:  refreshJti,
		TokenStatus: entity.UserTokenActive,
	}

	if err := tx.Create(userToken).Error; err != nil {
		tx.Rollback()
		return errors.WrapDatabase(err, "failed to update user token")
	}

	if err := tx.Commit().Error; err != nil {
		return errors.WrapDatabase(err, "failed to commit transaction")
	}

	return nil
}

func (r *userAuthRepository) UpdateUserToken(ctx context.Context, req *UpdateUserTokenRequest) error {
	var userToken entity.UserToken
	if err := r.db.WithContext(ctx).Model(&entity.UserToken{}).
		Where("user_id = ? AND token_status = ? AND refresh_jti = ? AND revoked_at IS NULL", req.UserID, entity.UserTokenActive, req.OldRefreshJti).
		First(&userToken).Error; err != nil {
		return errors.WrapDatabase(err, "failed to update user token")
	}

	userToken.AccessJti = req.AccessJti
	userToken.RefreshJti = req.RefreshJti

	if err := r.db.WithContext(ctx).Save(&userToken).Error; err != nil {
		return errors.WrapDatabase(err, "failed to update user token")
	}

	return nil
}

func (r *userAuthRepository) GetUserTokenByAccessJti(ctx context.Context, accessJti string) (*entity.UserToken, error) {
	var userToken entity.UserToken
	if err := r.db.WithContext(ctx).Preload("User").
		Where("access_jti = ? AND token_status = ? AND revoked_at IS NULL", accessJti, entity.UserTokenActive).
		First(&userToken).Error; err != nil {
		return nil, errors.WrapDatabase(err, "failed to get user token by access jti")
	}
	return &userToken, nil
}

func (r *userAuthRepository) RevokeAccessTokenByJTI(ctx context.Context, jti string) error {
	if err := r.db.WithContext(ctx).Model(&entity.UserToken{}).
		Where("access_jti = ?", jti).
		Update("token_status", entity.UserTokenInactive).Error; err != nil {
		return errors.WrapDatabase(err, "failed to update user token")
	}
	return nil
}

package user_auth

import (
	"flex-service/pkg/utils"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserJWT struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type UserClaims struct {
	UUID      string    `json:"uuid"`
	Email     string    `json:"email"`
	Type      string    `json:"type"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

func NewUserJWT(secret string, accessTTL, refreshTTL time.Duration, issuer string) *UserJWT {
	return &UserJWT{
		secret:          []byte(secret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		issuer:          issuer,
	}
}

func (j *UserJWT) GenerateUserToken(userUUID, email string, tokenType TokenType, jti string) (string, string, error) {

	ttl := j.accessTokenTTL
	if tokenType == TokenTypeRefresh {
		ttl = j.refreshTokenTTL
	}

	if jti == "" {
		jti = utils.GenerateUUID().String()
	}

	claims := UserClaims{
		UUID:      userUUID,
		Email:     email,
		Type:      "user",
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userUUID,
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", "", err
	}

	return tokenString, jti, nil
}

func (j *UserJWT) ValidateUserToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer
	if claims.Issuer != j.issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	// Validate token type if needed
	if claims.TokenType == "" {
		claims.TokenType = TokenTypeAccess // Default for backward compatibility
	}

	// Validate that token is not expired (additional check)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

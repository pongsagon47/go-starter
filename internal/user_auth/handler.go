package user_auth

import (
	"flex-service/internal/entity"
	"net/http"
	"strings"

	"flex-service/pkg/errors"
	"flex-service/pkg/response"
	"flex-service/pkg/validator"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserAuthHandler struct {
	usecase UserAuthUsecase
}

func NewUserAuthHandler(usecase UserAuthUsecase) *UserAuthHandler {
	return &UserAuthHandler{
		usecase: usecase,
	}
}

func (h *UserAuthHandler) Register(c *gin.Context) {
	var req entity.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	if errors := validator.ValidateStruct(&req); errors != nil {
		response.ValidationError(c, "Validation failed", errors)
		return
	}

	result, err := h.usecase.Register(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
		} else {
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		}
		return
	}

	response.Success(c, http.StatusCreated, "User registered successfully", result)
}

func (h *UserAuthHandler) RegisterWithSocialAccount(c *gin.Context) {
	var req RegisterWithSocialAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	if errors := validator.ValidateStruct(&req); errors != nil {
		response.ValidationError(c, "Validation failed", errors)
		return
	}

	result, err := h.usecase.RegisterWithSocialAccount(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
		} else {
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		}
		return
	}

	response.Success(c, http.StatusOK, "Register with social account successful", result)
}

func (h *UserAuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	if errors := validator.ValidateStruct(&req); errors != nil {
		response.ValidationError(c, "Validation failed", errors)
		return
	}

	result, err := h.usecase.Login(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
		} else {
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		}
		return
	}

	response.Success(c, http.StatusOK, "Login successful", result)
}

func (h *UserAuthHandler) LoginWithSocialAccount(c *gin.Context) {
	var req LoginWithSocialAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	if errors := validator.ValidateStruct(&req); errors != nil {
		response.ValidationError(c, "Validation failed", errors)
		return
	}

	result, err := h.usecase.LoginWithSocialAccount(c.Request.Context(), &req)
	if err != nil && err != gorm.ErrRecordNotFound {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		return
	}

	response.Success(c, http.StatusOK, "Login successful", result)
}

func (h *UserAuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err)
		return
	}

	if errors := validator.ValidateStruct(&req); errors != nil {
		response.ValidationError(c, "Validation failed", errors)
		return
	}

	result, err := h.usecase.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
		} else {
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		}
		return
	}

	response.Success(c, http.StatusOK, "Token refreshed successfully", result)
}

func (h *UserAuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	token, err := ExtractTokenFromHeader(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", err.Error(), nil)
		return
	}

	err = h.usecase.Logout(c.Request.Context(), token, userID.(int))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
		} else {
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		}
		return
	}

	response.Success(c, http.StatusOK, "Logout successful", nil)
}

func (h *UserAuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	user, err := h.usecase.GetUserProfile(c.Request.Context(), userID.(int))

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		return
	}

	response.Success(c, http.StatusOK, "User information retrieved successfully", user)
}

func ExtractTokenFromHeader(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header", "missing_authorization_header", http.StatusUnauthorized)
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid token format", "invalid_token_format", http.StatusUnauthorized)
	}

	return parts[1], nil
}

package middleware

import (
	"flex-service/internal/user_auth"
	"flex-service/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UserAuthenticate(userAuthUsecase user_auth.UserAuthUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := user_auth.ExtractTokenFromHeader(c)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", err.Error(), nil)
			c.Abort()
			return
		}

		data, err := userAuthUsecase.ValidateToken(c.Request.Context(), token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", err.Error(), nil)
			c.Abort()
			return
		}

		c.Set("user_id", data.User.ID)
		c.Set("email", data.User.Email)
		c.Set("type", data.UserClaims.Type)
		c.Next()
	}
}

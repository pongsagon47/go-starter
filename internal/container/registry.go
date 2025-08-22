package container

import (
	"errors"
	"flex-service/internal/user_auth"
	"flex-service/pkg/logger"
	"time"
)

// ServiceRegistry manages application service registration
type ServiceRegistry struct {
	container *Container
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(container *Container) *ServiceRegistry {
	return &ServiceRegistry{
		container: container,
	}
}

// RegisterAuth registers authentication-related services
func (r *ServiceRegistry) RegisterUserAuth() error {

	if r.container.Database == nil {
		return errors.New("database dependency not available")
	}

	jwtConfig := r.container.Config.JWT

	accessHours := jwtConfig.ExpirationHours
	if accessHours == 0 {
		accessHours = 24 // default 24 hours
	}
	accessTTL := time.Duration(accessHours) * time.Hour

	refreshHours := jwtConfig.RefreshExpirationHours
	if refreshHours == 0 {
		refreshHours = 720 // default 30 days (720 hours)
	}
	refreshTTL := time.Duration(refreshHours) * time.Hour

	issuer := r.container.Config.AppName

	authJWT := user_auth.NewUserJWT(jwtConfig.Secret, accessTTL, refreshTTL, issuer)

	db := r.container.Database.GetDB()

	// Create auth dependencies
	authRepo := user_auth.NewUserAuthRepository(db)
	authUsecase := user_auth.NewUserAuthUsecase(authRepo, authJWT, r.container.Cache)
	authHandler := user_auth.NewUserAuthHandler(authUsecase)

	// Register in container
	r.container.UserAuthRepo = authRepo
	r.container.UserAuthUsecase = authUsecase
	r.container.UserAuthHandler = authHandler

	logger.Info("User auth services registered successfully")
	return nil
}

// RegisterAll registers all available services
func (r *ServiceRegistry) RegisterAll() error {
	services := []func() error{
		r.RegisterUserAuth,
	}

	for _, registerService := range services {
		if err := registerService(); err != nil {
			return err
		}
	}

	logger.Info("All services registered successfully")
	return nil
}

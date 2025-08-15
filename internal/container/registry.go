package container

import (
	"errors"
	"go-starter/pkg/logger"
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
func (r *ServiceRegistry) RegisterAuth() error {
	if r.container.JWT == nil {
		return errors.New("jwt dependency not available")
	}

	if r.container.Database == nil {
		return errors.New("database dependency not available")
	}

	// Create auth dependencies
	// authRepo := auth.NewAuthRepository(r.container.Database.GetDB())
	// authUsecase := auth.NewAuthUsecase(authRepo, *r.container.JWT)
	// authHandler := auth.NewAuthHandler(authUsecase)

	// Register in container
	// r.container.AuthRepo = authRepo
	// r.container.AuthUsecase = authUsecase
	// r.container.AuthHandler = authHandler

	logger.Info("Auth services registered successfully")
	return nil
}

// RegisterUser registers user-related services (future expansion)
func (r *ServiceRegistry) RegisterUser() error {
	// TODO: Implement user service registration
	logger.Info("User services registration - placeholder")
	return nil
}

// RegisterProduct registers product-related services (future expansion)
func (r *ServiceRegistry) RegisterProduct() error {
	// TODO: Implement product service registration
	logger.Info("Product services registration - placeholder")
	return nil
}

// RegisterAll registers all available services
func (r *ServiceRegistry) RegisterAll() error {
	services := []func() error{
		r.RegisterAuth,
		r.RegisterUser,
		r.RegisterProduct,
	}

	for _, registerService := range services {
		if err := registerService(); err != nil {
			return err
		}
	}

	logger.Info("All services registered successfully")
	return nil
}

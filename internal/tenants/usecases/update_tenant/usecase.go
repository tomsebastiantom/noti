package updatetenant

import (
	"context"

	"getnoti.com/internal/tenants/domain"
	repository "getnoti.com/internal/tenants/repos"
)

type UpdateTenantUseCase interface {
	Execute(ctx context.Context, input UpdateTenantRequest) (UpdateTenantResponse, error)
}

type updateTenantUseCase struct {
	repo repository.TenantsRepository
}

func NewUpdateTenantUseCase(repo repository.TenantsRepository) UpdateTenantUseCase {
	return &updateTenantUseCase{
		repo: repo,
	}
}

func (uc *updateTenantUseCase) Execute(ctx context.Context, req UpdateTenantRequest) (UpdateTenantResponse, error) {
	tenant, err := uc.repo.GetTenantByID(ctx, req.ID)
	if err != nil {
		return UpdateTenantResponse{Success: false}, err
	}

	// Update tenant name
	tenant.Name = req.Name

	// Update DB configurations
	if req.DBConfig != nil {
		// Get the current DB credentials
		currentDBCreds := tenant.DBConfig

		// Update only the fields provided in the request
		updatedDBCreds := domain.DBCredentials{
			Type:     ifNotEmpty(req.DBConfig.Type, currentDBCreds.Type),
			DSN:      ifNotEmpty(req.DBConfig.DSN, currentDBCreds.DSN),
			Host:     ifNotEmpty(req.DBConfig.Host, currentDBCreds.Host),
			Port:     ifNotZero(req.DBConfig.Port, currentDBCreds.Port),
			Username: ifNotEmpty(req.DBConfig.Username, currentDBCreds.Username),
			Password: ifNotEmpty(req.DBConfig.Password, currentDBCreds.Password),
			DBName:   ifNotEmpty(req.DBConfig.DBName, currentDBCreds.DBName),
		}

		// Set the updated DB config
		tenant.SetDBConfig(&updatedDBCreds)
	}

	// Validate the updated tenant
	if err := tenant.Validate(); err != nil {
		return UpdateTenantResponse{Success: false}, err
	}

	if err := uc.repo.Update(ctx, tenant); err != nil {
		return UpdateTenantResponse{Success: false}, err
	}

	return UpdateTenantResponse{Success: true,
		ID:   tenant.ID,
		Name: tenant.Name,
	}, nil
}

func ifNotEmpty(new, current string) string {
	if new != "" {
		return new
	}
	return current
}

func ifNotZero(new, current int) int {
	if new != 0 {
		return new
	}
	return current
}

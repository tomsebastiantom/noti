import os

def create_file_structure(action):
    base_path = f"usecases/{action}"
    os.makedirs(base_path, exist_ok=True)
    
    files_content = {
        "usecase.go": f"""package {action}

import (
    "context"
    "errors"
    "getnoti.com/internal/tenants/domain"
    "getnoti.com/internal/tenants/repos"
)

type GetTenantsUseCase interface {{
    Execute(ctx context.Context, req GetTenantsRequest) (GetTenantsResponse, error)
}}

type getTenantsUseCase struct {{
    repo repos.TenantRepository
}}

func NewGetTenantsUseCase(repo repos.TenantRepository) GetTenantsUseCase {{
    return &getTenantsUseCase{{
        repo: repo,
    }}
}}

func (uc *getTenantsUseCase) Execute(ctx context.Context, req GetTenantsRequest) (GetTenantsResponse, error) {{
    if req.TenantID != "" {{
        tenant, err := uc.repo.GetTenantByID(ctx, req.TenantID)
        if err != nil {{
            return GetTenantsResponse{{}}, err
        }}
        return GetTenantsResponse{{Tenants: []domain.Tenant{{tenant}}}}, nil
    }}

    tenants, err := uc.repo.GetAllTenants(ctx)
    if err != nil {{
        return GetTenantsResponse{{}}, err
    }}
    return GetTenantsResponse{{Tenants: tenants}}, nil
}}
""",
        "controller.go": f"""package {action}

import (
    "context"
)

type GetTenantsController struct {{
    useCase GetTenantsUseCase
}}

func NewGetTenantsController(useCase GetTenantsUseCase) *GetTenantsController {{
    return &GetTenantsController{{useCase: useCase}}
}}

func (c *GetTenantsController) Execute(ctx context.Context, req GetTenantsRequest) (GetTenantsResponse, error) {{
    return c.useCase.Execute(ctx, req)
}}
""",
        "errors.go": f"""package {action}

import "errors"

var (
    ErrTenantNotFound = errors.New("tenant not found")
    ErrInternal       = errors.New("internal error")
)
""",
        "dto.go": f"""package {action}

import "getnoti.com/internal/tenants/domain"

type GetTenantsRequest struct {{
    TenantID string
}}

type GetTenantsResponse struct {{
    Tenants []domain.Tenant
}}
"""
    }

    for file, content in files_content.items():
        with open(os.path.join(base_path, file), 'w') as f:
            f.write(content)

if __name__ == "__main__":
    create_file_structure("gettenants")

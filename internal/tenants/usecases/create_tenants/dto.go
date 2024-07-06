package createtenants
import "getnoti.com/internal/tenants/domain"

type CreateTenantsRequest struct {
    Tenants []domain.Tenant
}



type CreateTenantsResponse struct {
    SuccessTenants []string
    FailedTenants  []FailedTenant
}

type FailedTenant struct {
    ID    string
    Error string
}

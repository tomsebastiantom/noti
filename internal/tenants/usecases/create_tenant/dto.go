package createtenant

import "getnoti.com/internal/tenants/domain"

type CreateTenantRequest struct {
    ID             string
    Name           string
    DBConfig *domain.DBCredentials
}



type CreateTenantResponse struct {
    ID             string
    Name           string
}

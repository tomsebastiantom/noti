package updateuser



type UpdateUserRequest struct {
    UserID        string
    TenantID      string
    Email         string
    PhoneNumber   string
    DeviceID      string
    WebPushToken  string
    Consents      map[string]bool
    PreferredMode string
}
func (r *UpdateUserRequest) SetTenantID(id string) {
    r.TenantID = id
}
type UpdateUserResponse struct {
    Success bool
    Message string
}

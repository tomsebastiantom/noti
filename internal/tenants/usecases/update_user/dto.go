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

type UpdateUserResponse struct {
    Success bool
    Message string
}

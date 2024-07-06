package domain


type User struct {
    ID            string
    TenantID      string
    Email         string
    PhoneNumber   string
    DeviceID      string
    WebPushToken  string
    Consents      map[NotificationChannel]bool
    PreferredMode NotificationChannel
}



package template



type Template struct {
    ID        string
    TenantID  string
    Name      string
    Content   string
    IsPublic  bool
    Variables  []string
}

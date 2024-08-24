package domain

type Template struct {
    ID        string
    Name      string
    Content   string
    IsPublic  bool
    Variables  []string
}

type TemplateVariable struct {
	Key   string
	Value string
}
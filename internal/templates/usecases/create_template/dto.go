package createtemplate

type CreateTemplateRequest struct {
    Name      string
    Content   string
    IsPublic  bool
    Variables []string
}



type CreateTemplateResponse struct {
    Success bool
    Message string
}
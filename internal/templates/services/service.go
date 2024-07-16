package templates

import (
	"context"
	"errors"
	"fmt"
	"getnoti.com/internal/notifications/domain"
	templates "getnoti.com/internal/templates/domain"
	"getnoti.com/internal/templates/repos"
	"strings"
)

type TemplateService struct {
	repo repos.TemplateRepository
}

func NewTemplateService(repo repos.TemplateRepository) *TemplateService {
	return &TemplateService{
		repo: repo,
	}
}

func (s *TemplateService) GetContent(ctx context.Context, templateID string, variables []domain.TemplateVariable) (string, error) {
	// Get the template
	template, err := s.repo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return "", err
	}

	if template == nil {
		return "", errors.New("template not found")
	}

	// Replace variables in the template
	content, err := s.replaceVariables(template, variables)

	return content, err
}

func (s *TemplateService) replaceVariables(template *templates.Template, variables []domain.TemplateVariable) (string, error) {
	content := template.Content
	variableMap := make(map[string]string)
	for _, v := range variables {
		variableMap[v.Key] = v.Value
	}

	missingVariables := []string{}
	for _, key := range template.Variables {
		if value, exists := variableMap[key]; exists {
			content = strings.ReplaceAll(content, "{{"+key+"}}", value)
		} else {
			missingVariables = append(missingVariables, key)
		}
	}

	if len(missingVariables) > 0 {
		return "", fmt.Errorf("missing variables: %s", strings.Join(missingVariables, ", "))
	}

	return content, nil
}

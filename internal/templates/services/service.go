package templates

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"getnoti.com/internal/notifications/domain"
	templates "getnoti.com/internal/templates/domain"
	"getnoti.com/internal/templates/repos"
	tenantServices "getnoti.com/internal/tenants/services"
	"getnoti.com/pkg/logger"
)

type TemplateService struct {
	repo          repos.TemplateRepository
	tenantService *tenantServices.TenantService
	logger        logger.Logger
}

func NewTemplateService(
	repo repos.TemplateRepository,
	tenantService *tenantServices.TenantService,
	logger logger.Logger,
) *TemplateService {
	return &TemplateService{
		repo:          repo,
		tenantService: tenantService,
		logger:        logger,
	}
}

func (s *TemplateService) GetContent(ctx context.Context, tenantID, templateID string, variables []domain.TemplateVariable) (string, error) {
	s.logger.DebugContext(ctx, "Getting template content",
		logger.String("tenant_id", tenantID),
		logger.String("template_id", templateID))

	// Validate tenant access
	err := s.tenantService.ValidateTenantAccess(ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("tenant validation failed: %w", err)
	}

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

	s.logger.DebugContext(ctx, "Template content processed successfully",
		logger.String("tenant_id", tenantID),
		logger.String("template_id", templateID))

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

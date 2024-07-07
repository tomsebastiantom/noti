package updatetemplate

import (
	"context"
	"errors"

	postgres "getnoti.com/internal/templates/repos/implementations"
	"getnoti.com/internal/templates/repos"
)

type UpdateTemplateUseCase interface {
    Execute(ctx context.Context, req UpdateTemplateRequest) UpdateTemplateResponse
}

type updateTemplateUseCase struct {
    repository repos.TemplateRepository
}

func NewUpdateTemplateUseCase(repository repos.TemplateRepository) UpdateTemplateUseCase {
    return &updateTemplateUseCase{repository: repository}
}

func (uc *updateTemplateUseCase) Execute(ctx context.Context, req UpdateTemplateRequest) UpdateTemplateResponse {
    // Check if the template exists
    existingTemplate, err := uc.repository.GetTemplateByID(ctx, req.TemplateID)
    if err != nil {
        if errors.Is(err, postgres.ErrTemplateNotFound) {
            return UpdateTemplateResponse{Success: false, Message: ErrTemplateNotFound.Error()}
        }
        return UpdateTemplateResponse{Success: false, Message: ErrUnexpected.Error()}
    }

    // Update the template with provided fields, retain existing values for fields not provided
    if req.TenantID != nil {
        existingTemplate.TenantID = *req.TenantID
    }
    if req.Name != nil {
        existingTemplate.Name = *req.Name
    }
    if req.Content != nil {
        existingTemplate.Content = *req.Content
    }
    if req.IsPublic != nil {
        existingTemplate.IsPublic = *req.IsPublic
    }
    if req.Variables != nil {
        existingTemplate.Variables = *req.Variables
    }

    err = uc.repository.UpdateTemplate(ctx, existingTemplate)
    if err != nil {
        return UpdateTemplateResponse{Success: false, Message: ErrTemplateUpdateFailed.Error()}
    }

    return UpdateTemplateResponse{
        Template: *existingTemplate,
        Success:  true,
        Message:  "Template updated successfully",
    }
}
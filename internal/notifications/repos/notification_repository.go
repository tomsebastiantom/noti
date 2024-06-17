
package repository

import (
    "context"
    "getnoti.com/internal/notifications/domain"
)

type NotificationRepository interface {
    CreateNotification(ctx context.Context, notification *domain.Notification) error
    GetNotificationByID(ctx context.Context, id string) (*domain.Notification, error)
    UpdateNotification(ctx context.Context, notification *domain.Notification) error
    DeleteNotification(ctx context.Context, id string) error
}

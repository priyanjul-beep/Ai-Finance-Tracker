// Package usecase – notification business logic.
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
)

// unreadCacheTTL is how long the unread count is cached in Redis.
const unreadCacheTTL = 5 * time.Minute

// unreadCacheKey returns the Redis key for a user's unread notification count.
func unreadCacheKey(userID string) string {
	return fmt.Sprintf("notif_unread:%s", userID)
}

// NotificationUseCase handles notification business logic.
type NotificationUseCase struct {
	repo  interfaces.NotificationRepository
	cache interfaces.CacheService
}

// NewNotification creates a new NotificationUseCase.
func NewNotification(repo interfaces.NotificationRepository, cache interfaces.CacheService) *NotificationUseCase {
	return &NotificationUseCase{repo: repo, cache: cache}
}

// List returns a paginated, optionally-filtered list of notifications for the user.
func (uc *NotificationUseCase) List(ctx context.Context, userID, notifType string, page, limit int) (*dto.NotificationListResponse, error) {
	offset := (page - 1) * limit
	rows, total, err := uc.repo.GetByUserIDFiltered(ctx, userID, notifType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("notification: list: %w", err)
	}

	out := make([]dto.NotificationDTO, len(rows))
	for i, n := range rows {
		out[i] = dto.NotificationDTO{
			ID:        n.ID,
			UserID:    n.UserID,
			Title:     n.Title,
			Message:   n.Message,
			Type:      string(n.Type),
			Priority:  string(n.Priority),
			IsRead:    n.IsRead,
			Metadata:  n.Metadata,
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
		}
	}

	unread, _ := uc.GetUnreadCount(ctx, userID)
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &dto.NotificationListResponse{
		Notifications: out,
		Total:         total,
		Page:          page,
		Limit:         limit,
		TotalPages:    totalPages,
		UnreadCount:   unread,
	}, nil
}

// GetUnreadCount returns the number of unread notifications, cached in Redis.
func (uc *NotificationUseCase) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	key := unreadCacheKey(userID)

	// Try cache first
	var count int64
	if err := uc.cache.Get(ctx, key, &count); err == nil {
		return count, nil
	}

	// Cache miss — query DB
	count, err := uc.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("notification: unread count: %w", err)
	}

	// Populate cache
	_ = uc.cache.Set(ctx, key, count, unreadCacheTTL)
	return count, nil
}

// MarkAsRead marks a single notification as read and invalidates the cache.
func (uc *NotificationUseCase) MarkAsRead(ctx context.Context, id, userID string) error {
	if err := uc.repo.MarkAsRead(ctx, id, userID); err != nil {
		return fmt.Errorf("notification: mark read: %w", err)
	}
	_ = uc.cache.Delete(ctx, unreadCacheKey(userID))
	return nil
}

// MarkAllAsRead marks all notifications read for the user and invalidates the cache.
func (uc *NotificationUseCase) MarkAllAsRead(ctx context.Context, userID string) error {
	if err := uc.repo.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("notification: mark all read: %w", err)
	}
	_ = uc.cache.Delete(ctx, unreadCacheKey(userID))
	return nil
}

// Delete removes a notification (ownership checked).
func (uc *NotificationUseCase) Delete(ctx context.Context, id, userID string) error {
	if err := uc.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("notification: delete: %w", err)
	}
	_ = uc.cache.Delete(ctx, unreadCacheKey(userID))
	return nil
}

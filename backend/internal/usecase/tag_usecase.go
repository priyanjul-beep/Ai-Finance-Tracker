// Package usecase – tag business logic.
package usecase

import (
	"context"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
)

// TagUseCase implements tag management.
type TagUseCase struct {
	tags interfaces.TagRepository
}

// NewTag creates a new TagUseCase.
func NewTag(tags interfaces.TagRepository) *TagUseCase {
	return &TagUseCase{tags: tags}
}

// Create adds a new tag for the user.
func (uc *TagUseCase) Create(ctx context.Context, userID string, req dto.CreateTagRequest) (*dto.TagDTO, error) {
	color := req.Color
	if color == "" {
		color = "#6366f1"
	}

	tag := &domain.Tag{
		UserID: userID,
		Name:   req.Name,
		Color:  color,
	}

	if err := uc.tags.Create(ctx, tag); err != nil {
		return nil, err
	}

	return tagToDTO(tag), nil
}

// GetByID returns a tag belonging to the user.
func (uc *TagUseCase) GetByID(ctx context.Context, userID, tagID string) (*dto.TagDTO, error) {
	tag, err := uc.tags.GetByID(ctx, tagID)
	if err != nil {
		return nil, err
	}
	if tag.UserID != userID {
		return nil, interfaces.ErrNotFound
	}
	return tagToDTO(tag), nil
}

// Update modifies an existing tag.
func (uc *TagUseCase) Update(ctx context.Context, userID, tagID string, req dto.UpdateTagRequest) (*dto.TagDTO, error) {
	tag, err := uc.tags.GetByID(ctx, tagID)
	if err != nil {
		return nil, err
	}
	if tag.UserID != userID {
		return nil, interfaces.ErrNotFound
	}

	if req.Name != "" {
		tag.Name = req.Name
	}
	if req.Color != "" {
		tag.Color = req.Color
	}

	if err := uc.tags.Update(ctx, tag); err != nil {
		return nil, err
	}

	return tagToDTO(tag), nil
}

// Delete removes a tag.
func (uc *TagUseCase) Delete(ctx context.Context, userID, tagID string) error {
	tag, err := uc.tags.GetByID(ctx, tagID)
	if err != nil {
		return err
	}
	if tag.UserID != userID {
		return interfaces.ErrNotFound
	}
	return uc.tags.Delete(ctx, tagID)
}

// List returns all tags belonging to the user.
func (uc *TagUseCase) List(ctx context.Context, userID string) ([]*dto.TagDTO, error) {
	tags, err := uc.tags.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]*dto.TagDTO, len(tags))
	for i, t := range tags {
		copy := t
		out[i] = tagToDTO(&copy)
	}
	return out, nil
}

// AddToExpense attaches a tag to an expense.
func (uc *TagUseCase) AddToExpense(ctx context.Context, userID, expenseID, tagID string) error {
	// Verify tag ownership
	tag, err := uc.tags.GetByID(ctx, tagID)
	if err != nil {
		return err
	}
	if tag.UserID != userID {
		return interfaces.ErrNotFound
	}
	return uc.tags.AddToExpense(ctx, expenseID, tagID)
}

// RemoveFromExpense detaches a tag from an expense.
func (uc *TagUseCase) RemoveFromExpense(ctx context.Context, userID, expenseID, tagID string) error {
	tag, err := uc.tags.GetByID(ctx, tagID)
	if err != nil {
		return err
	}
	if tag.UserID != userID {
		return interfaces.ErrNotFound
	}
	return uc.tags.RemoveFromExpense(ctx, expenseID, tagID)
}

// ─── helper ──────────────────────────────────────────────────────────────────

func tagToDTO(t *domain.Tag) *dto.TagDTO {
	return &dto.TagDTO{
		ID:    t.ID,
		Name:  t.Name,
		Color: t.Color,
	}
}

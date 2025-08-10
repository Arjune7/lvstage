package ads

import (
	"context"
	"fmt"

	"lystage-proj/internals/db"
	"lystage-proj/internals/models"

	"gorm.io/gorm"
)

type AdService struct {
	DB *gorm.DB
}

func NewAdService() *AdService {
	return &AdService{DB: db.GormDB}
}

// GetPaginatedAds fetches ads with pagination and total count.
// Context is used for cancellation and timeout safety.
func (s *AdService) GetPaginatedAds(ctx context.Context, page, limit int) ([]models.Ad, int64, error) {
	if page < 1 || limit < 1 {
		return nil, 0, fmt.Errorf("invalid pagination parameters: page=%d limit=%d", page, limit)
	}

	offset := (page - 1) * limit
	var (
		ads   []models.Ad
		total int64
	)

	// Count total ads (filtered by status if needed)
	if err := s.DB.WithContext(ctx).
		Model(&models.Ad{}).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count ads: %w", err)
	}

	// Fetch paginated ads
	if err := s.DB.WithContext(ctx).
		Select("id", "title", "image_url", "target_url", "status").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&ads).Error; err != nil {
		return nil, 0, fmt.Errorf("fetch ads: %w", err)
	}

	return ads, total, nil
}

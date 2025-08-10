package ads

import (
	"net/http"
	"strconv"
	"time"

	"lystage-proj/internals/observability"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxPageLimit = 100

type Handler struct {
	service *AdService
}

func NewHandler(s *AdService) *Handler {
	return &Handler{service: s}
}

// GetAdsHandler returns paginated ads with total count.
func (h *Handler) GetAdsHandler(c *gin.Context) {
	start := time.Now()

	// Parse query params with defaults
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be a positive integer"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
		return
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}

	// Fetch ads
	ads, total, err := h.service.GetPaginatedAds(c.Request.Context(), page, limit)
	if err != nil {
		observability.Logger.Error("failed to get paginated ads",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("limit", limit),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ads"})
		return
	}

	// Transform to API response format
	resp := make([]AdResponse, len(ads))
	for i, ad := range ads {
		resp[i] = AdResponse{
			ID:        ad.ID,
			Title:     ad.Title,
			ImageURL:  ad.ImageURL,
			TargetURL: ad.TargetURL,
			Status:    ad.Status,
		}
	}

	// Observability
	observability.Logger.Info("GET /ads success",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int("result_count", len(resp)),
		zap.Int64("total_count", total),
		zap.Duration("latency_ms", time.Since(start)),
	)

	c.JSON(http.StatusOK, gin.H{
		"page":  page,
		"limit": limit,
		"total": total,
		"data":  resp,
	})
}

package analytics

import (
	"lystage-proj/internals/observability"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
}

type AnalyticsResponse struct {
	Data       []AdAnalytics `json:"data"`
	Total      int           `json:"total"`
	Generated  time.Time     `json:"generated_at"`
	IsRealTime bool          `json:"is_real_time"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetAdAnalytics handles GET /ads/analytics requests with real-time capabilities
func (h *Handler) GetAdAnalytics(c *gin.Context) {
	// Parse query parameters with enhanced options
	filters, err := h.parseAnalyticsQueryParams(c)
	if err != nil {
		observability.Logger.Warn("Invalid analytics request parameters",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log request for monitoring
	observability.Logger.Debug("Analytics request received",
		zap.Int("ad_id", filters.AdID),
		zap.Bool("real_time", filters.RealTime),
		zap.Bool("include_ctr", filters.IncludeCTR),
		zap.String("client_ip", c.ClientIP()))

	// Fetch analytics data using enhanced service
	results, err := h.service.FetchAdAnalyticsWithFilters(filters)
	if err != nil {
		observability.Logger.Error("Failed to fetch ad analytics",
			zap.Error(err),
			zap.Int("ad_id", filters.AdID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch analytics data",
		})
		return
	}

	// Build response with metadata
	response := &AnalyticsResponse{
		Data:       results,
		Total:      len(results),
		Generated:  time.Now(),
		IsRealTime: filters.RealTime,
	}

	// Log successful response
	observability.Logger.Debug("Analytics request completed",
		zap.Int("results_count", len(results)),
		zap.Bool("from_cache", response.IsRealTime))

	c.JSON(http.StatusOK, response)
}

func (h *Handler) parseAnalyticsQueryParams(c *gin.Context) (AnalyticsFilters, error) {
	var filters AnalyticsFilters

	// Parse ad_id (optional)
	if adIDStr := c.Query("ad_id"); adIDStr != "" {
		adID, err := strconv.Atoi(adIDStr)
		if err != nil {
			return filters, err
		}
		filters.AdID = adID
	}

	// Parse pagination with defaults
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return filters, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max limit
	}
	filters.Limit = limit

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return filters, err
	}
	if offset < 0 {
		offset = 0
	}
	filters.Offset = offset

	// Parse time filters
	if sinceStr := c.Query("since"); sinceStr != "" {
		since, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			return filters, err
		}
		filters.Since = since
	}

	if untilStr := c.Query("until"); untilStr != "" {
		until, err := time.Parse(time.RFC3339, untilStr)
		if err != nil {
			return filters, err
		}
		filters.Until = until
	}

	// Parse time_window (e.g., "1h", "30m", "24h")
	if timeWindowStr := c.Query("time_window"); timeWindowStr != "" {
		timeWindow, err := time.ParseDuration(timeWindowStr)
		if err != nil {
			return filters, err
		}
		filters.TimeWindow = timeWindow
	}

	// Parse boolean flags
	filters.RealTime = c.DefaultQuery("real_time", "false") == "true"
	filters.IncludeCTR = c.DefaultQuery("include_ctr", "false") == "true"

	// If no time specified, default to last 24 hours
	if filters.Since.IsZero() && filters.Until.IsZero() && filters.TimeWindow == 0 {
		filters.TimeWindow = 24 * time.Hour
	}

	return filters, nil
}

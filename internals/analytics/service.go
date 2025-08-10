package analytics

import (
	"context"
	"fmt"
	"lystage-proj/internals/db"
	"lystage-proj/internals/observability"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	DB    *gorm.DB
	cache *AnalyticsCache
}

func NewService() *Service {
	cache := &AnalyticsCache{
		metrics: make(map[int]*AdAnalytics),
		ttl:     2 * time.Minute, // Cache TTL for real-time data
	}

	service := &Service{
		DB:    db.GormDB,
		cache: cache,
	}

	// Start background cache refresh for real-time analytics
	go service.startCacheRefresh()

	return service
}

// FetchAdAnalytics - Original method for backward compatibility
func (s *Service) FetchAdAnalytics(adID int, since time.Time, limit, offset int) ([]AdAnalytics, error) {
	filters := AnalyticsFilters{
		AdID:   adID,
		Since:  since,
		Limit:  limit,
		Offset: offset,
	}

	return s.FetchAdAnalyticsWithFilters(filters)
}

// FetchAdAnalyticsWithFilters - Enhanced method for GET /ads/analytics endpoint
func (s *Service) FetchAdAnalyticsWithFilters(filters AnalyticsFilters) ([]AdAnalytics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Validate and set defaults
	if err := s.validateAndSetDefaults(&filters); err != nil {
		return nil, fmt.Errorf("invalid filters: %w", err)
	}

	// Try cache first for real-time requests
	if filters.RealTime {
		if cached := s.getCachedAnalytics(filters); cached != nil {
			observability.Logger.Debug("Serving analytics from cache",
				zap.Int("ad_id", filters.AdID),
				zap.Int("results", len(cached)))
			return cached, nil
		}
		observability.Logger.Debug("Cache miss, falling back to database")
	}

	// Fetch from database
	return s.fetchFromDatabase(ctx, filters)
}

// validateAndSetDefaults ensures filters are valid and sets reasonable defaults
func (s *Service) validateAndSetDefaults(filters *AnalyticsFilters) error {
	// Set default pagination
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 1000 {
		filters.Limit = 1000 // Max limit for performance
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	// Apply time window if specified
	if filters.TimeWindow > 0 && filters.Since.IsZero() {
		filters.Since = time.Now().Add(-filters.TimeWindow)
	}

	// Default to last 24 hours if no time filters specified
	if filters.Since.IsZero() && filters.Until.IsZero() && filters.TimeWindow == 0 {
		filters.Since = time.Now().Add(-24 * time.Hour)
	}

	// Validate time range
	if !filters.Since.IsZero() && !filters.Until.IsZero() && filters.Until.Before(filters.Since) {
		return fmt.Errorf("until time cannot be before since time")
	}

	return nil
}

// getCachedAnalytics retrieves analytics from in-memory cache
func (s *Service) getCachedAnalytics(filters AnalyticsFilters) []AdAnalytics {
	s.cache.mu.RLock()
	defer s.cache.mu.RUnlock()

	var results []AdAnalytics
	now := time.Now()

	if filters.AdID != 0 {
		// Single ad request
		if cached, exists := s.cache.metrics[filters.AdID]; exists {
			if now.Sub(cached.LastUpdated) < s.cache.ttl {
				results = append(results, *cached)
			}
		}
	} else {
		// Multiple ads request - get all valid cached entries
		for _, cached := range s.cache.metrics {
			if now.Sub(cached.LastUpdated) < s.cache.ttl {
				results = append(results, *cached)
			}
		}
	}

	if len(results) == 0 {
		return nil // Cache miss
	}

	// Apply pagination to cached results
	return s.applyPagination(results, filters.Offset, filters.Limit)
}

// applyPagination applies offset and limit to results
func (s *Service) applyPagination(results []AdAnalytics, offset, limit int) []AdAnalytics {
	start := offset
	end := start + limit

	if start >= len(results) {
		return []AdAnalytics{}
	}
	if end > len(results) {
		end = len(results)
	}

	return results[start:end]
}

// fetchFromDatabase queries the database for analytics data
func (s *Service) fetchFromDatabase(ctx context.Context, filters AnalyticsFilters) ([]AdAnalytics, error) {
	var results []AdAnalytics

	// Build optimized query for clicks table
	query := s.buildAnalyticsQuery(ctx, filters)

	// Execute query
	if err := query.Scan(&results).Error; err != nil {
		observability.Logger.Error("Failed to fetch analytics from database",
			zap.Error(err),
			zap.Int("ad_id", filters.AdID))
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	// Add CTR calculation if requested
	if filters.IncludeCTR {
		s.addCTRToResults(ctx, results)
	}

	// Update cache asynchronously for future requests
	go s.updateCache(results)

	observability.Logger.Debug("Analytics fetched from database",
		zap.Int("results", len(results)),
		zap.Int("ad_id", filters.AdID))

	return results, nil
}

// buildAnalyticsQuery constructs the optimized SQL query
func (s *Service) buildAnalyticsQuery(ctx context.Context, filters AnalyticsFilters) *gorm.DB {
	query := s.DB.WithContext(ctx).
		Table("clicks").
		Select(`
			ad_id,
			COUNT(*) as click_count,
			COUNT(DISTINCT user_ip) as unique_clicks,
			AVG(playback_time_sec) as avg_playback_time,
			AVG(watched_percent) as avg_watch_percent,
			MAX(created_at) as last_updated
		`).
		Group("ad_id").
		Order("click_count DESC"). // Most clicked ads first
		Limit(filters.Limit).
		Offset(filters.Offset)

	// Apply filters
	if filters.AdID != 0 {
		query = query.Where("ad_id = ?", filters.AdID)
	}

	if !filters.Since.IsZero() {
		query = query.Where("created_at >= ?", filters.Since)
	}

	if !filters.Until.IsZero() {
		query = query.Where("created_at <= ?", filters.Until)
	}

	return query
}

// addCTRToResults calculates CTR by fetching impression data
func (s *Service) addCTRToResults(ctx context.Context, results []AdAnalytics) {
	if len(results) == 0 {
		return
	}

	var adIDs []int
	for _, result := range results {
		adIDs = append(adIDs, result.AdID)
	}

	type ImpressionData struct {
		AdID        int   `json:"ad_id"`
		Impressions int64 `json:"impressions"`
	}

	var impressions []ImpressionData
	err := s.DB.WithContext(ctx).
		Table("impressions").
		Select("ad_id, COUNT(*) as impressions").
		Where("ad_id IN ?", adIDs).
		Group("ad_id").
		Scan(&impressions).Error

	if err != nil {
		observability.Logger.Warn("Failed to fetch impressions for CTR calculation", zap.Error(err))
		return
	}

	// Create lookup map for impressions
	impressionMap := make(map[int]int64)
	for _, imp := range impressions {
		impressionMap[imp.AdID] = imp.Impressions
	}

	// Calculate CTR for each result
	for i := range results {
		if impressions, exists := impressionMap[results[i].AdID]; exists && impressions > 0 {
			results[i].Impressions = impressions
			results[i].CTR = (float64(results[i].ClickCount) / float64(impressions)) * 100
		}
	}

	observability.Logger.Debug("CTR calculated for analytics results",
		zap.Int("ads_with_ctr", len(impressionMap)))
}

// updateCache updates the in-memory cache with fresh data
func (s *Service) updateCache(results []AdAnalytics) {
	if len(results) == 0 {
		return
	}

	s.cache.mu.Lock()
	defer s.cache.mu.Unlock()

	updated := 0
	for _, result := range results {
		s.cache.metrics[result.AdID] = &result
		updated++
	}

	observability.Logger.Debug("Analytics cache updated",
		zap.Int("ads_updated", updated))
}

// startCacheRefresh runs background cache refresh for real-time analytics
func (s *Service) startCacheRefresh() {
	ticker := time.NewTicker(1 * time.Minute) // Refresh every minute
	defer ticker.Stop()

	observability.Logger.Info("Analytics cache refresh started")

	for range ticker.C {
		s.refreshTopAds()
	}
}

// refreshTopAds refreshes cache with most active ads from the last hour
func (s *Service) refreshTopAds() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filters := AnalyticsFilters{
		TimeWindow: time.Hour,
		Limit:      100, // Cache top 100 ads
		Offset:     0,
		IncludeCTR: false, // Skip CTR for background refresh to save time
	}

	results, err := s.fetchFromDatabase(ctx, filters)
	if err != nil {
		observability.Logger.Error("Failed to refresh analytics cache", zap.Error(err))
		return
	}

	observability.Logger.Debug("Analytics cache refreshed",
		zap.Int("ads_cached", len(results)))
}

// GetCacheStats returns cache statistics for monitoring
func (s *Service) GetCacheStats() map[string]interface{} {
	s.cache.mu.RLock()
	defer s.cache.mu.RUnlock()

	validEntries := 0
	expiredEntries := 0
	now := time.Now()

	for _, cached := range s.cache.metrics {
		if now.Sub(cached.LastUpdated) < s.cache.ttl {
			validEntries++
		} else {
			expiredEntries++
		}
	}

	return map[string]interface{}{
		"total_entries":   len(s.cache.metrics),
		"valid_entries":   validEntries,
		"expired_entries": expiredEntries,
		"ttl_minutes":     s.cache.ttl.Minutes(),
	}
}

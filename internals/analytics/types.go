package analytics

import (
	"sync"
	"time"
)

// AnalyticsCache provides thread-safe in-memory caching for real-time metrics
type AnalyticsCache struct {
	mu      sync.RWMutex
	metrics map[int]*AdAnalytics
	ttl     time.Duration
}

type AdAnalytics struct {
	AdID            int       `json:"ad_id"`
	ClickCount      int64     `json:"click_count"`
	UniqueClicks    int64     `json:"unique_clicks"`
	AvgPlaybackTime float64   `json:"avg_playback_time"`
	AvgWatchPercent float64   `json:"avg_watch_percent"`
	CTR             float64   `json:"ctr,omitempty"`
	Impressions     int64     `json:"impressions,omitempty"`
	LastUpdated     time.Time `json:"last_updated"`
}

type AnalyticsFilters struct {
	AdID       int           `json:"ad_id,omitempty"`
	Since      time.Time     `json:"since,omitempty"`
	Until      time.Time     `json:"until,omitempty"`
	TimeWindow time.Duration `json:"time_window,omitempty"` // e.g., last 15 minutes
	Limit      int           `json:"limit,omitempty"`
	Offset     int           `json:"offset,omitempty"`
	RealTime   bool          `json:"real_time,omitempty"`   // Use cache for real-time data
	IncludeCTR bool          `json:"include_ctr,omitempty"` // Calculate CTR
}

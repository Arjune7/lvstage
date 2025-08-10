package models

import (
	"time"

	"github.com/google/uuid"
)

type AdAnalytics struct {
	AdID             int       `json:"ad_id"`
	Impressions      int64     `json:"impressions"`       // Number of times the ad was served
	UniqueViewers    int64     `json:"unique_viewers"`    // Distinct users who viewed
	Clicks           int64     `json:"clicks"`            // Total clicks
	UniqueClicks     int64     `json:"unique_clicks"`     // Distinct users who clicked
	CTR              float64   `json:"ctr"`               // Click-through rate (Clicks / Impressions * 100)
	Conversions      int64     `json:"conversions"`       // Number of successful actions (purchase, signup, etc.)
	ConversionRate   float64   `json:"conversion_rate"`   // Conversions / Clicks * 100
	Revenue          float64   `json:"revenue"`           // Revenue from ad conversions
	AvgPlaybackTime  float64   `json:"avg_playback_time"` // Average playback time in seconds
	AvgWatchPercent  float64   `json:"avg_watch_percent"` // Percentage of ad watched on average
	BounceRate       float64   `json:"bounce_rate"`       // % of users who leave quickly
	FraudulentClicks int64     `json:"fraudulent_clicks"` // Click fraud detection
	ErrorRate        float64   `json:"error_rate"`        // Failures in ad load/play
	FirstSeen        time.Time `json:"first_seen"`        // First impression timestamp
	LastUpdated      time.Time `json:"last_updated"`      // Last analytics update timestamp
}

type Impression struct {
	ID        uint      `gorm:"primaryKey"`
	EventID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	AdID      uint      `gorm:"not null;index"`
	UserIP    string    `gorm:"size:45;not null"`
	UserAgent string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime;index"`
}

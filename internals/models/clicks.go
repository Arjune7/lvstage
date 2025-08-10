package models

import (
	"time"

	"github.com/google/uuid"
)

type Click struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	EventID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();not null;uniqueIndex" json:"event_id"` // Prevent duplicates, auto-generate
	AdID            uint      `gorm:"not null;index" json:"ad_id"`
	UserIP          string    `gorm:"size:45;not null" json:"user_ip"` // IPv4 & IPv6
	UserAgent       string    `gorm:"type:text" json:"user_agent"`
	PlaybackTimeSec float64   `json:"playback_time_sec"` // Video playback duration
	WatchedPercent  float64   `json:"watched_percent"`   // % watched
	IsFraudulent    bool      `gorm:"default:false" json:"is_fraudulent"`
	CreatedAt       time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

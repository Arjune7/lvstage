package clicks

import "github.com/google/uuid"

type ClickRequestData struct {
	AdID             uint    `json:"ad_id" binding:"required"`
	PlaybackTimeSecs float64 `json:"playback_time_secs" binding:"gte=0"`
	WatchedPercent   float64 `json:"watched_percent" binding:"gte=0,lte=100"`
	Timestamp        int64   `json:"timestamp"`
	UserIP           string
	UserAgent        string
	EventID          uuid.UUID
}

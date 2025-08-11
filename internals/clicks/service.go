package clicks

import (
	"context"
	"errors"
	"lystage-proj/internals/observability"
	"lystage-proj/internals/queue"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	RecordClick(data ClickRequestData) error
}

type clickService struct{}

func NewService() Service {
	return &clickService{}
}

func (s *clickService) RecordClick(data ClickRequestData) error {
	if data.AdID == 0 {
		return errors.New("invalid ad_id")
	}
	if data.EventID == uuid.Nil {
		data.EventID = uuid.New()
	}

	event := queue.ClickEvent{
		EventID:   data.EventID,
		AdID:      data.AdID,
		UserIP:    data.UserIP,
		Agent:     data.UserAgent,
		Watched:   data.WatchedPercent,
		PlayTime:  data.PlaybackTimeSecs,
		Timestamp: data.Timestamp,
	}

	go func(ev queue.ClickEvent) {
		maxRetries := 5
		retryInterval := time.Second

		for i := range maxRetries {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := queue.PublishClick(ctx, ev)
			cancel()

			if err == nil {
				observability.Logger.Debug("Click published to Kafka",
					zap.Uint("ad_id", ev.AdID),
					zap.String("event_id", ev.EventID.String()),
				)
				return // success, exit retry loop
			}

			observability.Logger.Warn("Failed to publish click to Kafka, retrying",
				zap.Error(err),
				zap.Uint("ad_id", ev.AdID),
				zap.String("event_id", ev.EventID.String()),
				zap.Int("attempt", i+1),
			)

			// Exponential backoff
			time.Sleep(retryInterval)
			retryInterval *= 2
		}

		observability.Logger.Error("Exhausted retries, failed to publish click to Kafka",
			zap.Uint("ad_id", ev.AdID),
			zap.String("event_id", ev.EventID.String()),
		)
	}(event)

	return nil
}

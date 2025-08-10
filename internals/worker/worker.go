package worker

import (
	"context"
	"encoding/json"
	"errors"
	"lystage-proj/internals/db"
	"lystage-proj/internals/models"
	"lystage-proj/internals/observability"
	"lystage-proj/internals/queue"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func StartClickConsumer(broker, topic, group string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		GroupID:     group,
		Topic:       topic,
		StartOffset: kafka.LastOffset,
		MaxBytes:    10e6, // handle large batches
	})

	go func() {
		for {
			msg, err := r.ReadMessage(context.Background())
			if err != nil {
				observability.Logger.Error("Kafka consumer read error", zap.Error(err))
				continue
			}

			var e queue.ClickEvent
			if err := json.Unmarshal(msg.Value, &e); err != nil {
				observability.Logger.Warn("Invalid click event format", zap.Error(err))
				continue
			}

			if err := saveClickEvent(e); err != nil {
				observability.Logger.Error("Failed to save click event", zap.Error(err))
				// Optionally: retry or send to DLQ
			}
		}
	}()
}

func saveClickEvent(e queue.ClickEvent) error {
	if e.AdID == 0 || e.EventID == uuid.Nil {
		return errors.New("invalid click event data")
	}

	click := models.Click{
		EventID:         e.EventID,
		AdID:            e.AdID,
		UserIP:          e.UserIP,
		UserAgent:       e.Agent,
		PlaybackTimeSec: e.PlayTime,
		WatchedPercent:  e.Watched,
		IsFraudulent:    false, // Hook for fraud detection
		CreatedAt:       time.Now(),
	}

	if err := db.GormDB.Create(&click).Error; err != nil {
		return err
	}

	observability.Logger.Info("Click stored in DB",
		zap.Uint("ad_id", e.AdID),
		zap.String("event_id", e.EventID.String()),
	)
	return nil
}

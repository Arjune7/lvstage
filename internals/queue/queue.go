package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sony/gobreaker"
)

// ClickEvent is the payload for click tracking.
type ClickEvent struct {
	EventID   uuid.UUID `json:"event_id"`
	AdID      uint      `json:"ad_id"`
	UserIP    string    `json:"user_ip"`
	Agent     string    `json:"agent"`
	PlayTime  float64   `json:"play_time_secs"`
	Watched   float64   `json:"watched_percent"`
	Timestamp int64     `json:"timestamp"`
}

// Producer wraps Kafka writer for publishing events.
type Producer struct {
	writer *kafka.Writer
	topic  string
}

// Global producer instance with thread safety
var (
	globalProducer      *Producer
	producerMutex       sync.RWMutex
	isInitialized       bool
	kafkaCircuitBreaker *gobreaker.CircuitBreaker
)

func init() {
	settings := gobreaker.Settings{
		Name:        "KafkaPublishCircuitBreaker",
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5 ||
				(counts.Requests >= 10 && counts.TotalFailures*2 >= counts.Requests)
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("Circuit breaker state changed: %s from %v to %v\n", name, from, to)
		},
	}
	kafkaCircuitBreaker = gobreaker.NewCircuitBreaker(settings)
}

// InitGlobalProducer initializes the global Kafka producer - call this once at app startup
func InitGlobalProducer(broker, topic string) error {
	producerMutex.Lock()
	defer producerMutex.Unlock()

	if isInitialized {
		return errors.New("global producer already initialized")
	}

	if broker == "" || topic == "" {
		return errors.New("broker and topic must be provided")
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{broker},
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 100 * time.Millisecond,
		Async:        true,
		RequiredAcks: int(kafka.RequireOne),
		BatchSize:    100,
	})

	globalProducer = &Producer{writer: writer, topic: topic}
	isInitialized = true

	log.Printf("✅ Global Kafka producer initialized (broker=%s, topic=%s)\n", broker, topic)
	return nil
}

// PublishClick sends a ClickEvent to Kafka using the global producer with circuit breaker protection
func PublishClick(ctx context.Context, event ClickEvent) error {
	producerMutex.RLock()
	producer := globalProducer
	initialized := isInitialized
	producerMutex.RUnlock()

	if !initialized || producer == nil {
		return errors.New("kafka producer is not initialized - call InitGlobalProducer first")
	}

	_, err := kafkaCircuitBreaker.Execute(func() (interface{}, error) {
		return nil, producer.publishClick(ctx, event)
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return fmt.Errorf("circuit breaker is open, skipping kafka publish")
		}
		if errors.Is(err, gobreaker.ErrTooManyRequests) {
			return fmt.Errorf("too many requests, circuit breaker limiting calls")
		}
		return err
	}

	return nil
}

// publishClick is the internal method that does the actual publishing with retry logic
func (p *Producer) publishClick(ctx context.Context, event ClickEvent) error {
	if p.writer == nil {
		return errors.New("kafka writer is not initialized")
	}

	msg, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal click event: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var lastErr error
	for i := range 3 {
		err = p.writer.WriteMessages(timeoutCtx, kafka.Message{
			Key:   []byte(event.EventID.String()),
			Value: msg,
			Time:  time.Now(),
		})
		if err == nil {
			log.Printf("📤 Click event published: AdID=%d, EventID=%s\n", event.AdID, event.EventID.String())
			return nil
		}
		lastErr = err
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond) // exponential backoff
	}

	return fmt.Errorf("failed to publish click event after retries: %w", lastErr)
}

// CloseGlobalProducer closes the global producer - call this at app shutdown
func CloseGlobalProducer() error {
	producerMutex.Lock()
	defer producerMutex.Unlock()

	if globalProducer != nil && globalProducer.writer != nil {
		err := globalProducer.writer.Close()
		globalProducer = nil
		isInitialized = false
		log.Println("✅ Global Kafka producer closed")
		return err
	}
	return nil
}

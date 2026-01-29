package event

import "time"

type Event struct {
	Type      string         `json:"type"`
	Key       string         `json:"key"` // Kafka key（比如 user id）
	OccurredAt time.Time     `json:"occurred_at"`
	Payload   map[string]any `json:"payload"`
}

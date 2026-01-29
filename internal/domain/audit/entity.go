package audit

import "time"

type Log struct {
	ID        uint64
	EventType string
	EventKey  string
	Payload   []byte
	CreatedAt time.Time
}

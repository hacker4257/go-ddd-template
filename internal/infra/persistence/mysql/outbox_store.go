package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/hacker4257/go-ddd-template/internal/domain/event"
)

type OutboxRow struct {
	ID      uint64
	Topic   string
	MsgKey  string
	Type    string
	Payload []byte
	Headers []byte
}

type OutboxStore struct {
	db *sql.DB
}

func NewOutboxStore(db *sql.DB) *OutboxStore {
	return &OutboxStore{db: db}
}

// 给 app 用：写入 outbox（要求在事务里）
func (s *OutboxStore) Add(ctx context.Context, m event.OutboxMessage) error {
	ex := getExecer(s.db, ctx)

	payload, err := json.Marshal(m.Payload)
	if err != nil {
		return err
	}
	headers, err := json.Marshal(m.Headers)
	if err != nil {
		return err
	}

	const q = `INSERT INTO outbox (topic, msg_key, event_type, payload, headers) VALUES (?, ?, ?, ?, ?)`
	_, err = ex.ExecContext(ctx, q, m.Topic, m.Key, m.Type, payload, headers)
	return err
}

// 给 worker 用：拉取未发送
func (s *OutboxStore) ListUnsent(ctx context.Context, limit int) ([]OutboxRow, error) {
	const q = `
SELECT id, topic, msg_key, event_type, payload, COALESCE(headers, 'null')
FROM outbox
WHERE sent_at IS NULL
ORDER BY id
LIMIT ?`

	rows, err := s.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []OutboxRow
	for rows.Next() {
		var r OutboxRow
		if err := rows.Scan(&r.ID, &r.Topic, &r.MsgKey, &r.Type, &r.Payload, &r.Headers); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, rows.Err()
}

func (s *OutboxStore) MarkSent(ctx context.Context, id uint64) error {
	const q = `UPDATE outbox SET sent_at = ? WHERE id = ? AND sent_at IS NULL`
	_, err := s.db.ExecContext(ctx, q, time.Now(), id)
	return err
}

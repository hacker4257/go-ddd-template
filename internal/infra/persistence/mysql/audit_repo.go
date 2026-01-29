package mysql

import (
	"context"
	"database/sql"
)

type AuditRepo struct {
	db *sql.DB
}

func NewAuditRepo(db *sql.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) Insert(ctx context.Context, eventType, eventKey string, payload []byte) error {
	ex := getExecer(r.db, ctx)
	const q = `INSERT INTO audit_logs (event_type, event_key, payload) VALUES (?, ?, ?)`
	_, err := ex.ExecContext(ctx, q, eventType, eventKey, payload)
	return err
}

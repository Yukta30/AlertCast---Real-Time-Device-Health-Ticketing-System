package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"alertcast/internal/models"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}
	p := &Postgres{pool: pool}
	if err := p.migrate(ctx); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Postgres) migrate(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS tickets (
  id BIGSERIAL PRIMARY KEY,
  device_id TEXT NOT NULL,
  severity TEXT NOT NULL CHECK (severity IN ('critical','high','medium','low')),
  reason TEXT NOT NULL,
  event_json JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_tickets_created_at ON tickets(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tickets_severity ON tickets(severity);
`
	_, err := p.pool.Exec(ctx, ddl)
	return err
}

func (p *Postgres) Close() { p.pool.Close() }

func (p *Postgres) InsertTicket(ctx context.Context, t *models.Ticket) error {
	const q = `
INSERT INTO tickets (device_id, severity, reason, event_json, created_at)
VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at;
`
	row := p.pool.QueryRow(ctx, q, t.DeviceID, t.Severity, t.Reason, t.EventJSON, time.Now().UTC())
	return row.Scan(&t.ID, &t.CreatedAt)
}

func (p *Postgres) RecentTickets(ctx context.Context, limit int) ([]models.Ticket, error) {
	const q = `
SELECT id, device_id, severity, reason, event_json, created_at
FROM tickets
ORDER BY created_at DESC
LIMIT $1;
`
	rows, err := p.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Ticket
	for rows.Next() {
		var t models.Ticket
		if err := rows.Scan(&t.ID, &t.DeviceID, &t.Severity, &t.Reason, &t.EventJSON, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (p *Postgres) SeverityCounts(ctx context.Context) (map[string]int64, error) {
	const q = `
SELECT severity, COUNT(*)
FROM tickets
GROUP BY severity;
`
	rows, err := p.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{"critical":0,"high":0,"medium":0,"low":0}
	for rows.Next() {
		var s string
		var c int64
		if err := rows.Scan(&s, &c); err != nil {
			return nil, err
		}
		out[s] = c
	}
	return out, rows.Err()
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *Postgres) ExecTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

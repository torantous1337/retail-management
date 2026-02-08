package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/torantous1337/retail-management/internal/core/domain"
)

// AuditLogRepository implements the audit log repository using SQLite.
type AuditLogRepository struct {
	db *sqlx.DB
}

// NewAuditLogRepository creates a new audit log repository instance.
func NewAuditLogRepository(db *sqlx.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// auditLogRow is a database row representation for audit logs.
type auditLogRow struct {
	ID          int64          `db:"id"`
	Action      string         `db:"action"`
	UserID      string         `db:"user_id"`
	Timestamp   time.Time      `db:"timestamp"`
	Payload     string         `db:"payload"`
	PrevHash    sql.NullString `db:"prev_hash"`
	CurrentHash string         `db:"current_hash"`
}

// Create creates a new audit log entry.
func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	// Serialize payload to JSON
	payloadJSON, err := json.Marshal(log.Payload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO audit_logs (action, user_id, timestamp, payload, prev_hash, current_hash)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	prevHash := sql.NullString{
		String: log.PrevHash,
		Valid:  log.PrevHash != "",
	}

	result, err := r.db.ExecContext(ctx, query,
		log.Action,
		log.UserID,
		log.Timestamp,
		string(payloadJSON),
		prevHash,
		log.CurrentHash,
	)
	if err != nil {
		return err
	}

	// Get the inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	log.ID = id

	return nil
}

// GetLastLog retrieves the most recent audit log entry.
func (r *AuditLogRepository) GetLastLog(ctx context.Context) (*domain.AuditLog, error) {
	query := `SELECT * FROM audit_logs ORDER BY id DESC LIMIT 1`

	var row auditLogRow
	err := r.db.GetContext(ctx, &row, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No logs yet
		}
		return nil, err
	}

	return r.toDomain(&row)
}

// List retrieves audit logs with pagination.
func (r *AuditLogRepository) List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error) {
	query := `SELECT * FROM audit_logs ORDER BY id DESC LIMIT ? OFFSET ?`

	var rows []auditLogRow
	err := r.db.SelectContext(ctx, &rows, query, limit, offset)
	if err != nil {
		return nil, err
	}

	logs := make([]*domain.AuditLog, 0, len(rows))
	for _, row := range rows {
		log, err := r.toDomain(&row)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// VerifyChain verifies the integrity of the audit log chain.
func (r *AuditLogRepository) VerifyChain(ctx context.Context) (bool, error) {
	query := `SELECT * FROM audit_logs ORDER BY id ASC`

	var rows []auditLogRow
	err := r.db.SelectContext(ctx, &rows, query)
	if err != nil {
		return false, err
	}

	if len(rows) == 0 {
		return true, nil // Empty chain is valid
	}

	var prevHash string
	for _, row := range rows {
		// Verify that prev_hash matches the previous record's current_hash
		if row.PrevHash.Valid && row.PrevHash.String != prevHash {
			return false, nil
		}

		// Verify that current_hash is correctly calculated
		expectedHash := r.calculateHash(row.Payload, row.Timestamp, prevHash)
		if row.CurrentHash != expectedHash {
			return false, nil
		}

		prevHash = row.CurrentHash
	}

	return true, nil
}

// toDomain converts a database row to a domain entity.
func (r *AuditLogRepository) toDomain(row *auditLogRow) (*domain.AuditLog, error) {
	log := &domain.AuditLog{
		ID:          row.ID,
		Action:      row.Action,
		UserID:      row.UserID,
		Timestamp:   row.Timestamp,
		CurrentHash: row.CurrentHash,
	}

	if row.PrevHash.Valid {
		log.PrevHash = row.PrevHash.String
	}

	// Deserialize payload from JSON
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(row.Payload), &payload)
	if err != nil {
		return nil, err
	}
	log.Payload = payload

	return log, nil
}

// calculateHash computes SHA256(payload + timestamp + prev_hash) for verification.
func (r *AuditLogRepository) calculateHash(payloadJSON string, timestamp time.Time, prevHash string) string {
	hashInput := fmt.Sprintf("%s%s%s", payloadJSON, timestamp.Format(time.RFC3339Nano), prevHash)
	hash := sha256.Sum256([]byte(hashInput))
	return fmt.Sprintf("%x", hash)
}

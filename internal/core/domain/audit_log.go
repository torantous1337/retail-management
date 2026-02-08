package domain

import "time"

// AuditLog represents an immutable audit trail entry.
// This is a pure business entity with no framework tags.
type AuditLog struct {
	ID          int64
	Action      string
	UserID      string
	Timestamp   time.Time
	Payload     map[string]interface{} // JSON payload of the change
	PrevHash    string                 // Hash of previous record (empty for first record)
	CurrentHash string                 // SHA256(payload + timestamp + prev_hash)
}
